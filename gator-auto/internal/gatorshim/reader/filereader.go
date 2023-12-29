package reader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/types"
	gatorreader "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var allowedExtensions = []string{".yaml", ".yml", ".json"}

type source struct {
	filename string
	image    string
	stdin    bool
	objs     []*unstructured.Unstructured
}

func readFile(filename string) ([]*source, error) {
	if err := verifyFile(filename); err != nil {
		return nil, err
	}

	var sources []*source
	expanded, err := expandDirectories([]string{filename})
	if err != nil {
		return nil, fmt.Errorf("normalizing filenames: %w", err)
	}

	for _, f := range expanded {
		file, err := os.Open(f)
		if err != nil {
			return nil, fmt.Errorf("opening file %q: %w", f, err)
		}
		defer file.Close()

		us, err := gatorreader.ReadK8sResources(bufio.NewReader(file))
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", f, err)
		}

		sources = append(sources, &source{
			filename: f,
			objs:     us,
		})
	}

	return sources, nil
}

func readFiles(filenames []string) ([]*source, error) {
	var sources []*source
	for _, f := range filenames {
		s, err := readFile(f)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s...)
	}

	return sources, nil
}

// verifyFile checks that the filenames aren't themselves disallowed extensions.
// This yields a much better user experience when the user mis-uses the
// --filename flag.
func verifyFile(filename string) error {
	// make sure it's a file, not a directory
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("stat on path %q: %w", filename, err)
	}

	if fileInfo.IsDir() {
		return nil
	}
	if !allowedExtension(filename) {
		return fmt.Errorf("path %q must be of extensions: %v", filename, allowedExtensions)
	}

	return nil
}

func expandDirectories(filenames []string) ([]string, error) {
	var output []string

	for _, filename := range filenames {
		paths, err := filesBelow(filename)
		if err != nil {
			return nil, fmt.Errorf("filename %q: %w", filename, err)
		}
		output = append(output, paths...)
	}

	return output, nil
}

// filesBelow walks the filetree from startPath and below, collecting a list of
// all the filepaths.  Directories are excluded.
func filesBelow(startPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// only add files to the normalized output
		if info.IsDir() {
			return nil
		}

		// make sure the file extension is valid
		if !allowedExtension(path) {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking: %w", err)
	}

	return files, nil
}

func allowedExtension(path string) bool {
	for _, ext := range allowedExtensions {
		if ext == filepath.Ext(path) {
			return true
		}
	}

	return false
}

func sourcesToUnstruct(sources []*source) []*unstructured.Unstructured {
	var us []*unstructured.Unstructured
	for _, s := range sources {
		us = append(us, s.objs...)
	}
	return us
}

func ReadTargets(filenames []string) (*types.TestTargets, error) {
	sources, err := readFiles(filenames)
	if err != nil {
		return nil, err
	}

	rv := &types.TestTargets{
		ObjectSources: make(map[*unstructured.Unstructured]string),
	}

	for _, source := range sources {
		rv.Objects = append(rv.Objects, source.objs...)
		for _, obj := range source.objs {
			rv.ObjectSources[obj] = source.filename
		}
	}

	return rv, nil
}
