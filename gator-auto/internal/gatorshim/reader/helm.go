package reader

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	gatorreader "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

	us, err := gatorreader.ReadK8sResources(stdout)
	if err != nil {
		return nil, err
	}

	rv := &TestTargets{
		ObjectSources: make(map[*unstructured.Unstructured]ObjectSource),
	}

	rv.Objects = append(rv.Objects, us...)
	for _, obj := range us {
		rv.ObjectSources[obj] = ObjectSource{
			SourceType: SourceTypeHelm,
			// TODO: read from `# Source: ...` annotation?
			FilePath: chart,
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
