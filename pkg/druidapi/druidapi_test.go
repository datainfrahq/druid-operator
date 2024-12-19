package druidapi

import (
	"testing"
)

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
