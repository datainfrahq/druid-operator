package druidapi

import (
	"context"
	"testing"

	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetAuthCreds(t *testing.T) {
	tests := []struct {
		name      string
		auth      Auth
		expected  internalhttp.BasicAuth
		expectErr bool
	}{
		{
			name: "default keys present",
			auth: Auth{
				Type:      BasicAuth,
				SecretRef: v1.SecretReference{Name: "test-default", Namespace: "test"},
			},
			expected:  internalhttp.BasicAuth{UserName: "test-user", Password: "test-password"},
			expectErr: false,
		},
		{
			name: "custom keys present",
			auth: Auth{
				Type:        BasicAuth,
				SecretRef:   v1.SecretReference{Name: "test", Namespace: "default"},
				UsernameKey: "usr",
				PasswordKey: "pwd",
			},
			expected:  internalhttp.BasicAuth{UserName: "admin", Password: "admin"},
			expectErr: false,
		},
		{
			name: "custom user key is  missing",
			auth: Auth{
				Type:        BasicAuth,
				SecretRef:   v1.SecretReference{Name: "test", Namespace: "default"},
				UsernameKey: "nope",
				PasswordKey: "pwd",
			},
			expected:  internalhttp.BasicAuth{},
			expectErr: true,
		},
		{
			name: "custom user key with default password key",
			auth: Auth{
				Type:        BasicAuth,
				SecretRef:   v1.SecretReference{Name: "test", Namespace: "default"},
				UsernameKey: "usr",
			},
			expected:  internalhttp.BasicAuth{UserName: "admin", Password: "also-admin"},
			expectErr: false,
		},
		{
			name: "custom password key is missing",
			auth: Auth{
				Type:        BasicAuth,
				SecretRef:   v1.SecretReference{Name: "test", Namespace: "default"},
				UsernameKey: "usr",
				PasswordKey: "nope",
			},
			expected:  internalhttp.BasicAuth{},
			expectErr: true,
		},
		{
			name:      "empty auth struct returns no creds",
			auth:      Auth{},
			expected:  internalhttp.BasicAuth{},
			expectErr: false,
		},
	}

	client := fake.NewClientBuilder().
		WithObjects(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-default",
				Namespace: "test",
			},
			Data: map[string][]byte{
				OperatorUserName: []byte("test-user"),
				OperatorPassword: []byte("test-password"),
			},
		}).
		WithObjects(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"usr":            []byte("admin"),
				"pwd":            []byte("admin"),
				OperatorPassword: []byte("also-admin"),
			},
		}).Build()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetAuthCreds(context.TODO(), client, tt.auth)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestMakePath(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		componentType   string
		apiType         string
		additionalPaths []string
		expected        string
	}{
		{
			name:          "NoAdditionalPath",
			baseURL:       "http://example-druid-service",
			componentType: "indexer",
			apiType:       "task",
			expected:      "http://example-druid-service/druid/indexer/v1/task",
		},
		{
			name:            "OneAdditionalPath",
			baseURL:         "http://example-druid-service",
			componentType:   "indexer",
			apiType:         "task",
			additionalPaths: []string{"extra"},
			expected:        "http://example-druid-service/druid/indexer/v1/task/extra",
		},
		{
			name:            "MultipleAdditionalPaths",
			baseURL:         "http://example-druid-service",
			componentType:   "coordinator",
			apiType:         "rules",
			additionalPaths: []string{"wikipedia", "history"},
			expected:        "http://example-druid-service/druid/coordinator/v1/rules/wikipedia/history",
		},
		{
			name:          "EmptyBaseURL",
			baseURL:       "",
			componentType: "indexer",
			apiType:       "task",
			expected:      "druid/indexer/v1/task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MakePath(tt.baseURL, tt.componentType, tt.apiType, tt.additionalPaths...)
			if actual != tt.expected {
				t.Errorf("makePath() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestParseProperties(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name: "simple properties",
			input: `druid.host=localhost
druid.port=8082`,
			expected: map[string]string{
				"druid.host": "localhost",
				"druid.port": "8082",
			},
		},
		{
			name: "properties with comments and empty lines",
			input: `# This is a comment
druid.host=localhost

druid.port=8082
# Another comment
druid.service=druid/broker`,
			expected: map[string]string{
				"druid.host":    "localhost",
				"druid.port":    "8082",
				"druid.service": "druid/broker",
			},
		},
		{
			name: "properties with spaces",
			input: `druid.host = localhost
druid.port= 8082
druid.service =druid/broker`,
			expected: map[string]string{
				"druid.host":    "localhost",
				"druid.port":    "8082",
				"druid.service": "druid/broker",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseProperties(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d properties, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("expected key %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("for key %s, expected %s, got %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestInferRouterConnectionFromConfig(t *testing.T) {
	tests := []struct {
		name         string
		commonConfig string
		routerConfig string
		expected     RouterConnectionInfo
	}{
		{
			name:         "default configuration",
			commonConfig: "",
			routerConfig: "",
			expected: RouterConnectionInfo{
				Protocol: "http",
				Port:     "8088",
			},
		},
		{
			name: "HTTP with custom port in common config",
			commonConfig: `druid.enablePlaintextPort=true
druid.plaintextPort=9090`,
			routerConfig: "",
			expected: RouterConnectionInfo{
				Protocol: "http",
				Port:     "9090",
			},
		},
		{
			name: "HTTPS enabled in common config",
			commonConfig: `druid.enableTlsPort=true
druid.tlsPort=8443`,
			routerConfig: "",
			expected: RouterConnectionInfo{
				Protocol: "https",
				Port:     "8443",
			},
		},
		{
			name: "HTTPS enabled without specific TLS port",
			commonConfig: `druid.enableTlsPort=true
druid.port=8283`,
			routerConfig: "",
			expected: RouterConnectionInfo{
				Protocol: "https",
				Port:     "8283",
			},
		},
		{
			name: "router config overrides common config",
			commonConfig: `druid.enableTlsPort=false
druid.plaintextPort=8088`,
			routerConfig: `druid.enableTlsPort=true
druid.tlsPort=8443`,
			expected: RouterConnectionInfo{
				Protocol: "https",
				Port:     "8443",
			},
		},
		{
			name: "plaintext disabled, TLS enabled",
			commonConfig: `druid.enablePlaintextPort=false
druid.enableTlsPort=true
druid.tlsPort=8443`,
			routerConfig: "",
			expected: RouterConnectionInfo{
				Protocol: "https",
				Port:     "8443",
			},
		},
		{
			name: "complex configuration with router override",
			commonConfig: `# Common configuration
druid.host=localhost
druid.enableTlsPort=false
druid.plaintextPort=8088`,
			routerConfig: `# Router specific configuration
druid.service=druid/router
druid.enableTlsPort=true
druid.tlsPort=8443
druid.router.http.numConnections=50`,
			expected: RouterConnectionInfo{
				Protocol: "https",
				Port:     "8443",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferRouterConnectionFromConfig(tt.commonConfig, tt.routerConfig)

			if result.Protocol != tt.expected.Protocol {
				t.Errorf("expected protocol %s, got %s", tt.expected.Protocol, result.Protocol)
			}

			if result.Port != tt.expected.Port {
				t.Errorf("expected port %s, got %s", tt.expected.Port, result.Port)
			}
		})
	}
}
