package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const defaultTenantSecretPrefix = "klutch/"

// TenantSecretName resolves the secret name for a tenant, allowing override.
func TenantSecretName(tenant string, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return fmt.Sprintf("%s%s/oidc-client", defaultTenantSecretPrefix, strings.TrimSpace(tenant))
}

// ListTenantSecrets lists Secrets Manager secrets matching the tenant prefix.
func ListTenantSecrets(ctx context.Context, region string, prefix string) ([]string, error) {
	if strings.TrimSpace(prefix) == "" {
		prefix = defaultTenantSecretPrefix
	}
	query := fmt.Sprintf("SecretList[?starts_with(Name, `%s`)].Name", prefix)
	out, err := runCmd(ctx, false, false, "aws", "secretsmanager", "list-secrets",
		"--region", region,
		"--max-results", "100",
		"--query", query,
		"--output", "text")
	if err != nil {
		return nil, fmt.Errorf("could not list secrets with prefix %s: %w (stderr: %s)", prefix, err, out)
	}
	if strings.TrimSpace(out) == "" || strings.TrimSpace(out) == "None" {
		return []string{}, nil
	}
	parts := strings.Fields(out)
	return parts, nil
}

// GetTenantCredentials retrieves an OIDCConnection from Secrets Manager.
func GetTenantCredentials(ctx context.Context, region string, secretName string) (OIDCConnection, error) {
	out, err := runCmd(ctx, false, false, "aws", "secretsmanager", "get-secret-value",
		"--region", region,
		"--secret-id", secretName,
		"--query", "SecretString",
		"--output", "text")
	if err != nil {
		return OIDCConnection{}, fmt.Errorf("could not read secret %s: %w (stderr: %s)", secretName, err, out)
	}
	var conn OIDCConnection
	if err := json.Unmarshal([]byte(out), &conn); err != nil {
		return OIDCConnection{}, fmt.Errorf("secret %s has invalid format: %w", secretName, err)
	}
	return conn, nil
}

// DeleteTenantSecret deletes a tenant secret from Secrets Manager.
func DeleteTenantSecret(ctx context.Context, region string, secretName string) error {
	if errOut, err := runCmd(ctx, true, false, "aws", "secretsmanager", "delete-secret",
		"--region", region,
		"--secret-id", secretName,
		"--force-delete-without-recovery"); err != nil {
		return fmt.Errorf("could not delete secret %s: %w (stderr: %s)", secretName, err, errOut)
	}
	return nil
}

// TenantSecretExists returns true if the secret exists.
func TenantSecretExists(ctx context.Context, region, secretName string) bool {
	errOut, err := runCmd(ctx, false, false, "aws", "secretsmanager", "describe-secret",
		"--region", region,
		"--secret-id", secretName,
		"--query", "ARN",
		"--output", "text")
	if err != nil {
		if strings.Contains(strings.ToLower(errOut), "resource not found") || strings.Contains(strings.ToLower(errOut), "resource notfound") {
			return false
		}
		return false
	}
	return true
}
