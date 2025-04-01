package docs

import (
	"github.com/gofiber/fiber/v2"
	"github.com/swaggo/swag"
)

// 下面的注释是自动生成的Swagger文档配置
// @title MiniDevOps API
// @version 1.0
// @description 迷你构建工具API文档
// @termsOfService http://example.com/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// API信息结构体
var SwaggerInfo = &swag.Spec{
	Version:     "1.0",
	Title:       "MiniDevOps API",
	Description: "迷你构建工具API文档",
	Host:        "localhost:8080",
	BasePath:    "/",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

// docTemplate 包含API文档的JSON模板
const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://example.com/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.example.com/support",
            "email": "support@example.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/login": {
            "post": {
                "description": "用户登录并获取JWT令牌",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "认证"
                ],
                "summary": "用户登录",
                "parameters": [
                    {
                        "description": "用户凭证",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UserCredentials"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/register": {
            "post": {
                "description": "注册新用户账号",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "认证"
                ],
                "summary": "注册新用户",
                "parameters": [
                    {
                        "description": "用户注册信息",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/projects": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "获取当前用户的所有项目",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "项目"
                ],
                "summary": "获取项目列表",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.ProjectResponse"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "创建新的构建项目",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "项目"
                ],
                "summary": "创建新项目",
                "parameters": [
                    {
                        "description": "项目信息",
                        "name": "project",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.ProjectRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.ProjectResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/projects/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "获取指定项目的详细信息",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "项目"
                ],
                "summary": "获取项目详情",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "项目ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ProjectResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "更新指定项目的信息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "项目"
                ],
                "summary": "更新项目",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "项目ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "项目更新信息",
                        "name": "project",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.ProjectRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ProjectResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "删除指定的项目",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "项目"
                ],
                "summary": "删除项目",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "项目ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/projects/{id}/build": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "为指定项目启动新的构建任务",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "构建"
                ],
                "summary": "启动构建任务",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "项目ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.BuildTaskResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/projects/{id}/builds": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "获取指定项目的所有构建任务历史",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "构建"
                ],
                "summary": "获取项目构建历史",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "项目ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.BuildTaskResponse"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/builds/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "获取指定构建任务的详细信息",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "构建"
                ],
                "summary": "获取构建任务详情",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "构建任务ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.BuildTaskResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/builds/{id}/logs": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "获取指定构建任务的日志内容",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "构建"
                ],
                "summary": "获取构建日志",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "构建任务ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/ws/builds/{id}/logs": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "通过WebSocket实时获取构建日志",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "构建"
                ],
                "summary": "构建日志WebSocket",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "构建任务ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "101": {
                        "description": "WebSocket升级成功",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handler.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handler.BuildTaskResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "duration": {
                    "type": "integer"
                },
                "finished_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "id": {
                    "type": "integer"
                },
                "log_path": {
                    "type": "string"
                },
                "project_id": {
                    "type": "integer"
                },
                "started_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "handler.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "handler.LoginResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/handler.UserProfile"
                }
            }
        },
        "handler.ProjectRequest": {
            "type": "object",
            "properties": {
                "branch": {
                    "type": "string"
                },
                "build_cmd": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "repo_url": {
                    "type": "string"
                }
            }
        },
        "handler.ProjectResponse": {
            "type": "object",
            "properties": {
                "branch": {
                    "type": "string"
                },
                "build_cmd": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "id": {
                    "type": "integer"
                },
                "last_build_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "name": {
                    "type": "string"
                },
                "repo_url": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "handler.RegisterRequest": {
            "type": "object",
            "properties": {
                "gitee_token": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "handler.SuccessResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "object"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "handler.UserCredentials": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "handler.UserProfile": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "id": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// 初始化swagger文档
func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}