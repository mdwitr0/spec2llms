package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJSON(t *testing.T) {
	// Создаём временный JSON файл
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"description": "Test description",
			"version": "1.0.0"
		},
		"servers": [{"url": "https://api.example.com"}],
		"tags": [
			{"name": "users", "description": "User operations"}
		],
		"paths": {
			"/users": {
				"get": {
					"tags": ["users"],
					"summary": "List users",
					"parameters": [
						{
							"name": "limit",
							"in": "query",
							"schema": {"type": "integer"}
						}
					],
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {
										"type": "array",
										"items": {"type": "object"}
									}
								}
							}
						}
					}
				},
				"post": {
					"tags": ["users"],
					"summary": "Create user",
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"name": {"type": "string"},
										"email": {"type": "string", "format": "email"}
									}
								}
							}
						}
					},
					"responses": {
						"201": {"description": "Created"}
					}
				}
			},
			"/users/{id}": {
				"get": {
					"tags": ["users"],
					"summary": "Get user by ID",
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": {"type": "integer"}
						}
					],
					"responses": {
						"200": {"description": "Success"}
					}
				}
			}
		},
		"components": {
			"securitySchemes": {
				"bearerAuth": {
					"type": "http",
					"scheme": "bearer"
				},
				"apiKey": {
					"type": "apiKey",
					"in": "header",
					"name": "X-API-Key"
				}
			}
		}
	}`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "openapi.json")
	if err := os.WriteFile(tmpFile, []byte(spec), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	api, err := Parse(tmpFile, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Проверяем базовые поля
	if api.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", api.Title)
	}
	if api.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", api.Description)
	}
	if api.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", api.Version)
	}
	if api.BaseURL != "https://api.example.com" {
		t.Errorf("Expected baseURL 'https://api.example.com', got '%s'", api.BaseURL)
	}

	// Проверяем теги
	if len(api.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(api.Tags))
	}
	if api.Tags[0].Name != "users" {
		t.Errorf("Expected tag name 'users', got '%s'", api.Tags[0].Name)
	}

	// Проверяем эндпоинты
	if len(api.Endpoints) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(api.Endpoints))
	}

	// Проверяем security schemes
	if len(api.SecuritySchemes) != 2 {
		t.Errorf("Expected 2 security schemes, got %d", len(api.SecuritySchemes))
	}

	// Находим GET /users
	var getUsersEndpoint *Endpoint
	for i := range api.Endpoints {
		if api.Endpoints[i].Path == "/users" && api.Endpoints[i].Method == "GET" {
			getUsersEndpoint = &api.Endpoints[i]
			break
		}
	}
	if getUsersEndpoint == nil {
		t.Fatal("GET /users endpoint not found")
	}
	if getUsersEndpoint.Summary != "List users" {
		t.Errorf("Expected summary 'List users', got '%s'", getUsersEndpoint.Summary)
	}
	if len(getUsersEndpoint.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(getUsersEndpoint.Parameters))
	}
	if getUsersEndpoint.Parameters[0].Name != "limit" {
		t.Errorf("Expected parameter name 'limit', got '%s'", getUsersEndpoint.Parameters[0].Name)
	}
}

func TestParseYAML(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: YAML Test API
  version: "2.0.0"
paths:
  /health:
    get:
      summary: Health check
      responses:
        "200":
          description: OK
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "openapi.yaml")
	if err := os.WriteFile(tmpFile, []byte(spec), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	api, err := Parse(tmpFile, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if api.Title != "YAML Test API" {
		t.Errorf("Expected title 'YAML Test API', got '%s'", api.Title)
	}
	if len(api.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(api.Endpoints))
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/api.json", true},
		{"http://localhost:8080/spec.yaml", true},
		{"./openapi.json", false},
		{"/path/to/spec.json", false},
		{"file.json", false},
	}

	for _, tt := range tests {
		result := isURL(tt.input)
		if result != tt.expected {
			t.Errorf("isURL(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
