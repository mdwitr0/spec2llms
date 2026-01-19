package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// ParseOptions опции парсинга
type ParseOptions struct {
	SkipValidation bool
}

// Parse парсит OpenAPI спецификацию из файла или URL
func Parse(source string, opts *ParseOptions) (*API, error) {
	if opts == nil {
		opts = &ParseOptions{}
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var doc *openapi3.T
	var err error

	if isURL(source) {
		doc, err = loadFromURL(loader, source)
	} else {
		doc, err = loader.LoadFromFile(source)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	if !opts.SkipValidation {
		if err := doc.Validate(context.Background()); err != nil {
			return nil, fmt.Errorf("invalid OpenAPI spec: %w\n\nUse --skip-validation to ignore validation errors", err)
		}
	}

	return convertToAPI(doc), nil
}

func loadFromURL(loader *openapi3.Loader, rawURL string) (*openapi3.T, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Скачиваем файл
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Определяем формат по расширению или Content-Type
	isYAML := strings.HasSuffix(u.Path, ".yaml") ||
		strings.HasSuffix(u.Path, ".yml") ||
		strings.Contains(resp.Header.Get("Content-Type"), "yaml")

	// Создаём временный файл
	ext := ".json"
	if isYAML {
		ext = ".yaml"
	}
	tmpFile, err := os.CreateTemp("", "openapi-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	return loader.LoadFromFile(tmpPath)
}

// ParseFile парсит OpenAPI спецификацию из локального файла (JSON или YAML)
func ParseFile(path string) (*API, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Проверяем расширение
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("unsupported file format: %s (expected .json, .yaml, or .yml)", ext)
	}

	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	return convertToAPI(doc), nil
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func convertToAPI(doc *openapi3.T) *API {
	api := &API{
		Title:       doc.Info.Title,
		Description: doc.Info.Description,
		Version:     doc.Info.Version,
	}

	// Извлекаем базовый URL из серверов
	if len(doc.Servers) > 0 {
		api.BaseURL = doc.Servers[0].URL
	}

	// Конвертируем теги
	for _, tag := range doc.Tags {
		api.Tags = append(api.Tags, Tag{
			Name:        tag.Name,
			Description: tag.Description,
		})
	}

	// Конвертируем эндпоинты
	for path, pathItem := range doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			if op == nil {
				continue
			}
			endpoint := convertOperation(path, method, op)
			api.Endpoints = append(api.Endpoints, endpoint)
		}
	}

	// Конвертируем security schemes
	if doc.Components != nil && doc.Components.SecuritySchemes != nil {
		for name, schemeRef := range doc.Components.SecuritySchemes {
			if schemeRef.Value == nil {
				continue
			}
			scheme := schemeRef.Value
			ss := SecurityScheme{
				Name:        name,
				Type:        scheme.Type,
				Description: scheme.Description,
				In:          scheme.In,
				ParamName:   scheme.Name,
				Scheme:      scheme.Scheme,
			}
			api.SecuritySchemes = append(api.SecuritySchemes, ss)
		}
	}

	return api
}

func convertOperation(path, method string, op *openapi3.Operation) Endpoint {
	endpoint := Endpoint{
		Method:      method,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Deprecated:  op.Deprecated,
		Responses:   make(map[string]Response),
	}

	// Конвертируем параметры
	for _, paramRef := range op.Parameters {
		if paramRef.Value == nil {
			continue
		}
		param := convertParameter(paramRef.Value)
		endpoint.Parameters = append(endpoint.Parameters, param)
	}

	// Конвертируем тело запроса
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		endpoint.RequestBody = convertRequestBody(op.RequestBody.Value)
	}

	// Конвертируем ответы
	if op.Responses != nil {
		for code, responseRef := range op.Responses.Map() {
			if responseRef.Value == nil {
				continue
			}
			endpoint.Responses[code] = convertResponse(responseRef.Value)
		}
	}

	return endpoint
}

func convertParameter(p *openapi3.Parameter) Parameter {
	param := Parameter{
		Name:        p.Name,
		In:          p.In,
		Description: p.Description,
		Required:    p.Required,
	}

	if p.Schema != nil && p.Schema.Value != nil {
		schema := p.Schema.Value
		param.Type = schema.Type.Slice()[0]
		param.Format = schema.Format
		param.Default = schema.Default
		param.Example = schema.Example

		for _, e := range schema.Enum {
			if s, ok := e.(string); ok {
				param.Enum = append(param.Enum, s)
			}
		}
	}

	return param
}

func convertRequestBody(rb *openapi3.RequestBody) *RequestBody {
	reqBody := &RequestBody{
		Description: rb.Description,
		Required:    rb.Required,
		Content:     make(map[string]MediaType),
	}

	for contentType, mediaType := range rb.Content {
		mt := MediaType{
			Example: mediaType.Example,
		}
		if mediaType.Schema != nil && mediaType.Schema.Value != nil {
			mt.Schema = convertSchema(mediaType.Schema.Value)
		}
		reqBody.Content[contentType] = mt
	}

	return reqBody
}

func convertResponse(r *openapi3.Response) Response {
	resp := Response{
		Content: make(map[string]MediaType),
	}

	if r.Description != nil {
		resp.Description = *r.Description
	}

	for contentType, mediaType := range r.Content {
		mt := MediaType{
			Example: mediaType.Example,
		}
		if mediaType.Schema != nil && mediaType.Schema.Value != nil {
			mt.Schema = convertSchema(mediaType.Schema.Value)
		}
		resp.Content[contentType] = mt
	}

	return resp
}

func convertSchema(s *openapi3.Schema) *Schema {
	if s == nil {
		return nil
	}

	schema := &Schema{
		Format:      s.Format,
		Description: s.Description,
		Required:    s.Required,
		Example:     s.Example,
	}

	if len(s.Type.Slice()) > 0 {
		schema.Type = s.Type.Slice()[0]
	}

	// Конвертируем enum
	for _, e := range s.Enum {
		if str, ok := e.(string); ok {
			schema.Enum = append(schema.Enum, str)
		}
	}

	// Конвертируем properties для объектов
	if s.Properties != nil {
		schema.Properties = make(map[string]*Schema)
		for name, propRef := range s.Properties {
			if propRef.Value != nil {
				schema.Properties[name] = convertSchema(propRef.Value)
			}
		}
	}

	// Обрабатываем allOf — собираем все properties из всех схем
	if len(s.AllOf) > 0 {
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}
		for _, ref := range s.AllOf {
			if ref.Value != nil {
				merged := convertSchema(ref.Value)
				if merged != nil {
					// Копируем тип если не задан
					if schema.Type == "" && merged.Type != "" {
						schema.Type = merged.Type
					}
					// Копируем properties
					for name, prop := range merged.Properties {
						schema.Properties[name] = prop
					}
					// Копируем items если это массив
					if merged.Items != nil && schema.Items == nil {
						schema.Items = merged.Items
					}
				}
			}
		}
	}

	// Обрабатываем oneOf/anyOf — берём первую схему как пример
	if len(s.OneOf) > 0 && len(schema.Properties) == 0 {
		if s.OneOf[0].Value != nil {
			first := convertSchema(s.OneOf[0].Value)
			if first != nil {
				schema.Type = first.Type
				schema.Properties = first.Properties
				schema.Items = first.Items
			}
		}
	}
	if len(s.AnyOf) > 0 && len(schema.Properties) == 0 {
		if s.AnyOf[0].Value != nil {
			first := convertSchema(s.AnyOf[0].Value)
			if first != nil {
				schema.Type = first.Type
				schema.Properties = first.Properties
				schema.Items = first.Items
			}
		}
	}

	// Конвертируем items для массивов
	if s.Items != nil && s.Items.Value != nil {
		schema.Items = convertSchema(s.Items.Value)
	}

	return schema
}
