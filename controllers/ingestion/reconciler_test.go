package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"

	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
)

func TestUpdateCompaction_Success(t *testing.T) {
	// Mock DruidIngestion data
	di := &v1alpha1.DruidIngestion{
		Spec: v1alpha1.DruidIngestionSpec{
			Ingestion: v1alpha1.IngestionSpec{
				Spec: `{"dataSource": "testDataSource"}`,
				Compaction: v1alpha1.Compaction{
					MetricsSpec: "testMetric",
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
				Compaction: v1alpha1.Compaction{
					MetricsSpec: "testMetric",
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

func TestExtractDataSource(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		expected  string
		expectErr bool
	}{
		{
			name:      "ValidDataSource",
			spec:      `{"spec": {"dataSchema": {"dataSource": "wikipedia"}}}`,
			expected:  "wikipedia",
			expectErr: false,
		},
		{
			name:      "MissingDataSource",
			spec:      `{"spec": {"dataSchema": {}}}`,
			expected:  "",
			expectErr: true,
		},
		{
			name:      "InvalidJSON",
			spec:      `{"spec": {"dataSchema": {"dataSource": "wikipedia"}`,
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di := &v1alpha1.DruidIngestion{
				Spec: v1alpha1.DruidIngestionSpec{
					Ingestion: v1alpha1.IngestionSpec{
						Spec: tt.spec,
					},
				},
			}
			actual, err := extractDataSource(di)

			if (err != nil) != tt.expectErr {
				t.Errorf("extractDataSource() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if actual != tt.expected {
				t.Errorf("extractDataSource() = %v, expected %v", actual, tt.expected)
			}
		})
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
