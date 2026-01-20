package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mdwit/spec2llms/internal/config"
	"github.com/mdwit/spec2llms/internal/parser"
)

func TestGenerate(t *testing.T) {
	api := &parser.API{
		Title:       "Test API",
		Description: "Test API description",
		Version:     "1.0.0",
		BaseURL:     "https://api.test.com",
		Tags: []parser.Tag{
			{Name: "users", Description: "User operations"},
		},
		Endpoints: []parser.Endpoint{
			{
				Method:      "GET",
				Path:        "/users",
				Summary:     "List users",
				Description: "Get list of all users",
				Tags:        []string{"users"},
				Parameters: []parser.Parameter{
					{Name: "limit", In: "query", Type: "integer", Description: "Max results"},
				},
				Responses: map[string]parser.Response{
					"200": {Description: "Success"},
				},
			},
			{
				Method:      "POST",
				Path:        "/users",
				Summary:     "Create user",
				Tags:        []string{"users"},
				RequestBody: &parser.RequestBody{
					Description: "User data",
					Content: map[string]parser.MediaType{
						"application/json": {
							Schema: &parser.Schema{
								Type: "object",
								Properties: map[string]*parser.Schema{
									"name":  {Type: "string"},
									"email": {Type: "string", Format: "email"},
								},
							},
						},
					},
				},
				Responses: map[string]parser.Response{
					"201": {Description: "Created"},
				},
			},
		},
		SecuritySchemes: []parser.SecurityScheme{
			{Name: "apiKey", Type: "apiKey", In: "header", ParamName: "X-API-Key"},
		},
	}

	tmpDir := t.TempDir()
	cfg := &config.Config{
		Output:   tmpDir,
		Language: "en",
	}

	gen := New(cfg, api)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Проверяем что файлы созданы
	llmsPath := filepath.Join(tmpDir, "llms.txt")
	if _, err := os.Stat(llmsPath); os.IsNotExist(err) {
		t.Error("llms.txt not created")
	}

	endpointsDir := filepath.Join(tmpDir, "endpoints")
	if _, err := os.Stat(endpointsDir); os.IsNotExist(err) {
		t.Error("endpoints directory not created")
	}

	// Проверяем что файлы для каждого endpoint созданы
	getUsersFile := filepath.Join(endpointsDir, "get-users.txt")
	if _, err := os.Stat(getUsersFile); os.IsNotExist(err) {
		t.Error("get-users.txt not created")
	}
	postUsersFile := filepath.Join(endpointsDir, "post-users.txt")
	if _, err := os.Stat(postUsersFile); os.IsNotExist(err) {
		t.Error("post-users.txt not created")
	}

	// Проверяем содержимое llms.txt
	llmsContent, err := os.ReadFile(llmsPath)
	if err != nil {
		t.Fatalf("Failed to read llms.txt: %v", err)
	}
	content := string(llmsContent)

	if !strings.Contains(content, "# Test API") {
		t.Error("llms.txt missing title")
	}
	if !strings.Contains(content, "Test API description") {
		t.Error("llms.txt missing description")
	}
	if !strings.Contains(content, "https://api.test.com") {
		t.Error("llms.txt missing base URL")
	}
	if !strings.Contains(content, "## Authentication") {
		t.Error("llms.txt missing authentication section")
	}
	if !strings.Contains(content, "X-API-Key") {
		t.Error("llms.txt missing API key info")
	}
	if !strings.Contains(content, "[GET /users](./endpoints/get-users.txt)") {
		t.Error("llms.txt missing GET /users link")
	}
	if !strings.Contains(content, "[POST /users](./endpoints/post-users.txt)") {
		t.Error("llms.txt missing POST /users link")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "users"},
		{"User Operations", "user-operations"},
		{"api/v1", "api-v1"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateCurlExample(t *testing.T) {
	api := &parser.API{
		BaseURL: "https://api.example.com",
		SecuritySchemes: []parser.SecurityScheme{
			{Name: "apiKey", Type: "apiKey", In: "header", ParamName: "Authorization"},
		},
	}
	cfg := &config.Config{}
	gen := New(cfg, api)

	ep := parser.Endpoint{
		Method:  "GET",
		Path:    "/users/{id}",
		Summary: "Get user",
		Parameters: []parser.Parameter{
			{Name: "id", In: "path", Type: "integer"},
			{Name: "expand", In: "query", Type: "boolean"},
		},
	}

	result := gen.generateCurlExample(ep)

	if !strings.Contains(result, "curl -X GET") {
		t.Error("Missing curl command")
	}
	if !strings.Contains(result, "https://api.example.com/users/1") {
		t.Error("Missing URL with path parameter")
	}
	if !strings.Contains(result, "expand=true") {
		t.Error("Missing query parameter")
	}
	if !strings.Contains(result, "Authorization: YOUR_API_KEY") {
		t.Error("Missing auth header")
	}
}

func TestGenerateSchemaDoc(t *testing.T) {
	api := &parser.API{}
	cfg := &config.Config{}
	gen := New(cfg, api)

	schema := &parser.Schema{
		Type: "object",
		Properties: map[string]*parser.Schema{
			"name":  {Type: "string", Description: "User name"},
			"age":   {Type: "integer"},
			"email": {Type: "string", Format: "email"},
		},
	}

	result := gen.generateSchemaDoc(schema, 0)

	if !strings.Contains(result, "```json") {
		t.Error("Missing JSON code block")
	}
	if !strings.Contains(result, "\"name\"") {
		t.Error("Missing name field")
	}
	if !strings.Contains(result, "| Field | Type | Description |") {
		t.Error("Missing fields table")
	}
}
