package cmd

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestBuildKlutchPGInstanceManifest(t *testing.T) {
	manifest, err := buildKlutchPGInstanceManifest(
		"pg-claim",
		"workloads",
		"a9s-postgresql13",
		"postgresql-single-nano",
		"Internal",
		"a8s-postgresql",
	)
	if err != nil {
		t.Fatalf("expected manifest generation to succeed, got error: %v", err)
	}

	var decoded map[string]interface{}
	if err := yaml.Unmarshal(manifest, &decoded); err != nil {
		t.Fatalf("expected valid YAML output, got error: %v", err)
	}

	if got := decoded["apiVersion"]; got != klutchPGAPIVersion {
		t.Fatalf("expected apiVersion %q, got %#v", klutchPGAPIVersion, got)
	}
	if got := decoded["kind"]; got != klutchPGInstanceKind {
		t.Fatalf("expected kind %q, got %#v", klutchPGInstanceKind, got)
	}

	metadata := decoded["metadata"].(map[string]interface{})
	if got := metadata["name"]; got != "pg-claim" {
		t.Fatalf("expected metadata.name pg-claim, got %#v", got)
	}
	if got := metadata["namespace"]; got != "workloads" {
		t.Fatalf("expected metadata.namespace workloads, got %#v", got)
	}

	spec := decoded["spec"].(map[string]interface{})
	if got := spec["service"]; got != "a9s-postgresql13" {
		t.Fatalf("expected spec.service a9s-postgresql13, got %#v", got)
	}
	if got := spec["plan"]; got != "postgresql-single-nano" {
		t.Fatalf("expected spec.plan postgresql-single-nano, got %#v", got)
	}
	if got := spec["expose"]; got != "Internal" {
		t.Fatalf("expected spec.expose Internal, got %#v", got)
	}

	composition := spec["compositionRef"].(map[string]interface{})
	if got := composition["name"]; got != "a8s-postgresql" {
		t.Fatalf("expected spec.compositionRef.name a8s-postgresql, got %#v", got)
	}
}

func TestBuildKlutchPGServiceBindingManifest(t *testing.T) {
	manifest, err := buildKlutchPGServiceBindingManifest(
		"pg-binding",
		"workloads",
		"pg-claim",
		"postgresql",
		"a8s-servicebinding",
	)
	if err != nil {
		t.Fatalf("expected manifest generation to succeed, got error: %v", err)
	}

	var decoded map[string]interface{}
	if err := yaml.Unmarshal(manifest, &decoded); err != nil {
		t.Fatalf("expected valid YAML output, got error: %v", err)
	}

	if got := decoded["apiVersion"]; got != klutchPGAPIVersion {
		t.Fatalf("expected apiVersion %q, got %#v", klutchPGAPIVersion, got)
	}
	if got := decoded["kind"]; got != klutchPGServiceBindingKind {
		t.Fatalf("expected kind %q, got %#v", klutchPGServiceBindingKind, got)
	}

	metadata := decoded["metadata"].(map[string]interface{})
	if got := metadata["name"]; got != "pg-binding" {
		t.Fatalf("expected metadata.name pg-binding, got %#v", got)
	}
	if got := metadata["namespace"]; got != "workloads" {
		t.Fatalf("expected metadata.namespace workloads, got %#v", got)
	}

	spec := decoded["spec"].(map[string]interface{})
	if got := spec["instanceRef"]; got != "pg-claim" {
		t.Fatalf("expected spec.instanceRef pg-claim, got %#v", got)
	}
	if got := spec["serviceInstanceType"]; got != "postgresql" {
		t.Fatalf("expected spec.serviceInstanceType postgresql, got %#v", got)
	}

	composition := spec["compositionRef"].(map[string]interface{})
	if got := composition["name"]; got != "a8s-servicebinding" {
		t.Fatalf("expected spec.compositionRef.name a8s-servicebinding, got %#v", got)
	}
}
