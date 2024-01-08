package reader

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"time"

	gatorreader "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// helmSourceDelimitedReaders delimits the helm rendered chart output using the `# Source: ...` annotation.
func helmSourceDelimitedReaders(source io.Reader) (map[string]io.Reader, error) {
	const unknownFile = ""
	const sourceAnnotationPrefix = "# Source: "

	fileBuffers := make(map[string]*bytes.Buffer)
	var (
		currentFile       string
		currentFileBuffer *bytes.Buffer
	)

	saveCurrentFileBuffer := func() error {
		if currentFileBuffer == nil {
			return nil
		}

		previousBuffer, ok := fileBuffers[currentFile]
		if ok {
			if _, err := previousBuffer.Write(currentFileBuffer.Bytes()); err != nil {
				return err
			}
		} else {
			fileBuffers[currentFile] = currentFileBuffer
		}

		currentFileBuffer = nil
		currentFile = unknownFile

		return nil
	}

	s := bufio.NewScanner(source)
	for s.Scan() {
		line := s.Text()

		// start a new file buffer
		if strings.HasPrefix(line, sourceAnnotationPrefix) {
			if err := saveCurrentFileBuffer(); err != nil {
				return nil, err
			}

			currentFile = strings.TrimPrefix(line, sourceAnnotationPrefix)
		}

		if currentFileBuffer == nil {
			currentFileBuffer = &bytes.Buffer{}
		}
		if _, err := currentFileBuffer.WriteString(line + "\n"); err != nil {
			return nil, err
		}
	}
	// save the last file buffer
	if err := saveCurrentFileBuffer(); err != nil {
		return nil, err
	}

	rv := make(map[string]io.Reader)
	for file, buffer := range fileBuffers {
		rv[file] = buffer
	}

	return rv, nil
}

func resolveHelmCommand(helmCommand string) (string, error) {
	if helmCommand == "" {
		helmCommand = "helm"
	}

	return exec.LookPath(helmCommand)
}

func readHelmChart(
	helmCommand string,
	chart string,
) (*TestTargets, error) {
	helmCommand, err := resolveHelmCommand(helmCommand)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, helmCommand, "template", chart)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		// TODO: log stderr
		return nil, err
	}

	readers, err := helmSourceDelimitedReaders(stdout)
	if err != nil {
		return nil, err
	}

	rv := &TestTargets{
		ObjectSources: make(map[*unstructured.Unstructured]ObjectSource),
	}

	for filePath, reader := range readers {
		us, err := gatorreader.ReadK8sResources(reader)
		if err != nil {
			return nil, err
		}

		rv.Objects = append(rv.Objects, us...)
		for _, obj := range us {
			rv.ObjectSources[obj] = ObjectSource{
				SourceType: SourceTypeHelm,
				FilePath:   filePath,
				// TODO: save chart reference
			}
		}
	}

	return rv, nil
}

func readHelmCharts(
	helmCommand string,
	charts []string,
) (*TestTargets, error) {
	var rv *TestTargets
	for _, chart := range charts {
		targets, err := readHelmChart(helmCommand, chart)
		if err != nil {
			return nil, err
		}
		rv = rv.merge(targets)
	}

	return rv, nil
}
