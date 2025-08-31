package utils

import "github.com/gofiber/fiber/v2"

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
}

var ErrorMap2Msg map[int]string = map[int]string{
	400: "请求参数错误",
	401: "未授权",
	403: "禁止访问",
	404: "未找到资源",
	500: "服务器错误",
}

// SuccessWithData 返回包含数据的成功响应
func SuccessWithData(c *fiber.Ctx, data interface{}, message ...string) error {
	msg := "操作成功"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

// Success 返回成功响应
func Success(c *fiber.Ctx, message string) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Message: message,
	})
}

// Error 返回错误响应
func Error(c *fiber.Ctx, statusCode int, message ...string) error {
	msg := ErrorMap2Msg[statusCode]
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	response := ErrorResponse{
		Success: false,
		Message: msg,
	}

	return c.Status(statusCode).JSON(response)
}

// PaginationSuccess 返回分页成功响应
func PaginationSuccess(c *fiber.Ctx, message string, data interface{}, total int64, page, limit int) error {
	return c.JSON(PaginationResponse{
		Success: true,
		Message: message,
		Data:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}
