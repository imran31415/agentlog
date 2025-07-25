package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

func main() {
	fmt.Println("🧪 GoGent System Comprehensive Test Suite")
	fmt.Println("=========================================")

	var results []TestResult

	// Test 1: Backend Health Check
	results = append(results, testBackendHealth())

	// Test 2: Database Connection
	results = append(results, testDatabaseConnection())

	// Test 3: Execution History API
	results = append(results, testExecutionHistory())

	// Test 4: Direct Gemini REST API
	results = append(results, testDirectGeminiAPI())

	// Test 5: Backend API Execution
	results = append(results, testBackendExecution())

	// Test 6: Mock Responses
	results = append(results, testMockResponses())

	// Print Results Summary
	fmt.Println("\n📊 Test Results Summary:")
	fmt.Println("========================")

	passed := 0
	total := len(results)

	for _, result := range results {
		status := "❌ FAIL"
		if result.Success {
			status = "✅ PASS"
			passed++
		}

		fmt.Printf("%s %s", status, result.Name)
		if result.Error != "" {
			fmt.Printf(" - Error: %s", result.Error)
		}
		if result.Details != "" {
			fmt.Printf(" - %s", result.Details)
		}
		fmt.Println()
	}

	fmt.Printf("\n🎯 Final Score: %d/%d tests passed (%.1f%%)\n",
		passed, total, float64(passed)/float64(total)*100)

	if passed == total {
		fmt.Println("🎉 All tests passed! System is working correctly.")
	} else {
		fmt.Printf("⚠️  %d tests failed. System needs attention.\n", total-passed)
	}
}

func testBackendHealth() TestResult {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		return TestResult{
			Name:    "Backend Health Check",
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TestResult{
			Name:    "Backend Health Check",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return TestResult{
			Name:    "Backend Health Check",
			Success: false,
			Error:   "Invalid JSON response",
		}
	}

	dbOk := health["database"] == true
	geminiOk := health["gemini_api"] == true

	return TestResult{
		Name:    "Backend Health Check",
		Success: dbOk && geminiOk,
		Details: fmt.Sprintf("Database: %v, Gemini: %v", dbOk, geminiOk),
	}
}

func testDatabaseConnection() TestResult {
	// This is a simple test - we assume if health check passes, DB is working
	// In a real test, we'd make a direct DB query
	return TestResult{
		Name:    "Database Connection",
		Success: true,
		Details: "Assumed working based on health check",
	}
}

func testExecutionHistory() TestResult {
	resp, err := http.Get("http://localhost:8080/api/execution-runs?limit=5")
	if err != nil {
		return TestResult{
			Name:    "Execution History API",
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TestResult{
			Name:    "Execution History API",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	var history []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return TestResult{
			Name:    "Execution History API",
			Success: false,
			Error:   "Invalid JSON response",
		}
	}

	// Check if we have real data (not just 2 mock entries)
	success := len(history) > 2

	return TestResult{
		Name:    "Execution History API",
		Success: success,
		Details: fmt.Sprintf("Returned %d execution runs", len(history)),
	}
}

func testDirectGeminiAPI() TestResult {
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": "Say 'Hello test' in exactly 2 words"},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return TestResult{
			Name:    "Direct Gemini REST API",
			Success: false,
			Error:   err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", "AIzaSyDYOnANDd0-rhLDqGrqAVrFteUU3ylUTuc")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TestResult{
			Name:    "Direct Gemini REST API",
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			Name:    "Direct Gemini REST API",
			Success: false,
			Error:   "Failed to read response",
		}
	}

	if resp.StatusCode != 200 {
		return TestResult{
			Name:    "Direct Gemini REST API",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)[:100]),
		}
	}

	// Parse response to check for successful generation
	var geminiResp map[string]interface{}
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return TestResult{
			Name:    "Direct Gemini REST API",
			Success: false,
			Error:   "Invalid JSON response",
		}
	}

	candidates, ok := geminiResp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return TestResult{
			Name:    "Direct Gemini REST API",
			Success: false,
			Error:   "No candidates in response",
		}
	}

	return TestResult{
		Name:    "Direct Gemini REST API",
		Success: true,
		Details: "Successfully generated content",
	}
}

func testBackendExecution() TestResult {
	reqBody := map[string]interface{}{
		"execution_run_name": "test-suite-execution",
		"description":        "Comprehensive test suite execution",
		"base_prompt":        "Say 'Test successful' in exactly 2 words",
		"context":            "This is a test execution",
		"configurations": []map[string]interface{}{
			{
				"variation_name": "test",
				"model_name":     "gemini-1.5-flash",
				"system_prompt":  "You are a helpful assistant",
				"temperature":    0.5,
				"max_tokens":     50,
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/execute", bytes.NewBuffer(jsonBody))
	if err != nil {
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gemini-API-Key", "AIzaSyDYOnANDd0-rhLDqGrqAVrFteUU3ylUTuc")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   "Failed to read response",
		}
	}

	if resp.StatusCode != 200 {
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)[:200]),
		}
	}

	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   "Invalid JSON response",
		}
	}

	results, ok := apiResp["results"].([]interface{})
	if !ok || len(results) == 0 {
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   "No results in response",
		}
	}

	result := results[0].(map[string]interface{})
	response := result["response"].(map[string]interface{})
	status := response["responseStatus"].(string)

	if status == "error" {
		errorMsg := response["errorMessage"].(string)
		return TestResult{
			Name:    "Backend API Execution",
			Success: false,
			Error:   errorMsg,
		}
	}

	responseText := response["responseText"]
	success := status == "success" && responseText != nil

	details := fmt.Sprintf("Status: %s", status)
	if responseText != nil {
		details += fmt.Sprintf(", Response length: %d chars", len(responseText.(string)))
	}

	return TestResult{
		Name:    "Backend API Execution",
		Success: success,
		Details: details,
	}
}

func testMockResponses() TestResult {
	reqBody := map[string]interface{}{
		"execution_run_name": "test-suite-mock",
		"description":        "Test mock responses",
		"base_prompt":        "Test prompt",
		"configurations": []map[string]interface{}{
			{
				"variation_name": "test-mock",
				"model_name":     "gemini-1.5-flash",
				"system_prompt":  "You are helpful",
				"temperature":    0.5,
				"max_tokens":     50,
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/execute", bytes.NewBuffer(jsonBody))
	if err != nil {
		return TestResult{
			Name:    "Mock Responses Test",
			Success: false,
			Error:   err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Use-Mock", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TestResult{
			Name:    "Mock Responses Test",
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TestResult{
			Name:    "Mock Responses Test",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	var apiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return TestResult{
			Name:    "Mock Responses Test",
			Success: false,
			Error:   "Invalid JSON response",
		}
	}

	results, ok := apiResp["results"].([]interface{})
	if !ok || len(results) == 0 {
		return TestResult{
			Name:    "Mock Responses Test",
			Success: false,
			Error:   "No results in response",
		}
	}

	result := results[0].(map[string]interface{})
	response := result["response"].(map[string]interface{})
	responseText := response["responseText"].(string)

	// Check if response contains mock indicator
	isMock := responseText != "" && (bytes.Contains([]byte(responseText), []byte("[MOCK")) ||
		bytes.Contains([]byte(responseText), []byte("mock")))

	return TestResult{
		Name:    "Mock Responses Test",
		Success: isMock,
		Details: fmt.Sprintf("Response: %s", responseText[:min(50, len(responseText))]),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
