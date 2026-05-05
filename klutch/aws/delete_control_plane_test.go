package aws

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"k8s.io/utils/ptr"
)

func TestFindKlutchVPCUsesRegion(t *testing.T) {
	origRunCmd := runCmd
	defer func() {
		runCmd = origRunCmd
	}()

	var gotArgs []string
	runCmd = func(ctx context.Context, _, _ bool, name string, args ...string) (string, error) {
		gotArgs = append([]string(nil), args...)
		return "vpc-123", nil
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
	runCmd = func(ctx context.Context, _, _ bool, name string, args ...string) (string, error) {
		gotArgs = append([]string(nil), args...)
		return "", nil
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

type outcome struct {
	output string
	err    error
}

func TestDeleteUserPoolByArn(t *testing.T) {
	tests := []struct {
		name               string
		arn                string
		dryRun             bool
		subcommandOutcomes map[string]*outcome
		wantErr            *string
	}{
		{
			name:    "malformed ARN without slash",
			arn:     "arn:aws:cognito-idp:us-east-1:123456789:userpool-no-slash",
			wantErr: ptr.To("malformed user pool arn"),
		},
		{
			name: "pool already gone (ResourceNotFoundException)",
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool": {"An error occurred (ResourceNotFoundException) when calling the DescribeUserPool operation: User pool us-east-1_ABCDEF does not exist.", fmt.Errorf("exit status 254")}},
		},
		{
			name: "describe-user-pool fails with other error",
			arn:  "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF",
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool": {"access denied", fmt.Errorf("exit status 1")},
			},
			wantErr: ptr.To("could not describe user pool"),
		},
		{
			name:   "dry run returns early",
			arn:    "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF",
			dryRun: true,
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool":      {"my-domain", nil},
				"delete-user-pool-domain": {"", fmt.Errorf("dry-run must not call delete commands")},
				"delete-user-pool":        {"", fmt.Errorf("dry-run must not call delete commands")},
			},
		},
		{
			name: "domain exists, deletion succeeds, pool deletion succeeds",
			arn:  "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF",
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool":      {"my-domain", nil},
				"delete-user-pool-domain": {"", nil},
				"delete-user-pool":        {"", nil},
			},
		},
		{
			name: "no domain (none), pool deletion succeeds",
			arn:  "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF",
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool": {"None", nil},
				"delete-user-pool":   {"", nil},
			},
		},
		{
			name: "domain deletion fails",
			arn:  "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF",
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool":      {"my-domain", nil},
				"delete-user-pool-domain": {"domain in use", fmt.Errorf("exit status 1")},
			},
			wantErr: ptr.To("could not delete domain"),
		},
		{
			name: "pool deletion fails",
			arn:  "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF",
			subcommandOutcomes: map[string]*outcome{
				"describe-user-pool": {"None", nil},
				"delete-user-pool":   {"internal error", fmt.Errorf("exit status 1")},
			},
			wantErr: ptr.To("could not delete Cognito user pool"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origRunCmd := runCmd
			defer func() { runCmd = origRunCmd }()

			if tc.subcommandOutcomes != nil {
				runCmd = newRunCmdFromSubcommandOutcomes(tc.subcommandOutcomes)
			}

			if tc.arn == "" {
				tc.arn = "arn:aws:cognito-idp:us-east-1:123456789:userpool/us-east-1_ABCDEF"
			}

			err := deleteUserPoolByArn(context.Background(), tc.arn, "us-east-1", tc.dryRun)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error with prefix %q, got nil", *tc.wantErr)
			}
			if !strings.HasPrefix(err.Error(), *tc.wantErr) {
				t.Fatalf("expected error containing %q, got: %v", *tc.wantErr, err)
			}
		})
	}
}

func newRunCmdFromSubcommandOutcomes(outcomes map[string]*outcome) func(ctx context.Context, showPrompt bool, suppressOutput bool, name string, args ...string) (string, error) {
	return func(ctx context.Context, showPrompt, suppressOutput bool, name string, args ...string) (string, error) {
		for _, a := range args {
			if outcomes[a] != nil {
				return outcomes[a].output, outcomes[a].err
			}
		}
		return "", fmt.Errorf("unexpected subcommand")
	}
}

func TestDeleteUserPoolsByTags(t *testing.T) {
	tests := []struct {
		name               string
		dryRun             bool
		subcommandOutcomes map[string]*outcome
		wantErr            *string
	}{
		{
			name: "discovery fails",
			subcommandOutcomes: map[string]*outcome{
				"get-resources": {"service error", fmt.Errorf("exit status 1")},
			},
			wantErr: ptr.To("could not discover user pools by tag"),
		},
		{
			name: "no pools found",
			subcommandOutcomes: map[string]*outcome{
				"get-resources": {"", nil},
			},
		},
		{
			name: "single pool, deletion succeeds",
			subcommandOutcomes: map[string]*outcome{
				"get-resources":      {"arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_AAA", nil},
				"describe-user-pool": {"None", nil},
				"delete-user-pool":   {"", nil},
			},
		},
		{
			name: "single pool, deletion fails",
			subcommandOutcomes: map[string]*outcome{
				"get-resources":      {"arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_AAA", nil},
				"describe-user-pool": {"access denied", fmt.Errorf("exit status 1")},
			},
			wantErr: ptr.To("could not delete user pool"),
		},
		{
			name: "multiple pools, all succeed",
			subcommandOutcomes: map[string]*outcome{
				"get-resources":      {"arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_AAA arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_BBB", nil},
				"describe-user-pool": {"None", nil},
				"delete-user-pool":   {"", nil},
			},
		},
		{
			name: "multiple pools, all fail returns joined errors",
			subcommandOutcomes: map[string]*outcome{
				"get-resources":      {"arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_AAA arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_BBB", nil},
				"describe-user-pool": {"access denied", fmt.Errorf("exit status 1")},
			},
			wantErr: ptr.To("could not delete user pool arn:aws:cognito-idp:us-east-1:123:userpool/us-east-1_AAA"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origRunCmd := runCmd
			defer func() { runCmd = origRunCmd }()

			runCmd = newRunCmdFromSubcommandOutcomes(tc.subcommandOutcomes)

			err := DeleteUserPoolsByTags(context.Background(), tc.dryRun, "us-east-1", "Name", "my-cluster")

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", *tc.wantErr)
			}
			if !strings.Contains(err.Error(), *tc.wantErr) {
				t.Fatalf("expected error containing %q, got: %v", *tc.wantErr, err)
			}
		})
	}
}
