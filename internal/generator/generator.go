package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mdwit/spec2llms/internal/config"
	"github.com/mdwit/spec2llms/internal/parser"
)

// Generator генерирует llms.txt файлы
type Generator struct {
	cfg *config.Config
	api *parser.API
}

// New создаёт новый генератор
func New(cfg *config.Config, api *parser.API) *Generator {
	return &Generator{cfg: cfg, api: api}
}

// Generate генерирует все файлы
func (g *Generator) Generate() error {
	// Создаём директории
	endpointsDir := filepath.Join(g.cfg.Output, "endpoints")
	if err := os.MkdirAll(endpointsDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Сортируем эндпоинты
	endpoints := g.sortEndpoints()

	// Генерируем файл для каждого эндпоинта
	for _, ep := range endpoints {
		filename := g.getEndpointFilename(ep)
		path := filepath.Join(endpointsDir, filename)
		content := g.generateSingleEndpointFile(ep)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	// Генерируем индексный файл llms.txt
	indexPath := filepath.Join(g.cfg.Output, "llms.txt")
	indexContent := g.generateIndex(endpoints)
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("failed to write llms.txt: %w", err)
	}

	return nil
}

// getEndpointFilename генерирует имя файла для endpoint'а
func (g *Generator) getEndpointFilename(ep parser.Endpoint) string {
	// GET /v1.4/person/search -> get-v1.4-person-search.txt
	path := strings.TrimPrefix(ep.Path, "/")
	path = strings.ReplaceAll(path, "/", "-")
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")
	return strings.ToLower(ep.Method) + "-" + path + ".txt"
}

// sortEndpoints сортирует эндпоинты по пути и методу
func (g *Generator) sortEndpoints() []parser.Endpoint {
	endpoints := make([]parser.Endpoint, len(g.api.Endpoints))
	copy(endpoints, g.api.Endpoints)

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path == endpoints[j].Path {
			return methodOrder(endpoints[i].Method) < methodOrder(endpoints[j].Method)
		}
		return endpoints[i].Path < endpoints[j].Path
	})

	return endpoints
}

// generateSingleEndpointFile генерирует содержимое файла для одного endpoint'а
func (g *Generator) generateSingleEndpointFile(ep parser.Endpoint) string {
	var sb strings.Builder

	// Заголовок с тегом если есть
	if len(ep.Tags) > 0 {
		sb.WriteString("# " + ep.Tags[0] + "\n\n")
	}

	sb.WriteString(g.generateEndpoint(ep))
	return sb.String()
}

func methodOrder(method string) int {
	order := map[string]int{"GET": 1, "POST": 2, "PUT": 3, "PATCH": 4, "DELETE": 5}
	if o, ok := order[method]; ok {
		return o
	}
	return 99
}

func (g *Generator) generateIndex(endpoints []parser.Endpoint) string {
	var sb strings.Builder

	// Заголовок
	title := g.cfg.Title
	if title == "" {
		title = g.api.Title
	}
	sb.WriteString("# " + title + "\n\n")

	// Описание
	if g.api.Description != "" {
		sb.WriteString("> " + g.api.Description + "\n\n")
	}

	// Базовый URL
	baseURL := g.cfg.BaseURL
	if baseURL == "" {
		baseURL = g.api.BaseURL
	}
	if baseURL != "" {
		sb.WriteString("Base URL: `" + baseURL + "`\n\n")
	}

	// Версия
	if g.api.Version != "" {
		sb.WriteString("Version: " + g.api.Version + "\n\n")
	}

	// Аутентификация
	if len(g.api.SecuritySchemes) > 0 {
		sb.WriteString("## Authentication\n\n")
		for _, scheme := range g.api.SecuritySchemes {
			sb.WriteString(g.formatSecurityScheme(scheme))
		}
		sb.WriteString("\n")
	}

	// Список эндпоинтов
	sb.WriteString("## Endpoints\n\n")

	// Формируем базовый путь для ссылок на документацию
	linksBase := "./endpoints"
	if g.cfg.DocsBaseURL != "" {
		linksBase = strings.TrimSuffix(g.cfg.DocsBaseURL, "/") + "/endpoints"
	}

	for _, ep := range endpoints {
		filename := g.getEndpointFilename(ep)
		summary := ep.Summary
		if summary == "" {
			summary = ep.Path
		}
		sb.WriteString(fmt.Sprintf("- [%s %s](%s/%s) — %s\n",
			ep.Method, ep.Path, linksBase, filename, summary))
	}

	return sb.String()
}

func (g *Generator) generateEndpoint(ep parser.Endpoint) string {
	var sb strings.Builder

	// Заголовок: METHOD /path - Summary
	header := fmt.Sprintf("## %s %s", ep.Method, ep.Path)
	if ep.Summary != "" {
		header += " - " + ep.Summary
	}
	if ep.Deprecated {
		header += " ⚠️ DEPRECATED"
	}
	sb.WriteString(header + "\n\n")

	// Описание
	if ep.Description != "" {
		sb.WriteString(ep.Description + "\n\n")
	}

	// Параметры
	if len(ep.Parameters) > 0 {
		sb.WriteString("### Parameters\n\n")
		sb.WriteString("| Name | In | Type | Required | Description |\n")
		sb.WriteString("|------|-----|------|----------|-------------|\n")

		for _, p := range ep.Parameters {
			required := ""
			if p.Required {
				required = "✓"
			}
			desc := p.Description
			if len(p.Enum) > 0 {
				desc += fmt.Sprintf(" Enum: `%s`", strings.Join(p.Enum, "`, `"))
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				p.Name, p.In, p.Type, required, desc))
		}
		sb.WriteString("\n")
	}

	// Request Body
	if ep.RequestBody != nil {
		sb.WriteString("### Request Body\n\n")
		if ep.RequestBody.Description != "" {
			sb.WriteString(ep.RequestBody.Description + "\n\n")
		}
		for contentType, media := range ep.RequestBody.Content {
			sb.WriteString("Content-Type: `" + contentType + "`\n\n")
			if media.Schema != nil {
				sb.WriteString(g.generateSchemaDoc(media.Schema, 0))
			}
		}
	}

	// Responses
	if len(ep.Responses) > 0 {
		sb.WriteString("### Responses\n\n")

		// Сортируем коды ответов
		codes := make([]string, 0, len(ep.Responses))
		for code := range ep.Responses {
			codes = append(codes, code)
		}
		sort.Strings(codes)

		for _, code := range codes {
			resp := ep.Responses[code]
			sb.WriteString(fmt.Sprintf("**%s** - %s\n\n", code, resp.Description))

			for contentType, media := range resp.Content {
				sb.WriteString("Content-Type: `" + contentType + "`\n\n")
				if media.Schema != nil {
					sb.WriteString(g.generateSchemaDoc(media.Schema, 0))
				}
			}
		}
	}

	// Пример curl
	sb.WriteString("### Example\n\n")
	sb.WriteString(g.generateCurlExample(ep))

	return sb.String()
}

// maxNestedDepth — максимальная глубина раскрытия вложенных объектов
const maxNestedDepth = 4

func (g *Generator) generateSchemaDoc(schema *parser.Schema, depth int) string {
	if schema == nil || depth > 4 {
		return ""
	}

	var sb strings.Builder

	if schema.Type == "object" && len(schema.Properties) > 0 {
		sb.WriteString("```json\n")
		sb.WriteString(g.renderJSONSchema(schema, 0, maxNestedDepth))
		sb.WriteString("```\n\n")

		// Добавляем описание полей в виде таблицы
		sb.WriteString(g.generateFieldsTable(schema, ""))
	} else if schema.Type == "array" && schema.Items != nil {
		itemType := schema.Items.Type
		if itemType == "" {
			itemType = "object"
		}
		sb.WriteString(fmt.Sprintf("Array of `%s`\n\n", itemType))
		if schema.Items.Type == "object" && len(schema.Items.Properties) > 0 {
			sb.WriteString(g.generateSchemaDoc(schema.Items, depth+1))
		}
	}

	return sb.String()
}

func (g *Generator) renderJSONSchema(schema *parser.Schema, indent, maxDepth int) string {
	if schema == nil || indent > maxDepth*2 {
		return ""
	}

	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	if schema.Type == "object" && len(schema.Properties) > 0 {
		sb.WriteString("{\n")

		props := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			props = append(props, name)
		}
		sort.Strings(props)

		for i, name := range props {
			prop := schema.Properties[name]
			comma := ","
			if i == len(props)-1 {
				comma = ""
			}

			sb.WriteString(prefix + "  \"" + name + "\": ")
			value := g.renderPropertyValue(prop, indent+1, maxDepth)
			if value == "" {
				// Fallback для пустых значений
				if prop.Type == "array" {
					value = "[{}]"
				} else if prop.Type == "object" {
					value = "{}"
				} else {
					value = "null"
				}
			}
			sb.WriteString(value)
			sb.WriteString(comma + "\n")
		}

		sb.WriteString(prefix + "}")
	} else if schema.Type == "array" {
		if schema.Items != nil && schema.Items.Type == "object" && len(schema.Items.Properties) > 0 {
			sb.WriteString("[\n" + prefix + "  ")
			sb.WriteString(g.renderJSONSchema(schema.Items, indent+1, maxDepth))
			sb.WriteString("\n" + prefix + "]")
		} else if schema.Items != nil {
			sb.WriteString("[" + g.getTypeExample(schema.Items) + "]")
		} else {
			sb.WriteString("[]")
		}
	} else if schema.Type == "object" {
		sb.WriteString("{}")
	} else {
		sb.WriteString(g.getTypeExample(schema))
	}

	return sb.String()
}

func (g *Generator) renderPropertyValue(prop *parser.Schema, indent, maxDepth int) string {
	if prop == nil {
		return "null"
	}

	// Если есть пример - используем его
	if prop.Example != nil {
		return g.formatExample(prop.Example)
	}

	// Для объектов рекурсивно разворачиваем
	if prop.Type == "object" && len(prop.Properties) > 0 && indent < maxDepth*2 {
		return g.renderJSONSchema(prop, indent, maxDepth)
	}

	// Для массивов
	if prop.Type == "array" {
		if prop.Items != nil {
			// Объект с properties - разворачиваем
			if prop.Items.Type == "object" && len(prop.Items.Properties) > 0 {
				return g.renderJSONSchema(prop, indent, maxDepth)
			}
			// Объект без properties или другой тип
			example := g.getTypeExample(prop.Items)
			if example == "" || example == "null" {
				example = "{}"
			}
			return "[" + example + "]"
		}
		return "[{}]"
	}

	result := g.getTypeExample(prop)
	if result == "" {
		return "null"
	}
	return result
}

func (g *Generator) getTypeExample(schema *parser.Schema) string {
	if schema == nil {
		return "null"
	}

	// Если есть пример - используем его
	if schema.Example != nil {
		return g.formatExample(schema.Example)
	}

	// Если есть enum - показываем первое значение
	if len(schema.Enum) > 0 {
		return fmt.Sprintf("\"%s\"", schema.Enum[0])
	}

	switch schema.Type {
	case "string":
		if schema.Format == "date-time" {
			return "\"2024-01-15T10:00:00Z\""
		}
		if schema.Format == "date" {
			return "\"2024-01-15\""
		}
		if schema.Format == "email" {
			return "\"user@example.com\""
		}
		if schema.Format == "uri" || schema.Format == "url" {
			return "\"https://example.com\""
		}
		return "\"string\""
	case "integer":
		return "0"
	case "number":
		return "0.0"
	case "boolean":
		return "true"
	case "array":
		if schema.Items != nil {
			return "[" + g.getTypeExample(schema.Items) + "]"
		}
		return "[]"
	case "object":
		return "{}"
	case "":
		// Тип не указан - возвращаем object
		return "{}"
	default:
		return "null"
	}
}

func (g *Generator) formatExample(example any) string {
	switch v := example.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case float64, float32, int, int64, int32:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (g *Generator) generateFieldsTable(schema *parser.Schema, prefix string) string {
	if schema == nil || len(schema.Properties) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("| Field | Type | Description |\n")
	sb.WriteString("|-------|------|-------------|\n")

	props := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		props = append(props, name)
	}
	sort.Strings(props)

	for _, name := range props {
		prop := schema.Properties[name]
		fieldName := name
		if prefix != "" {
			fieldName = prefix + "." + name
		}

		typeStr := prop.Type
		if prop.Format != "" {
			typeStr += " (" + prop.Format + ")"
		}
		if prop.Type == "array" && prop.Items != nil {
			typeStr = "array[" + prop.Items.Type + "]"
		}

		desc := prop.Description
		if len(prop.Enum) > 0 {
			desc += " Values: `" + strings.Join(prop.Enum, "`, `") + "`"
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", fieldName, typeStr, desc))
	}

	sb.WriteString("\n")
	return sb.String()
}

func sanitizeFilename(name string) string {
	// Заменяем пробелы и спецсимволы на дефисы
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}

func (g *Generator) generateCurlExample(ep parser.Endpoint) string {
	var sb strings.Builder

	baseURL := g.cfg.BaseURL
	if baseURL == "" {
		baseURL = g.api.BaseURL
	}
	if baseURL == "" || strings.HasPrefix(baseURL, "/") {
		baseURL = "https://api.example.com" + baseURL
	}

	// Убираем trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Формируем путь с примерами параметров
	path := ep.Path
	for _, p := range ep.Parameters {
		if p.In == "path" {
			var example string
			if p.Example != nil {
				example = fmt.Sprintf("%v", p.Example)
			} else if p.Type == "integer" {
				example = "1"
			} else {
				example = "example"
			}
			path = strings.ReplaceAll(path, "{"+p.Name+"}", example)
		}
	}

	// Query параметры
	var queryParams []string
	for _, p := range ep.Parameters {
		if p.In == "query" {
			example := ""
			if p.Example != nil {
				example = fmt.Sprintf("%v", p.Example)
			} else if len(p.Enum) > 0 {
				example = p.Enum[0]
			} else if p.Type == "integer" || p.Type == "number" {
				example = "1"
			} else if p.Type == "boolean" {
				example = "true"
			} else {
				example = "value"
			}
			queryParams = append(queryParams, p.Name+"="+example)
		}
	}

	url := baseURL + path
	if len(queryParams) > 0 {
		url += "?" + strings.Join(queryParams, "&")
	}

	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X %s \"%s\"", ep.Method, url))

	// Headers
	sb.WriteString(" \\\n  -H \"Content-Type: application/json\"")

	// Auth header (если есть security schemes)
	if len(g.api.SecuritySchemes) > 0 {
		for _, scheme := range g.api.SecuritySchemes {
			if scheme.Type == "apiKey" && scheme.In == "header" {
				sb.WriteString(fmt.Sprintf(" \\\n  -H \"%s: YOUR_API_KEY\"", scheme.ParamName))
				break
			} else if scheme.Type == "http" && scheme.Scheme == "bearer" {
				sb.WriteString(" \\\n  -H \"Authorization: Bearer YOUR_TOKEN\"")
				break
			}
		}
	}

	// Request body
	if ep.RequestBody != nil && (ep.Method == "POST" || ep.Method == "PUT" || ep.Method == "PATCH") {
		for _, media := range ep.RequestBody.Content {
			if media.Schema != nil {
				body := g.renderJSONSchema(media.Schema, 0, maxNestedDepth)
				if body != "" {
					sb.WriteString(" \\\n  -d '" + body + "'")
				}
			}
			break // Берём только первый content type
		}
	}

	sb.WriteString("\n```\n\n")
	return sb.String()
}

func (g *Generator) formatSecurityScheme(scheme parser.SecurityScheme) string {
	var sb strings.Builder

	sb.WriteString("### " + scheme.Name + "\n\n")

	if scheme.Description != "" {
		sb.WriteString(scheme.Description + "\n\n")
	}

	switch scheme.Type {
	case "apiKey":
		sb.WriteString("- **Type**: API Key\n")
		sb.WriteString(fmt.Sprintf("- **Parameter**: `%s`\n", scheme.ParamName))
		sb.WriteString(fmt.Sprintf("- **In**: %s\n", scheme.In))
	case "http":
		sb.WriteString(fmt.Sprintf("- **Type**: HTTP %s\n", scheme.Scheme))
		if scheme.Scheme == "bearer" {
			sb.WriteString("- **Header**: `Authorization: Bearer <token>`\n")
		} else if scheme.Scheme == "basic" {
			sb.WriteString("- **Header**: `Authorization: Basic <credentials>`\n")
		}
	case "oauth2":
		sb.WriteString("- **Type**: OAuth 2.0\n")
	case "openIdConnect":
		sb.WriteString("- **Type**: OpenID Connect\n")
	}

	sb.WriteString("\n")
	return sb.String()
}
