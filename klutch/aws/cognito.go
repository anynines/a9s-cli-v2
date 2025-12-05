package aws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// OIDCConnection holds the issuer and client credentials for Cognito.
type OIDCConnection struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	Scope        string
}

// EnsureCognitoOIDC provisions (or reuses) a minimal Cognito setup for client-credentials:
// - user pool (existing or created)
// - resource server with scope klutch/bind
// - app client with secret and client_credentials flow
// - Amazon-hosted domain
func EnsureCognitoOIDC(ctx context.Context, region string, namePrefix string, userPoolID string) (OIDCConnection, error) {
	prefix := strings.ToLower(strings.TrimSpace(namePrefix))
	if prefix == "" {
		prefix = "klutch"
	}

	if err := ensureCognitoOAuthSupport(ctx); err != nil {
		return OIDCConnection{}, err
	}

	userPoolName := fmt.Sprintf("%s-klutch", prefix)
	resourceServerID := "klutch"
	resourceScope := "klutch/bind"
	clientName := fmt.Sprintf("%s-konnector", prefix)

	poolID := strings.TrimSpace(userPoolID)
	if poolID == "" {
		poolID = discoverUserPool(ctx, region, userPoolName)
		if poolID == "" {
			poolID = mustRun(ctx, "aws", "cognito-idp", "create-user-pool",
				"--region", region,
				"--pool-name", userPoolName,
				"--query", "UserPool.Id",
				"--output", "text")
		}
	} else {
		// Validate provided pool ID
		if _, errOut, err := runCmd(ctx, "aws", "cognito-idp", "describe-user-pool",
			"--region", region,
			"--user-pool-id", poolID,
			"--query", "UserPool.Id",
			"--output", "text"); err != nil {
			return OIDCConnection{}, fmt.Errorf("provided user pool id %s is not accessible: %w (stderr: %s)", poolID, err, errOut)
		}
	}

	// Create resource server + scope (best effort).
	_, _, _ = runCmd(ctx, "aws", "cognito-idp", "create-resource-server",
		"--region", region,
		"--user-pool-id", poolID,
		"--identifier", resourceServerID,
		"--name", "Klutch",
		"--scopes", fmt.Sprintf("ScopeName=bind,ScopeDescription=%q", "Klutch bind"))

	client, err := discoverOrCreateClient(ctx, region, poolID, clientName, resourceScope)
	if err != nil {
		return OIDCConnection{}, err
	}

	domain := ensureUserPoolDomain(ctx, region, poolID, prefix)
	_ = domain // domain is still useful for browser flows; issuer comes from the identity provider endpoint.

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, poolID)

	return OIDCConnection{
		IssuerURL:    issuer,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Scope:        resourceScope,
	}, nil
}

func discoverUserPool(ctx context.Context, region, name string) string {
	out, _, _ := runCmd(ctx, "aws", "cognito-idp", "list-user-pools",
		"--region", region,
		"--max-results", "50",
		"--query", fmt.Sprintf("UserPools[?Name==`%s`].Id | [0]", name),
		"--output", "text")
	if out == "None" || out == "null" {
		return ""
	}
	return strings.TrimSpace(out)
}

type oidcClient struct {
	ClientID     string `json:"ClientId"`
	ClientSecret string `json:"ClientSecret"`
}

func discoverOrCreateClient(ctx context.Context, region, poolID, name, scope string) (oidcClient, error) {
	clientID := ""
	if out, _, _ := runCmd(ctx, "aws", "cognito-idp", "list-user-pool-clients",
		"--region", region,
		"--user-pool-id", poolID,
		"--query", fmt.Sprintf("UserPoolClients[?ClientName==`%s`].ClientId | [0]", name),
		"--output", "text"); out != "" && out != "None" && out != "null" {
		clientID = strings.TrimSpace(out)
	}

	if clientID == "" {
		out, errOut, err := runCmd(ctx, "aws", "cognito-idp", "create-user-pool-client",
			"--region", region,
			"--user-pool-id", poolID,
			"--client-name", name,
			"--generate-secret",
			"--allowed-o-auth-flows-user-pool-client",
			"--allowed-o-auth-flows", "client_credentials",
			"--allowed-o-auth-scopes", scope,
			"--supported-identity-providers", "COGNITO",
			"--query", "UserPoolClient.{ClientId:ClientId,ClientSecret:ClientSecret}",
			"--output", "json")
		if err != nil {
			if strings.Contains(errOut, "Unknown options") || strings.Contains(errOut, "unknown option") {
				return oidcClient{}, fmt.Errorf("your AWS CLI does not support Cognito OAuth flags. Please upgrade awscli v2 (or newer) or provide OIDC values via flags (issuer/client id/secret). stderr: %s", errOut)
			}
			return oidcClient{}, fmt.Errorf("create-user-pool-client failed: %w (stderr: %s)", err, errOut)
		}
		var c oidcClient
		_ = json.Unmarshal([]byte(out), &c)
		return c, nil
	}

	out := mustRun(ctx, "aws", "cognito-idp", "describe-user-pool-client",
		"--region", region,
		"--user-pool-id", poolID,
		"--client-id", clientID,
		"--query", "UserPoolClient.{ClientId:ClientId,ClientSecret:ClientSecret}",
		"--output", "json")
	var c oidcClient
	_ = json.Unmarshal([]byte(out), &c)
	return c, nil
}

func ensureUserPoolDomain(ctx context.Context, region, poolID, prefix string) string {
	for i := 0; i < 5; i++ {
		suffix := randomHex(3)
		domain := fmt.Sprintf("%s-%s", prefix, suffix)
		if _, _, err := runCmd(ctx, "aws", "cognito-idp", "create-user-pool-domain",
			"--region", region,
			"--domain", domain,
			"--user-pool-id", poolID); err == nil {
			return domain
		}
		// If domain exists, try another suffix; ignore other errors for now.
	}
	// Fallback: describe any existing domain for the pool.
	out, _, _ := runCmd(ctx, "aws", "cognito-idp", "describe-user-pool",
		"--region", region,
		"--user-pool-id", poolID,
		"--query", "UserPool.Domain",
		"--output", "text")
	if out == "None" || out == "null" || strings.TrimSpace(out) == "" {
		return fmt.Sprintf("%s-%s", prefix, randomHex(4))
	}
	return strings.TrimSpace(out)
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "abc123"
	}
	return hex.EncodeToString(b)
}

// ensureCognitoOAuthSupport checks that the AWS CLI supports the OAuth flags needed for client_credentials.
func ensureCognitoOAuthSupport(ctx context.Context) error {
	out, errOut, err := runCmd(ctx, "aws", "cognito-idp", "create-user-pool-client", "help")
	if err != nil {
		return fmt.Errorf("aws cli help failed: %w (stderr: %s)", err, errOut)
	}
	if !strings.Contains(out+errOut, "allowed-o-auth-flows-user-pool-client") {
		return fmt.Errorf("installed aws cli does not support Cognito OAuth flags (allowed-o-auth-flows-user-pool-client). Please upgrade awscli v2 or provide OIDC issuer/client credentials via flags.")
	}
	return nil
}

// StoreCognitoCredentialsSecret stores OIDC client credentials in AWS Secrets Manager (create or update).
func StoreCognitoCredentialsSecret(ctx context.Context, region, secretName string, conn OIDCConnection) error {
	payload := fmt.Sprintf(`{"issuer":"%s","client_id":"%s","client_secret":"%s","scope":"%s"}`, conn.IssuerURL, conn.ClientID, conn.ClientSecret, conn.Scope)

	if _, errOut, err := runCmd(ctx, "aws", "secretsmanager", "create-secret",
		"--region", region,
		"--name", secretName,
		"--secret-string", payload); err != nil {
		if strings.Contains(strings.ToLower(errOut), "resourceexistsexception") {
			if _, errOut2, err2 := runCmd(ctx, "aws", "secretsmanager", "put-secret-value",
				"--region", region,
				"--secret-id", secretName,
				"--secret-string", payload); err2 != nil {
				return fmt.Errorf("could not update existing secret %s: %w (stderr: %s)", secretName, err2, errOut2)
			}
			return nil
		}
		return fmt.Errorf("could not create secret %s: %w (stderr: %s)", secretName, err, errOut)
	}
	return nil
}
