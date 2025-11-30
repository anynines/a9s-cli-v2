package cmd

import (
	"context"
	"strings"
	"testing"

	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
)

func TestRunKlutchClusterCreationCallsAWS(t *testing.T) {
	original := createKlutchControlPlane
	defer func() { createKlutchControlPlane = original }()

	called := false
	var gotOpts klutchaws.CreateOptions

	createKlutchControlPlane = func(ctx context.Context, opts klutchaws.CreateOptions) {
		called = true
		gotOpts = opts
	}

	if err := runKlutchClusterCreation("aws", true); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatalf("expected createKlutchControlPlane to be invoked")
	}

	if !gotOpts.DryRun {
		t.Fatalf("expected DryRun to propagate to CreateOptions")
	}
}

func TestRunKlutchClusterCreationTrimsAndLowercases(t *testing.T) {
	original := createKlutchControlPlane
	defer func() { createKlutchControlPlane = original }()

	called := false

	createKlutchControlPlane = func(ctx context.Context, opts klutchaws.CreateOptions) {
		called = true
	}

	if err := runKlutchClusterCreation("  AWS  ", false); err != nil {
		t.Fatalf("expected provider parsing to succeed, got %v", err)
	}

	if !called {
		t.Fatalf("expected createKlutchControlPlane to be invoked after parsing provider")
	}
}

func TestRunKlutchClusterCreationRequiresProvider(t *testing.T) {
	if err := runKlutchClusterCreation("", false); err == nil {
		t.Fatalf("expected error when provider not set")
	} else if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRunKlutchClusterCreationRejectsUnsupportedProvider(t *testing.T) {
	if err := runKlutchClusterCreation("minikube", false); err == nil {
		t.Fatalf("expected error for unsupported provider")
	} else if !strings.Contains(err.Error(), "only supports") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
