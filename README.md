# spec2llms

[![CI](https://github.com/mdwitr0/spec2llms/actions/workflows/ci.yml/badge.svg)](https://github.com/mdwitr0/spec2llms/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mdwitr0/spec2llms)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Generate [llms.txt](https://llmstxt.org/) from OpenAPI/Swagger specifications.**

Transform your OpenAPI 3.x specs into LLM-friendly documentation that AI agents (ChatGPT, Claude, Cursor, Copilot, etc.) can easily understand and use to interact with your API.

## Why llms.txt?

- **AI-Native Documentation** — Optimized format for LLM consumption
- **Better AI Integrations** — Help AI agents understand your API structure
- **Auto-Generated** — No manual documentation writing
- **Always Up-to-Date** — Generate from your OpenAPI spec on every build

## Features

- Parse OpenAPI 3.0/3.1 specifications (JSON & YAML)
- Load specs from local files or remote URLs
- Group endpoints by tags with clean file naming
- Generate curl examples with authentication
- Include request/response schemas
- Support for `--skip-validation` for specs with minor issues
- Multi-language support (English, Russian)

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap mdwitr0/tap
brew install spec2llms
```

### From source

```bash
go install github.com/mdwitr0/spec2llms/cmd/spec2llms@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/mdwitr0/spec2llms/releases).

## Usage

### Basic

```bash
# From local file
spec2llms ./openapi.json

# From URL (use quotes)
spec2llms "https://petstore3.swagger.io/api/v3/openapi.json"

# With options
spec2llms ./openapi.yaml -o ./docs -t "My API"

# Skip validation for specs with minor issues
spec2llms "https://api.example.com/openapi.json" --skip-validation
```

### Options

```
  -o, --output string     Output directory (default "./llms")
  -t, --title string      API title (overrides spec title)
  -b, --base-url string   Base URL for API (overrides spec servers)
  -c, --config string     Config file (spec2llms.json)
  -l, --lang string       Output language: en, ru (default "en")
      --skip-validation   Skip OpenAPI spec validation
  -v, --version           Print version
  -h, --help              Help
```

### Config file

Create `spec2llms.json`:

```json
{
  "source": "./openapi.json",
  "output": "./docs/llms",
  "baseUrl": "https://api.example.com",
  "title": "My API",
  "language": "en"
}
```

Run with config:

```bash
spec2llms -c spec2llms.json
```

## Output

```
llms/
├── llms.txt              # Index with links
└── endpoints/
    ├── users.txt         # Endpoints grouped by tag
    ├── orders.txt
    └── products.txt
```

### Example llms.txt

```markdown
# My API

> REST API for managing users and orders

Base URL: `https://api.example.com`

Version: 1.0.0

## Authentication

### apiKey

- **Type**: API Key
- **Parameter**: `X-API-Key`
- **In**: header

## Endpoints

- [users](./endpoints/users.txt) — User operations (5 endpoints)
- [orders](./endpoints/orders.txt) — Order management (3 endpoints)
```

### Example endpoint file

```markdown
# users

> User operations

## GET /users - List users

Get paginated list of users.

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| limit | query | integer |  | Max results |
| offset | query | integer |  | Skip first N |

### Responses

**200** - Success

Content-Type: `application/json`

```json
{
  "data": [
    {
      "id": 0,
      "name": "string",
      "email": "user@example.com"
    }
  ],
  "total": 0
}
```

### Example

```bash
curl -X GET "https://api.example.com/users?limit=10" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY"
```
```

## Supported Formats

- OpenAPI 3.x (JSON, YAML)
- Local files and URLs

## Development

```bash
# Build
go build -o spec2llms ./cmd/spec2llms

# Test
go test ./...

# Run locally
./spec2llms ./examples/petstore.json
```

## Use Cases

- **AI Agent Integration** — Provide context about your API to AI coding assistants
- **Documentation for LLMs** — Create machine-readable API docs for ChatGPT plugins
- **API Discovery** — Help AI tools understand available endpoints
- **CI/CD Pipeline** — Auto-generate llms.txt on every release

## Related

- [llms.txt specification](https://llmstxt.org/) — Standard for LLM-friendly documentation
- [OpenAPI Specification](https://www.openapis.org/) — API description format

## License

MIT

---

**Keywords:** llms.txt, OpenAPI, Swagger, API documentation, LLM, AI agents, ChatGPT, Claude, Cursor, Copilot, API spec, openapi-to-llms, swagger-to-llms, AI-friendly docs, machine-readable API, llmstxt generator
