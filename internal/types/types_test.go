package types

import (
	"testing"
	"time"
)

func TestToJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			wantErr: false,
		},
		{
			name: "API configuration",
			input: APIConfiguration{
				ID:            "test-id",
				VariationName: "test-variation",
				ModelName:     "gemini-1.5-flash",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := ToJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && jsonStr == "" {
				t.Errorf("ToJSON() returned empty string for valid input")
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "valid JSON to map",
			jsonStr: `{"key1": "value1", "key2": 42}`,
			target:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "empty string",
			jsonStr: "",
			target:  &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			jsonStr: `{"key1": value1}`, // invalid JSON
			target:  &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromJSON(tt.jsonStr, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequestType(t *testing.T) {
	tests := []struct {
		name string
		rt   RequestType
		want string
	}{
		{"generate", RequestTypeGenerate, "generate"},
		{"chat", RequestTypeChat, "chat"},
		{"function_call", RequestTypeFunctionCall, "function_call"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.rt) != tt.want {
				t.Errorf("RequestType %v = %v, want %v", tt.rt, string(tt.rt), tt.want)
			}
		})
	}
}

func TestResponseStatus(t *testing.T) {
	tests := []struct {
		name string
		rs   ResponseStatus
		want string
	}{
		{"success", ResponseStatusSuccess, "success"},
		{"error", ResponseStatusError, "error"},
		{"timeout", ResponseStatusTimeout, "timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.rs) != tt.want {
				t.Errorf("ResponseStatus %v = %v, want %v", tt.rs, string(tt.rs), tt.want)
			}
		})
	}
}

func TestExecutionRunCreation(t *testing.T) {
	now := time.Now()
	run := ExecutionRun{
		ID:          "test-run-id",
		Name:        "test-run",
		Description: "test description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if run.ID != "test-run-id" {
		t.Errorf("Expected ID to be 'test-run-id', got %s", run.ID)
	}
	if run.Name != "test-run" {
		t.Errorf("Expected Name to be 'test-run', got %s", run.Name)
	}
	if run.Description != "test description" {
		t.Errorf("Expected Description to be 'test description', got %s", run.Description)
	}
}

func TestAPIConfigurationWithPointers(t *testing.T) {
	temp := float32(0.7)
	maxTokens := int32(100)
	topP := float32(0.9)
	topK := int32(40)

	config := APIConfiguration{
		ID:            "config-id",
		VariationName: "test-variation",
		ModelName:     "gemini-1.5-flash",
		Temperature:   &temp,
		MaxTokens:     &maxTokens,
		TopP:          &topP,
		TopK:          &topK,
	}

	if config.Temperature == nil || *config.Temperature != 0.7 {
		t.Errorf("Expected Temperature to be 0.7, got %v", config.Temperature)
	}
	if config.MaxTokens == nil || *config.MaxTokens != 100 {
		t.Errorf("Expected MaxTokens to be 100, got %v", config.MaxTokens)
	}
	if config.TopP == nil || *config.TopP != 0.9 {
		t.Errorf("Expected TopP to be 0.9, got %v", config.TopP)
	}
	if config.TopK == nil || *config.TopK != 40 {
		t.Errorf("Expected TopK to be 40, got %v", config.TopK)
	}
}
