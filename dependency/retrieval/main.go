package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type NodeRelease struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}

type ReleaseSchedule map[string]struct {
	End string `json:"end"`
}

type NodeMetadata struct {
	SemverVersion *semver.Version
}

func (nodeMetadata NodeMetadata) Version() *semver.Version {
	return nodeMetadata.SemverVersion
}

func main() {
	retrieve.NewMetadataWithPlatforms("node", getAllVersions, generateMetadataWithPlatform)
}

func generateMetadataWithPlatform(versionFetcher versionology.VersionFetcher, platform retrieve.Platform) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()

	body, err := httpGet("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("could not get release index: %w", err)
	}

	var nodeReleases []NodeRelease
	err = json.Unmarshal(body, &nodeReleases)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w\n%s", err, body)
	}

	releaseSchedule, err := getReleaseSchedule()
	if err != nil {
		return nil, err
	}

	for _, release := range nodeReleases {
		if strings.TrimPrefix(release.Version, "v") == version {
			return createDependencyMetadata(release, releaseSchedule, platform)
		}
	}

	return nil, fmt.Errorf("could not find version %s", version)
}

func getAllVersions() (versionology.VersionFetcherArray, error) {
	body, err := httpGet("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("could not get release index: %w", err)
	}

	var nodeReleases []NodeRelease
	err = json.Unmarshal(body, &nodeReleases)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w\n%s", err, body)
	}

	sort.SliceStable(nodeReleases, func(i, j int) bool {
		return nodeReleases[i].Date > nodeReleases[j].Date
	})

	var versions []versionology.VersionFetcher
	for _, release := range nodeReleases {
		versions = append(versions, NodeMetadata{
			semver.MustParse(release.Version),
		})
	}

	return versions, nil
}

func getReleaseSchedule() (ReleaseSchedule, error) {
	body, err := httpGet("https://raw.githubusercontent.com/nodejs/Release/master/schedule.json")
	if err != nil {
		return ReleaseSchedule{}, fmt.Errorf("could not get release schedule: %w", err)
	}

	var releaseSchedule map[string]struct {
		End string `json:"end"`
	}
	err = json.Unmarshal(body, &releaseSchedule)
	if err != nil {
		return ReleaseSchedule{}, fmt.Errorf("could not unmarshal release schedule: %w\n%s", err, body)
	}

	return releaseSchedule, nil
}

func createDependencyMetadata(release NodeRelease, releaseSchedule ReleaseSchedule, platform retrieve.Platform) ([]versionology.Dependency, error) {

	var nodeArch string
	if platform.Arch == "amd64" {
		nodeArch = "x64"
	} else {
		nodeArch = platform.Arch
	}

	version := release.Version
	url := fmt.Sprintf("https://nodejs.org/dist/%[1]s/node-%[1]s-%[2]s-%[3]s.tar.xz", version, platform.OS, nodeArch)

	checksum, err := getChecksum(version, platform)
	if err != nil {
		return nil, err
	}

	deprecationDate := getDeprecationDate(version, releaseSchedule)

	dep := cargo.ConfigMetadataDependency{
		Version:         strings.TrimPrefix(version, "v"),
		ID:              "node",
		Name:            "Node Engine",
		Source:          url,
		SourceChecksum:  fmt.Sprintf("sha256:%s", checksum),
		CPE:             fmt.Sprintf("cpe:2.3:a:nodejs:node.js:%s:*:*:*:*:*:*:*", strings.TrimPrefix(version, "v")),
		PURL:            retrieve.GeneratePURL("node", version, checksum, url),
		URI:             url,
		Checksum:        fmt.Sprintf("sha256:%s", checksum),
		Licenses:        retrieve.LookupLicenses(url, upstream.DefaultDecompress),
		DeprecationDate: deprecationDate,
		StripComponents: 1,
		Stacks:          []string{"*"},
		OS:              platform.OS,
		Arch:            platform.Arch,
	}

	allStacksDependency, err := versionology.NewDependency(dep, "*")
	if err != nil {
		return nil, fmt.Errorf("could not get create * dependency: %w", err)
	}

	return []versionology.Dependency{allStacksDependency}, nil
}

func getDeprecationDate(version string, releaseSchedule ReleaseSchedule) *time.Time {
	versionIndex := strings.Split(version, ".")[0]
	if versionIndex == "v0" {
		versionIndex = strings.Join(strings.Split(version, ".")[0:2], ".")
	}
	release, ok := releaseSchedule[versionIndex]
	if !ok {
		return nil
	}

	deprecationDate, err := time.Parse("2006-01-02", release.End)
	if err != nil {
		return nil
	}

	return &deprecationDate
}

func getChecksum(version string, platform retrieve.Platform) (string, error) {

	var nodeArch string
	if platform.Arch == "amd64" {
		nodeArch = "x64"
	} else {
		nodeArch = platform.Arch
	}

	body, err := httpGet(fmt.Sprintf("https://nodejs.org/dist/%s/SHASUMS256.txt", version))
	if err != nil {
		return "", fmt.Errorf("could not get SHA256 file: %w", err)
	}

	var dependencySHA string
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasSuffix(line, fmt.Sprintf("node-%[1]s-%[2]s-%[3]s.tar.xz", version, platform.OS, nodeArch)) {
			dependencySHA = strings.Fields(line)[0]
		}
	}
	if dependencySHA == "" {
		return "", fmt.Errorf("could not find SHA256 for node-%s-%s-%s.tar.xz", version, platform.OS, nodeArch)
	}
	return dependencySHA, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not make get request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	return body, nil
}
