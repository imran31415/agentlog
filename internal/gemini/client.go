package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gogent/internal/types"
)

// GeminiClient wraps the Google Generative AI REST API
type GeminiClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini API client using the REST API
func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &GeminiClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Close closes the Gemini client (no-op for REST API)
func (c *GeminiClient) Close() error {
	return nil
}

// GenerateContent generates content using the Gemini REST API (matches official documentation)
func (c *GeminiClient) GenerateContent(ctx context.Context, config *types.APIConfiguration, prompt, contextStr string) (*types.APIResponse, error) {
	startTime := time.Now()

	// Build the full prompt with system prompt and context
	fullPrompt := prompt
	if config.SystemPrompt != "" {
		fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s", config.SystemPrompt, prompt)
	}
	if contextStr != "" {
		fullPrompt = fmt.Sprintf("%s\n\nContext: %s", fullPrompt, contextStr)
	}

	log.Printf("Gemini REST API call - Model: %s, Prompt length: %d", config.ModelName, len(fullPrompt))

	// Build the REST API request (following official documentation format)
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": fullPrompt},
			},
			},
		},
	}

	// Add generation config if specified
	generationConfig := make(map[string]interface{})
	if config.Temperature != nil {
		generationConfig["temperature"] = *config.Temperature
	}
	if config.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *config.MaxTokens
	}
	if config.TopP != nil {
		generationConfig["topP"] = *config.TopP
	}
	if config.TopK != nil {
		generationConfig["topK"] = *config.TopK
	}
	if len(generationConfig) > 0 {
		requestBody["generationConfig"] = generationConfig
	}

	// Serialize request
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("REST API - Marshal error: %v", err)
		return &types.APIResponse{
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   fmt.Sprintf("Failed to marshal request: %v", err),
			ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
		}, nil
	}

	// Make HTTP request to Gemini REST API (following official documentation)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", config.ModelName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("REST API - Request creation error: %v", err)
		return &types.APIResponse{
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   fmt.Sprintf("Failed to create request: %v", err),
			ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("REST API - HTTP request error: %v", err)
		return &types.APIResponse{
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   fmt.Sprintf("Failed to make request: %v", err),
			ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
		}, nil
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("REST API - Response read error: %v", err)
		return &types.APIResponse{
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   fmt.Sprintf("Failed to read response: %v", err),
			ResponseTimeMs: int32(time.Since(startTime).Milliseconds()),
		}, nil
	}

	responseTime := time.Since(startTime)
	log.Printf("REST API - Response status: %d, Time: %dms", resp.StatusCode, responseTime.Milliseconds())

	if resp.StatusCode != 200 {
		log.Printf("REST API - Error response: %s", string(body))
		return &types.APIResponse{
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)),
			ResponseTimeMs: int32(responseTime.Milliseconds()),
		}, nil
	}

	// Parse response (following official documentation format)
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		log.Printf("REST API - JSON parse error: %v", err)
		return &types.APIResponse{
			ResponseStatus: types.ResponseStatusError,
			ErrorMessage:   fmt.Sprintf("Failed to parse response: %v", err),
			ResponseTimeMs: int32(responseTime.Milliseconds()),
		}, nil
	}

	// Extract response text
	var responseText string
	var finishReason string
	if len(geminiResp.Candidates) > 0 {
		candidate := geminiResp.Candidates[0]
		if len(candidate.Content.Parts) > 0 {
			responseText = candidate.Content.Parts[0].Text
		}
		finishReason = candidate.FinishReason
	}

	// Build usage metadata
	usageMetadata := map[string]interface{}{
		"prompt_tokens":     geminiResp.UsageMetadata.PromptTokenCount,
		"completion_tokens": geminiResp.UsageMetadata.CandidatesTokenCount,
		"total_tokens":      geminiResp.UsageMetadata.TotalTokenCount,
	}

	log.Printf("REST API - Success! Response length: %d chars", len(responseText))

	return &types.APIResponse{
		ResponseStatus: types.ResponseStatusSuccess,
		ResponseText:   responseText,
		UsageMetadata:  usageMetadata,
		FinishReason:   finishReason,
		ResponseTimeMs: int32(responseTime.Milliseconds()),
	}, nil
}
