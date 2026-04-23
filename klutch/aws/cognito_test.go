package aws

import (
	"context"
	"strings"
	"testing"
)

func TestFirstAWSValue(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "none-only", in: "None\nNone", want: ""},
		{name: "null-only", in: "null", want: ""},
		{name: "first-valid", in: "None\nabc123", want: "abc123"},
		{name: "trim-whitespace", in: "  abc123  ", want: "abc123"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstAWSValue(tc.in); got != tc.want {
				t.Fatalf("firstAWSValue(%q): expected %q, got %q", tc.in, tc.want, got)
			}
		})
	}
}

func TestDiscoverOrCreateClientCreatesWhenListReturnsNoneTokens(t *testing.T) {
	origRunCmd := runCmd
	defer func() {
		runCmd = origRunCmd
	}()

	createCalled := false
	runCmd = func(ctx context.Context, withPrompt bool, suppressOutput bool, name string, args ...string) (string, error) {
		if name != "aws" {
			t.Fatalf("expected aws command, got %s", name)
		}
		if len(args) < 2 || args[0] != "cognito-idp" {
			t.Fatalf("unexpected args: %v", args)
		}

		if withPrompt {
			switch args[1] {
			case "create-user-pool-client":
				createCalled = true
				return `{"ClientId":"new-client-id","ClientSecret":"new-secret"}`, nil
			default:
				t.Fatalf("unexpected runCmdWithPrompt command: %s", strings.Join(args, " "))
			}
			return "", nil
		}

		switch args[1] {
		case "list-user-pool-clients":
			return "None\nNone", nil
		case "describe-user-pool-client":
			t.Fatalf("describe-user-pool-client should not be called when no client exists")
		default:
			t.Fatalf("unexpected runCmd command: %s", strings.Join(args, " "))
		}
		return "", nil
	}

	client, err := discoverOrCreateClient(context.Background(), "eu-central-1", "pool-1", "tenant-a", "klutch/bind")
	if err != nil {
		t.Fatalf("discoverOrCreateClient returned error: %v", err)
	}
	if !createCalled {
		t.Fatalf("expected create-user-pool-client to be called")
	}
	if client.ClientID != "new-client-id" || client.ClientSecret != "new-secret" {
		t.Fatalf("unexpected client credentials: %#v", client)
	}
}

func TestDiscoverOrCreateClientUsesExistingClientIDFromPagedOutput(t *testing.T) {
	origRunCmd := runCmd
	defer func() {
		runCmd = origRunCmd
	}()

	describeCalled := false
	createCalled := false
	runCmd = func(ctx context.Context, withPrompt bool, _ bool, name string, args ...string) (string, error) {
		if name != "aws" {
			t.Fatalf("expected aws command, got %s", name)
		}
		if len(args) < 2 || args[0] != "cognito-idp" {
			t.Fatalf("unexpected args: %v", args)
		}

		if withPrompt {
			switch args[1] {
			case "create-user-pool-client":
				createCalled = true
				t.Fatalf("create-user-pool-client should not be called when client id exists")
			default:
				t.Fatalf("unexpected runCmdWithPrompt command: %s", strings.Join(args, " "))
			}
			return "", nil
		}

		switch args[1] {
		case "list-user-pool-clients":
			return "None\nexisting-client-id", nil
		case "describe-user-pool-client":
			describeCalled = true
			return `{"ClientId":"existing-client-id","ClientSecret":"existing-secret"}`, nil
		default:
			t.Fatalf("unexpected runCmd command: %s", strings.Join(args, " "))
		}
		return "", nil
	}

	client, err := discoverOrCreateClient(context.Background(), "eu-central-1", "pool-1", "tenant-a", "klutch/bind")
	if err != nil {
		t.Fatalf("discoverOrCreateClient returned error: %v", err)
	}
	if !describeCalled {
		t.Fatalf("expected describe-user-pool-client to be called")
	}
	if createCalled {
		t.Fatalf("create-user-pool-client should not be called")
	}
	if client.ClientID != "existing-client-id" || client.ClientSecret != "existing-secret" {
		t.Fatalf("unexpected client credentials: %#v", client)
	}
}
