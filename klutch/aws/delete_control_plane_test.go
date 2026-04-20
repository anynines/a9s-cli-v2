package aws

import (
	"context"
	"testing"
)

func TestFindKlutchVPCUsesRegion(t *testing.T) {
	origRunCmd := runCmd
	defer func() {
		runCmd = origRunCmd
	}()

	var gotArgs []string
	runCmd = func(ctx context.Context, name string, args ...string) (string, string, error) {
		gotArgs = append([]string(nil), args...)
		return "vpc-123", "", nil
	}

	vpcID := findKlutchVPC(Config{}, context.Background(), "eu-central-1")
	if vpcID != "vpc-123" {
		t.Fatalf("expected vpc-123, got %q", vpcID)
	}

	if value, ok := flagValue(gotArgs, "--region"); !ok || value != "eu-central-1" {
		t.Fatalf("expected --region eu-central-1 in args, got %v", gotArgs)
	}
}

func TestDeleteVPCUsesRegion(t *testing.T) {
	origRunCmd := runCmd
	defer func() {
		runCmd = origRunCmd
	}()

	var gotArgs []string
	runCmd = func(ctx context.Context, name string, args ...string) (string, string, error) {
		gotArgs = append([]string(nil), args...)
		return "", "", nil
	}

	deleteVPC(context.Background(), "vpc-123", DeleteOptions{Region: "eu-central-1"})

	if value, ok := flagValue(gotArgs, "--region"); !ok || value != "eu-central-1" {
		t.Fatalf("expected --region eu-central-1 in args, got %v", gotArgs)
	}
}

func flagValue(args []string, flag string) (string, bool) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return args[i+1], true
		}
	}
	return "", false
}
