package routers

import (
	global "golang-template-v3.1/pkg/global/json_response"
	loggerV1 "golang-template-v3.1/pkg/middleware/logger/v1"
	crlDataEncryptionV1 "golang-template-v3.1/pkg/services/data_encryption/controller/v1"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func AppRoutes(app *fiber.App) {

	app.Get("/", func(c fiber.Ctx) error {
		loggerV1.SystemLogger("API Health Check", "HealthCheck", "api_health", "HealthCheck", "Success", nil, "API is running...")
		return global.JSONResponseV1(c, "200", "API is running...", http.StatusOK)
	})

	apiV1 := app.Group("/api/v1")
	apiV1.Get("/", func(c fiber.Ctx) error {
		loggerV1.SystemLogger("API V1 Health Check", "system", "api_v1_health", "HealthCheck", "Success", nil, "API version 1 is running...")
		return global.JSONResponseV1(c, "200", "API version 1 is running...", http.StatusOK)
	})

	// TOOLS GROUP
	// FOR
	utility := apiV1.Group("/utility")
	utility.Post("/encryp-data", crlDataEncryptionV1.EncrypDecryptV1)
	utility.Post("/decrypt-data", crlDataEncryptionV1.DecryptDataV1)

}
