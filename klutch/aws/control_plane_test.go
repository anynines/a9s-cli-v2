package aws

import (
	"context"
	"slices"
	"testing"
)

func TestCreateControlPlaneCluster_DryRunSkipsExternalCalls(t *testing.T) {
	origRunCmd := runCmd
	origRunCmdWithPrompt := runCmdWithPrompt
	origLookPath := execLookPath
	runCmdCalled := false
	lookPathCalled := false
	runCmd = func(ctx context.Context, name string, args ...string) (string, error) {
		runCmdCalled = true
		t.Fatalf("runCmd should not be called during dry-run, got %s %v", name, args)
		return "", nil
	}
	runCmdWithPrompt = func(ctx context.Context, name string, args ...string) (string, error) {
		runCmdCalled = true
		t.Fatalf("runCmdWithPrompt should not be called during dry-run, got %s %v", name, args)
		return "", nil
	}
	execLookPath = func(file string) (string, error) {
		lookPathCalled = true
		t.Fatalf("execLookPath should not be called during dry-run, got %s", file)
		return "", nil
	}
	defer func() {
		runCmd = origRunCmd
		runCmdWithPrompt = origRunCmdWithPrompt
		execLookPath = origLookPath
	}()

	CreateControlPlaneCluster(context.Background(), CreateOptions{DryRun: true})

	if runCmdCalled || lookPathCalled {
		t.Fatalf("expected no external commands during dry-run")
	}
}

func TestTagEC2ResourceAddsClusterTags(t *testing.T) {
	origRunCmdWithPrompt := runCmdWithPrompt
	defer func() {
		runCmdWithPrompt = origRunCmdWithPrompt
		setClusterTagContext("", "")
	}()

	setClusterTagContext("demo", "arn:aws:eks:eu-central-1:111111111111:cluster/demo")

	var gotName string
	var gotArgs []string
	runCmdWithPrompt = func(ctx context.Context, name string, args ...string) (string, error) {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return "", nil
	}

	tagEC2Resource(context.Background(), "vpc-123", "my-resource")

	if gotName != "aws" {
		t.Fatalf("expected aws CLI, got %s", gotName)
	}

	expectedPrefix := []string{"ec2", "create-tags", "--resources", "vpc-123", "--tags"}
	if !slices.Equal(gotArgs[:len(expectedPrefix)], expectedPrefix) {
		t.Fatalf("unexpected command prefix: %v", gotArgs[:len(expectedPrefix)])
	}

	expectedTags := []string{
		"Key=Klutch,Value=ControlPlane",
		"Key=Name,Value=my-resource",
		"Key=eks.cluster/name,Value=demo",
		"Key=eks.cluster/id,Value=arn:aws:eks:eu-central-1:111111111111:cluster/demo",
	}

	if len(gotArgs) != len(expectedPrefix)+len(expectedTags) {
		t.Fatalf("unexpected arg length: got %d", len(gotArgs))
	}

	if !slices.Equal(gotArgs[len(expectedPrefix):], expectedTags) {
		t.Fatalf("unexpected tags: %v", gotArgs[len(expectedPrefix):])
	}
}
