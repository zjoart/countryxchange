// Package docs provides API documentation
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "This service provides country information with currency exchange rates and estimated GDP.",
        "title": "Country Xchange API",
        "contact": {},
        "version": "1.0"
    },
    "host": "{{.Host}}",
    "basePath": "/",
    "paths": {
        "/countries/refresh": {
            "post": {
                "description": "Fetches fresh country data and exchange rates from external APIs",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["countries"],
                "summary": "Refresh country data",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "example": "refreshed"
                                },
                                "total": {
                                    "type": "integer",
                                    "example": 250
                                },
                                "last_refreshed_at": {
                                    "type": "string",
                                    "example": "2025-10-26T14:30:00Z"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    }
                }
            }
        },
        "/countries": {
            "get": {
                "description": "Get all countries with optional filtering by region and currency",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["countries"],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Filter by region",
                        "name": "region",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by currency code",
                        "name": "currency",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Sort by GDP (gdp_asc or gdp_desc)",
                        "name": "sort",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Country"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    }
                }
            }
        },
        "/countries/image": {
            "get": {
                "description": "Get a PNG image summarizing country data",
                "produces": ["image/png"],
                "tags": ["countries"],
                "responses": {
                    "200": {
                        "description": "PNG image",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    }
                }
            }
        },
        "/countries/{name}": {
            "get": {
                "description": "Get detailed information about a specific country",
                "produces": ["application/json"],
                "tags": ["countries"],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Country name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {"$ref": "#/definitions/Country"}
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    }
                }
            },
            "delete": {
                "description": "Delete a country from the database",
                "produces": ["application/json"],
                "tags": ["countries"],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Country name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "example": "deleted"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    }
                }
            }
        },
        "/status": {
            "get": {
                "description": "Get the current status of the service",
                "produces": ["application/json"],
                "tags": ["status"],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {"$ref": "#/definitions/StatusResponse"}
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {"$ref": "#/definitions/ErrorResponse"}
                    }
                }
            }
        }
    },
    "definitions": {
        "Country": {
            "type": "object",
            "properties": {
                "id": {"type": "integer", "example": 1},
                "name": {"type": "string", "example": "United States"},
                "capital": {"type": "string", "example": "Washington, D.C."},
                "region": {"type": "string", "example": "Americas"},
                "population": {"type": "integer", "example": 331002651},
                "currency_code": {"type": "string", "example": "USD"},
                "exchange_rate": {"type": "number", "example": 1.0},
                "estimated_gdp": {"type": "number", "example": 21433225.0},
                "flag_url": {"type": "string", "example": "https://example.com/us-flag.png"},
                "last_refreshed_at": {"type": "string", "example": "2025-10-26T14:30:00Z"}
            }
        },
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {"type": "string"},
                "details": {"type": "object"}
            }
        },
        "StatusResponse": {
            "type": "object",
            "properties": {
                "total_countries": {"type": "integer", "example": 250},
                "last_refreshed_at": {"type": "string", "example": "2025-10-26T14:30:00Z"}
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "", // This will be set from environment
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "Country Xchange API",
	Description:      "This service provides country information with currency exchange rates and estimated GDP.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
