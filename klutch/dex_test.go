package klutch

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type MockRandomByteGenerator struct{}

func (MockRandomByteGenerator) GenerateRandom32BytesBase64() string {
	return "generated-string"
}

func TestGetOIDCIssuerClientSecret(t *testing.T) {
	type testCase struct {
		name     string
		secrets  []runtime.Object
		expected string
	}

	objectMeta := metav1.ObjectMeta{
		Name:      "oidc-config",
		Namespace: "default",
	}

	testCases := []testCase{
		{
			name: "returns the secret value",
			secrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: objectMeta,
					Data: map[string][]byte{
						"oidc-issuer-client-secret": []byte("existing-string"),
					},
				},
			},
			expected: "existing-string",
		},
		{
			name:     "returns the generated value when secret doesn't exist",
			secrets:  []runtime.Object{},
			expected: "generated-string",
		},
		{
			name: "returns the generated value when secret doesn't have the key",
			secrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: objectMeta,
					Data: map[string][]byte{
						"another-key": []byte("existing-string"),
					},
				},
			},
			expected: "generated-string",
		},
		{
			name: "returns the generated value if the value is empty",
			secrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: objectMeta,
					Data: map[string][]byte{
						"another-key": []byte(""),
					},
				},
			},
			expected: "generated-string",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			client := testclient.NewSimpleClientset(test.secrets...)
			fakeGenerator := MockRandomByteGenerator{}
			k := NewKlutchManager()

			result := k.getOIDCIssuerClientSecret(client, fakeGenerator)

			if result != test.expected {
				t.Fatalf("expected result %s, got %s", test.expected, result)
			}
		})
	}
}
