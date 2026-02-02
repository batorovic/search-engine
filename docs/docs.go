package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "email": "support@example.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/search": {
            "post": {
                "description": "Search for content across all providers with filtering and sorting",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["search"],
                "summary": "Search content",
                "parameters": [
                    {
                        "description": "Search request body",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "query": {
                                    "type": "string",
                                    "description": "Search query"
                                },
                                "types": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    },
                                    "description": "Content type filters (e.g., ['video', 'text'])"
                                },
                                "tags": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    },
                                    "description": "Tag filters"
                                },
                                "orderBy": {
                                    "type": "string",
                                    "default": "relevant_score",
                                    "description": "Sort field (popularity or relevant_score)"
                                },
                                "page": {
                                    "type": "integer",
                                    "default": 1,
                                    "description": "Page number"
                                },
                                "perPage": {
                                    "type": "integer",
                                    "default": 20,
                                    "description": "Items per page (max 100)"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful search response",
                        "schema": {
                            "type": "object"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object"
                        }
                    },
                    "429": {
                        "description": "Rate limit exceeded",
                        "schema": {
                            "type": "object"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        },
        "/health": {
            "get": {
                "description": "Check if the service is alive",
                "produces": ["application/json"],
                "tags": ["health"],
                "summary": "Liveness probe",
                "responses": {
                    "200": {
                        "description": "Service is healthy",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        },
        "/health/ready": {
            "get": {
                "description": "Check if the service is ready to accept traffic",
                "produces": ["application/json"],
                "tags": ["health"],
                "summary": "Readiness probe",
                "responses": {
                    "200": {
                        "description": "Service is ready",
                        "schema": {
                            "type": "object"
                        }
                    },
                    "503": {
                        "description": "Service is not ready",
                        "schema": {
                            "type": "object"
                        }
                    }
                }
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "Search Engine API",
	Description:      "A content search engine API that aggregates data from multiple providers",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
