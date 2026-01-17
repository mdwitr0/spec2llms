package parser

// API представляет распарсенную OpenAPI спецификацию
type API struct {
	Title           string
	Description     string
	Version         string
	BaseURL         string
	Tags            []Tag
	Endpoints       []Endpoint
	SecuritySchemes []SecurityScheme
}

// SecurityScheme представляет схему аутентификации
type SecurityScheme struct {
	Name        string
	Type        string // apiKey, http, oauth2, openIdConnect
	Description string
	In          string // header, query, cookie (для apiKey)
	ParamName   string // имя параметра (для apiKey)
	Scheme      string // bearer, basic (для http)
}

// Tag представляет группу эндпоинтов
type Tag struct {
	Name        string
	Description string
}

// Endpoint представляет один API эндпоинт
type Endpoint struct {
	Method      string // GET, POST, PUT, DELETE, PATCH
	Path        string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]Response
	Deprecated  bool
}

// Parameter представляет параметр запроса
type Parameter struct {
	Name        string
	In          string // query, path, header, cookie
	Description string
	Required    bool
	Type        string
	Format      string
	Enum        []string
	Default     any
	Example     any
}

// RequestBody представляет тело запроса
type RequestBody struct {
	Description string
	Required    bool
	Content     map[string]MediaType // application/json, etc.
}

// MediaType представляет тип контента
type MediaType struct {
	Schema  *Schema
	Example any
}

// Response представляет ответ API
type Response struct {
	Description string
	Content     map[string]MediaType
}

// Schema представляет JSON Schema
type Schema struct {
	Type        string
	Format      string
	Description string
	Properties  map[string]*Schema
	Items       *Schema // для массивов
	Required    []string
	Enum        []string
	Example     any
	Ref         string // ссылка на компонент
}
