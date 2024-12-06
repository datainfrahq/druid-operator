package ingestion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"

	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestUpdateCompaction_Success(t *testing.T) {
	// Mock DruidIngestion data
	di := &v1alpha1.DruidIngestion{
		Spec: v1alpha1.DruidIngestionSpec{
			Ingestion: v1alpha1.IngestionSpec{
				Spec: `{"dataSource": "testDataSource"}`,
				Compaction: runtime.RawExtension{
					Raw: []byte(`{"metricsSpec": "testMetric"}`),
				},
			},
		},
	}

	dataSource := "testDataSource"

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Mock Auth
	auth := internalhttp.Auth{
		BasicAuth: internalhttp.BasicAuth{
			UserName: "user",
			Password: "pass",
		},
	}

	// Mock DruidIngestionReconciler
	r := &DruidIngestionReconciler{}

	// Call UpdateCompaction
	success, err := r.UpdateCompaction(di, server.URL, dataSource, auth)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !success {
		t.Fatalf("expected success, got failure")
	}
}

func TestUpdateCompaction_Failure(t *testing.T) {
	// Mock DruidIngestion data
	di := &v1alpha1.DruidIngestion{
		Spec: v1alpha1.DruidIngestionSpec{
			Ingestion: v1alpha1.IngestionSpec{
				Spec: `{"dataSource": "testDataSource"}`,
				Compaction: runtime.RawExtension{
					Raw: []byte(`{"metricsSpec": "testMetric"}`),
				},
			},
		},
	}

	dataSource := "testDataSource"

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Mock Auth
	auth := internalhttp.Auth{
		BasicAuth: internalhttp.BasicAuth{
			UserName: "user",
			Password: "pass",
		},
	}

	// Mock DruidIngestionReconciler
	r := &DruidIngestionReconciler{}

	// Call UpdateCompaction
	success, err := r.UpdateCompaction(di, server.URL, dataSource, auth)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if success {
		t.Fatalf("expected failure, got success")
	}
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		name          string
		ingestionType v1alpha1.DruidIngestionMethod
		svcName       string
		httpMethod    string
		taskId        string
		shutDownTask  bool
		expected      string
	}{
		{
			name:          "NativeBatchGetTask",
			ingestionType: v1alpha1.NativeBatchIndexParallel,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodGet,
			taskId:        "task1",
			expected:      "http://example-druid-service/druid/indexer/v1/task/task1",
		},
		{
			name:          "NativeBatchCreateUpdateTask",
			ingestionType: v1alpha1.NativeBatchIndexParallel,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodPost,
			shutDownTask:  false,
			expected:      "http://example-druid-service/druid/indexer/v1/task",
		},
		{
			name:          "NativeBatchShutdownTask",
			ingestionType: v1alpha1.NativeBatchIndexParallel,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodPost,
			taskId:        "task1",
			shutDownTask:  true,
			expected:      "http://example-druid-service/druid/indexer/v1/task/task1/shutdown",
		},
		{
			name:          "KafkaGetSupervisorTask",
			ingestionType: v1alpha1.Kafka,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodGet,
			taskId:        "supervisor1",
			expected:      "http://example-druid-service/druid/indexer/v1/supervisor/supervisor1",
		},
		{
			name:          "KafkaCreateUpdateSupervisorTask",
			ingestionType: v1alpha1.Kafka,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodPost,
			shutDownTask:  false,
			expected:      "http://example-druid-service/druid/indexer/v1/supervisor",
		},
		{
			name:          "KafkaShutdownSupervisor",
			ingestionType: v1alpha1.Kafka,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodPost,
			taskId:        "supervisor1",
			shutDownTask:  true,
			expected:      "http://example-druid-service/druid/indexer/v1/supervisor/supervisor1/shutdown",
		},
		{
			name:          "UnsupportedIngestionType",
			ingestionType: v1alpha1.Kinesis,
			svcName:       "http://example-druid-service",
			httpMethod:    http.MethodGet,
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getPath(tt.ingestionType, tt.svcName, tt.httpMethod, tt.taskId, tt.shutDownTask)
			if actual != tt.expected {
				t.Errorf("getPath() = %v, expected %v", actual, tt.expected)
			}
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
			actual := makePath(tt.baseURL, tt.componentType, tt.apiType, tt.additionalPaths...)
			if actual != tt.expected {
				t.Errorf("makePath() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestGetCurrentSpec(t *testing.T) {
	tests := []struct {
		name           string
		di             *v1alpha1.DruidIngestion
		expectedSpec   map[string]interface{}
		expectingError bool
	}{
		{
			name: "NativeSpec is used",
			di: &v1alpha1.DruidIngestion{
				Spec: v1alpha1.DruidIngestionSpec{
					Ingestion: v1alpha1.IngestionSpec{
						NativeSpec: runtime.RawExtension{
							Raw: []byte(`{"key": "value"}`),
						},
					},
				},
			},
			expectedSpec: map[string]interface{}{
				"key": "value",
			},
			expectingError: false,
		},
		{
			name: "Spec is used when NativeSpec is empty",
			di: &v1alpha1.DruidIngestion{
				Spec: v1alpha1.DruidIngestionSpec{
					Ingestion: v1alpha1.IngestionSpec{
						Spec: `{"key": "value"}`,
					},
				},
			},
			expectedSpec: map[string]interface{}{
				"key": "value",
			},
			expectingError: false,
		},
		{
			name: "Error when both NativeSpec and Spec are empty",
			di: &v1alpha1.DruidIngestion{
				Spec: v1alpha1.DruidIngestionSpec{
					Ingestion: v1alpha1.IngestionSpec{
						NativeSpec: runtime.RawExtension{},
						Spec:       "",
					},
				},
			},
			expectedSpec:   nil,
			expectingError: true,
		},
		{
			name: "Error when Spec is invalid JSON",
			di: &v1alpha1.DruidIngestion{
				Spec: v1alpha1.DruidIngestionSpec{
					Ingestion: v1alpha1.IngestionSpec{
						Spec: `{"key": "value"`, // Invalid JSON
					},
				},
			},
			expectedSpec:   nil,
			expectingError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := getCurrentSpec(tt.di)
			if tt.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSpec, spec)
			}
		})
	}
}

func TestExtractDataSourceFromSpec(t *testing.T) {
	tests := []struct {
		name         string
		specJSON     string
		expected     string
		expectingErr bool
	}{
		{
			name: "Valid dataSource extraction",
			specJSON: `
            {
                "spec": {
                    "dataSchema": {
                        "dataSource": "wikipedia-2"
                    }
                }
            }`,
			expected:     "wikipedia-2",
			expectingErr: false,
		},
		{
			name: "Missing dataSource",
			specJSON: `
            {
                "spec": {
                    "dataSchema": {}
                }
            }`,
			expected:     "",
			expectingErr: true,
		},
		{
			name: "Incorrect dataSource type",
			specJSON: `
            {
                "spec": {
                    "dataSchema": {
                        "dataSource": 123
                    }
                }
            }`,
			expected:     "",
			expectingErr: true,
		},
		{
			name: "Missing spec section",
			specJSON: `
            {
                "otherField": {}
            }`,
			expected:     "",
			expectingErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var specData map[string]interface{}
			err := json.Unmarshal([]byte(tt.specJSON), &specData)
			assert.NoError(t, err, "Failed to unmarshal JSON")

			dataSource, err := extractDataSourceFromSpec(specData)
			if tt.expectingErr {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expected, dataSource, "DataSource does not match expected value")
			}
		})
	}
}
