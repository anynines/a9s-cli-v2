package aws

import "testing"

func TestParseImageRef(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		registry   string
		repository string
		tag        string
		digest     string
	}{
		{
			name:       "ecr-tag",
			input:      "032720848313.dkr.ecr.eu-central-1.amazonaws.com/a9s-tenants-operator:0.1.4",
			registry:   "032720848313.dkr.ecr.eu-central-1.amazonaws.com",
			repository: "a9s-tenants-operator",
			tag:        "0.1.4",
		},
		{
			name:       "digest",
			input:      "ghcr.io/org/op@sha256:deadbeef",
			registry:   "ghcr.io",
			repository: "org/op",
			digest:     "sha256:deadbeef",
		},
		{
			name:       "implicit-latest",
			input:      "busybox",
			repository: "busybox",
			tag:        "latest",
		},
		{
			name:       "localhost-tag",
			input:      "localhost:5000/my/op:1.2.3",
			registry:   "localhost:5000",
			repository: "my/op",
			tag:        "1.2.3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseImageRef(tc.input)
			if got.registry != tc.registry {
				t.Fatalf("registry: expected %q, got %q", tc.registry, got.registry)
			}
			if got.repository != tc.repository {
				t.Fatalf("repository: expected %q, got %q", tc.repository, got.repository)
			}
			if got.tag != tc.tag {
				t.Fatalf("tag: expected %q, got %q", tc.tag, got.tag)
			}
			if got.digest != tc.digest {
				t.Fatalf("digest: expected %q, got %q", tc.digest, got.digest)
			}
		})
	}
}

func TestECRRegionFromHost(t *testing.T) {
	cases := []struct {
		host     string
		expected string
	}{
		{"032720848313.dkr.ecr.eu-central-1.amazonaws.com", "eu-central-1"},
		{"123456789012.dkr.ecr.us-gov-west-1.amazonaws.com", "us-gov-west-1"},
		{"example.com", ""},
	}

	for _, tc := range cases {
		if got := ecrRegionFromHost(tc.host); got != tc.expected {
			t.Fatalf("host %q: expected %q, got %q", tc.host, tc.expected, got)
		}
	}
}

func TestIsOCIChartDescriptorError(t *testing.T) {
	if !isOCIChartDescriptorError("manifest does not contain minimum number of descriptors (2), descriptors found: 1") {
		t.Fatalf("expected OCI descriptor error to be detected")
	}
	if isOCIChartDescriptorError("some other error") {
		t.Fatalf("expected unrelated error to be ignored")
	}
}
