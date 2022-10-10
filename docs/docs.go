// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/integration-system/isp-lib/v2/docs"

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "GNU GPL v3.0"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Авторизация с получением токена администратора",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Авторизация по логину и паролю",
                "parameters": [
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "401": {
                        "description": "Данные для авторизации не верны",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/auth/login_with_sudir": {
            "post": {
                "description": "Авторизация с получением токена администратора",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Авторизация по авторизационному коду от СУДИР",
                "parameters": [
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.LoginSudirRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.LoginResponse"
                        }
                    },
                    "401": {
                        "description": "Некорректный код для авторизации",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "412": {
                        "description": "Авторизация СУДИР не настроена на сервере",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/auth/logout": {
            "post": {
                "description": "Выход из авторизованной сессии администрирования",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Выход из авторизованной сессии",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    },
                    "400": {
                        "description": "Невалидный токен",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/secure/authenticate": {
            "post": {
                "description": "Проверяет токен и возвращает идентификатор администратора",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "secure"
                ],
                "summary": "Метод аутентификации токена",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.SecureAuthRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.SecureAuthResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/user/create_user": {
            "post": {
                "description": "Создать пользователя",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Создать пользователя",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.CreateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.User"
                        }
                    },
                    "400": {
                        "description": "Невалидное тело запроса",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "409": {
                        "description": "Пользователь с указанным email уже существует",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/user/delete_user": {
            "post": {
                "description": "Удалить существующего пользователя",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Удалить пользователя",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.IdentitiesRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.DeleteResponse"
                        }
                    },
                    "400": {
                        "description": "Невалидное тело запроса",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/user/get_design": {
            "post": {
                "description": "Получение внешнего вида (палитра и наименование) админ-интерфейса",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Получение внешнего вида",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/conf.UIDesign"
                        }
                    },
                    "400": {
                        "description": "Невалидный токен",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/user/get_profile": {
            "post": {
                "description": "Получить данные профиля",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Получить профиль",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.AdminUserShort"
                        }
                    },
                    "400": {
                        "description": "Невалидный токен",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "401": {
                        "description": "Токен не соответствует ни одному пользователю",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/user/get_users": {
            "post": {
                "description": "Получить список пользователей",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Список пользователей",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.UsersRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.UsersResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        },
        "/user/update_user": {
            "post": {
                "description": "Обновить данные существующего пользователя",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Обновить пользователя",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Токен администратора",
                        "name": "X-AUTH-ADMIN",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Тело запроса",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.UpdateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.User"
                        }
                    },
                    "400": {
                        "description": "Невалидное тело запроса",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "404": {
                        "description": "Пользователь с указанным id не существует",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "409": {
                        "description": "Пользователь с указанным email уже существует",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/domain.GrpcError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "conf.UIDesign": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "primaryColor": {
                    "type": "string"
                }
            }
        },
        "domain.AdminUserShort": {
            "type": "object",
            "required": [
                "email"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string"
                },
                "lastName": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                }
            }
        },
        "domain.CreateUserRequest": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string"
                },
                "lastName": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "roleId": {
                    "type": "integer"
                }
            }
        },
        "domain.DeleteResponse": {
            "type": "object",
            "properties": {
                "deleted": {
                    "type": "integer"
                }
            }
        },
        "domain.GrpcError": {
            "type": "object",
            "properties": {
                "details": {
                    "type": "array",
                    "items": {
                        "type": "object"
                    }
                },
                "errorCode": {
                    "type": "string"
                },
                "errorMessage": {
                    "type": "string"
                }
            }
        },
        "domain.IdentitiesRequest": {
            "type": "object",
            "required": [
                "ids"
            ],
            "properties": {
                "ids": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "domain.LoginRequest": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "domain.LoginResponse": {
            "type": "object",
            "properties": {
                "expired": {
                    "type": "string"
                },
                "headerName": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "domain.LoginSudirRequest": {
            "type": "object",
            "required": [
                "authCode"
            ],
            "properties": {
                "authCode": {
                    "type": "string"
                }
            }
        },
        "domain.SecureAuthRequest": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "domain.SecureAuthResponse": {
            "type": "object",
            "properties": {
                "adminId": {
                    "type": "integer"
                },
                "authenticated": {
                    "type": "boolean"
                },
                "errorReason": {
                    "type": "string"
                }
            }
        },
        "domain.UpdateUserRequest": {
            "type": "object",
            "required": [
                "id"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "lastName": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "roleId": {
                    "type": "integer"
                }
            }
        },
        "domain.User": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "lastName": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "roleId": {
                    "type": "integer"
                },
                "sudirUserId": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "domain.UsersRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "ids": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                }
            }
        },
        "domain.UsersResponse": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.User"
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = docs.SwaggerInfo{
	Version:     "1.0.0",
	Host:        "localhost:9000",
	BasePath:    "/api/admin",
	Schemes:     []string{},
	Title:       "msp-admin-service",
	Description: "сервис администрирования",
}

func init() {
	docs.InitSwagger(SwaggerInfo, doc)
}
