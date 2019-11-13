package node

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/scribe"
)

//go:generate faux --interface Transport --output fakes/transport.go
type Transport interface {
	Drop(root, uri string) (io.ReadCloser, error)
}

type EngineDependencyManager struct {
	transport Transport
	logger    scribe.Logger
}

func NewEngineDependencyManager(transport Transport, logger scribe.Logger) EngineDependencyManager {
	return EngineDependencyManager{
		transport: transport,
		logger:    logger,
	}
}

func (e EngineDependencyManager) Resolve(dependencies []BuildpackMetadataDependency, defaultVersion, stack string, entry packit.BuildpackPlanEntry) (BuildpackMetadataDependency, error) {
	if entry.Version == "" {
		entry.Version = "default"
	}

	if entry.Version == "default" {
		entry.Version = "*"
		if defaultVersion != "" {
			entry.Version = defaultVersion
		}
	}

	var compatibleVersions []BuildpackMetadataDependency
	versionConstraint, err := semver.NewConstraint(entry.Version)
	if err != nil {
		return BuildpackMetadataDependency{}, err
	}

	for _, dependency := range dependencies {
		if dependency.ID != entry.Name || !dependency.Stacks.Include(stack) {
			continue
		}

		sVersion, err := semver.NewVersion(dependency.Version)
		if err != nil {
			return BuildpackMetadataDependency{}, err
		}

		if versionConstraint.Check(sVersion) {
			compatibleVersions = append(compatibleVersions, dependency)
		}
	}

	if len(compatibleVersions) == 0 {
		return BuildpackMetadataDependency{}, fmt.Errorf("failed to satisfy %q dependency version constraint %q: no compatible versions", entry.Name, entry.Version)
	}

	sort.Slice(compatibleVersions, func(i, j int) bool {
		iVersion := semver.MustParse(compatibleVersions[i].Version)
		jVersion := semver.MustParse(compatibleVersions[j].Version)
		return iVersion.GreaterThan(jVersion)
	})

	return compatibleVersions[0], nil
}

func (e EngineDependencyManager) Install(dependency BuildpackMetadataDependency, cnbPath, layerPath string) error {
	e.logger.Subprocess("Installing Node Engine %s", dependency.Version)
	then := time.Now()

	bundle, err := e.transport.Drop(cnbPath, dependency.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch dependency: %s", err)
	}
	defer bundle.Close()

	validatedReader := cargo.NewValidatedReader(bundle, dependency.SHA256)
	zr, err := gzip.NewReader(validatedReader)
	if err != nil {
		return fmt.Errorf("failed to read gzip response: %s", err)
	}

	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar response: %s", err)
		}

		name := strings.Join(strings.Split(hdr.Name, string(filepath.Separator))[1:], string(filepath.Separator))
		if name == "" {
			continue
		}

		path := filepath.Join(layerPath, name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create archived directory: %s", err)
			}

		case tar.TypeReg:
			err := writeStreamingFile(tr, path, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to stream file from archive: %s", err)
			}

		case tar.TypeSymlink:
			err = os.Symlink(hdr.Linkname, path)
			if err != nil {
				return fmt.Errorf("failed to extract symlink: %s", err)
			}
		}
	}

	ok, err := validatedReader.Valid()
	if err != nil {
		return fmt.Errorf("failed to validate dependency: %s", err)
	}

	if !ok {
		return fmt.Errorf("checksum does not match: %s", err)
	}

	e.logger.Action("Completed in %s", time.Since(then).Round(time.Millisecond))
	e.logger.Break()

	return nil
}

func writeStreamingFile(tr io.Reader, path string, fileMode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("failed to create file %s", err)
	}

	_, err = io.Copy(file, tr)
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
