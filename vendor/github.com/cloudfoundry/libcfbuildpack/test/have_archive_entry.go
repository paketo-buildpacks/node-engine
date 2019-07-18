package test

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/onsi/gomega/types"
	"io"
	"os"
)

// HaveContent tests that a file has expected content.
func HaveArchiveEntry(expected string) types.GomegaMatcher {
	return &haveArchiveEntryMatcher{
		expected: expected,
	}
}

type haveArchiveEntryMatcher struct {
	expected string
	actualEntries []string
}

func (m *haveArchiveEntryMatcher) Match(actual interface{}) (bool, error) {
	path, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("HaveArchiveEntry matcher expects a path")
	}

	fh, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open tar file: %s", err)
	}
	defer fh.Close()

	gzr, err := gzip.NewReader(fh)
	if err != nil {
		return false, fmt.Errorf("failed to crete gzip reader: %s", err)
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to read next archive entry: %s", err)
		}
		if header.Name == m.expected {
			return true, nil
		}
		m.actualEntries = append(m.actualEntries, header.Name)
	}
}

func (m *haveArchiveEntryMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto contain archive entry\n\t%#v\ngot entries\n\t%#v", actual, m.expected, m.actualEntries)
}

func (m *haveArchiveEntryMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to contain archive entry\n\t%#v", actual, m.expected)
}
