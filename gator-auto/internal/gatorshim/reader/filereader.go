package reader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	gatorreader "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ref: https://github.com/open-policy-agent/gatekeeper/blob/5978ea8a08b7494dbf82a54fc5c109f9e00f33ae/pkg/gator/reader/filereader.go

var allowedExtensions = []string{".yaml", ".yml", ".json"}

func readFile(filename string) (*TestTargets, error) {
	if err := verifyFile(filename); err != nil {
		return nil, err
	}

	rv := &TestTargets{
		ObjectSources: make(map[*unstructured.Unstructured]ObjectSource),
	}

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

		rv.Objects = append(rv.Objects, us...)
		for _, obj := range us {
			rv.ObjectSources[obj] = ObjectSource{
				SourceType: SourceTypeFile,
				FilePath:   f,
			}
		}
	}

	return rv, nil
}

func readFiles(filenames []string) (*TestTargets, error) {
	var rv *TestTargets
	for _, f := range filenames {
		s, err := readFile(f)
		if err != nil {
			return nil, err
		}
		rv = rv.merge(s)
	}

	return rv, nil
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
