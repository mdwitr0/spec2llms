# spec2llms

CLI tool to generate [llms.txt](https://llmstxt.org/) files from OpenAPI 3.x specifications.

Creates LLM-friendly documentation for your API that AI agents can easily understand and use.

## Installation

### From source

```bash
go install github.com/mdwit/spec2llms/cmd/spec2llms@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/mdwit/spec2llms/releases).

## Usage

### Basic

```bash
# From local file
spec2llms ./openapi.json

# From URL
spec2llms https://petstore3.swagger.io/api/v3/openapi.json

# With options
spec2llms ./openapi.yaml -o ./docs -t "My API"
```

### Options

```
  -o, --output string     Output directory (default "./llms")
  -t, --title string      API title (overrides spec title)
  -b, --base-url string   Base URL for API (overrides spec servers)
  -c, --config string     Config file (spec2llms.json)
  -l, --lang string       Output language: en, ru (default "en")
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

## License

MIT
