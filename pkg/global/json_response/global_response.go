package helper

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type (
	Response struct {
		ResponseTime string `json:"response_time"`
		Device       string `json:"device"`
		RetCode      string `json:"ret_code"`
		Message      string `json:"message"`
		Result       any    `json:"validation,omitempty"`
		Data         any    `json:"response,omitempty"`
		Error        any    `json:"error,omitempty"`
	}

	EPResponse struct {
		ProcessTime string `json:"process_time"`
		Response    any    `json:"response"`
	}
)

func JSONResponseV1(c fiber.Ctx, recCode, retMessage string, httpStatusCode int) error {
	return c.Status(httpStatusCode).JSON(Response{
		ResponseTime: time.Now().Format(time.DateTime),
		Device:       string(c.RequestCtx().UserAgent()),
		RetCode:      recCode,
		Message:      retMessage,
	})
}

func JSONResponseWithDataV1(c fiber.Ctx, recCode, retMessage string, data any, httpStatusCode int) error {
	return c.Status(httpStatusCode).JSON(Response{
		ResponseTime: time.Now().Format(time.DateTime),
		Device:       string(c.RequestCtx().UserAgent()),
		RetCode:      recCode,
		Message:      retMessage,
		Data:         data,
	})
}

func JSONResponseWithErrorV1(c fiber.Ctx, recCode, retMessage string, err error, httpStatusCode int) error {
	return c.Status(httpStatusCode).JSON(Response{
		ResponseTime: time.Now().Format(time.DateTime),
		Device:       string(c.RequestCtx().UserAgent()),
		RetCode:      recCode,
		Message:      retMessage,
		Error:        err.Error(),
	})
}
