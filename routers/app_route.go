package routers

import (
	global "EMPC_BE/pkg/global/json_response"
	loggerV1 "EMPC_BE/pkg/middleware/logger/v1"
	crtlAuth "EMPC_BE/pkg/services/auth/controller"
	crlDataEncryptionV1 "EMPC_BE/pkg/services/data_encryption/controller/v1"
	ctrRbac "EMPC_BE/pkg/services/rbac/controller"
	auth "EMPC_BE/pkg/middleware/auth"
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

	// UTILITY
	utility := apiV1.Group("/utility")
	utility.Post("/encrypt-data", crlDataEncryptionV1.EncrypDecryptV1)
	utility.Post("/decrypt-data", crlDataEncryptionV1.DecryptDataV1)

	// RBAC
	rbac := apiV1.Group("/rbac")
	rbac.Post("/assign-navigation-access", ctrRbac.AssignNavigationAccess)
	rbac.Get("/all-roles-navigation-access", ctrRbac.GetAllRolesNavigationAccess)
	rbac.Get("/role/:roleId/navigation-access", ctrRbac.GetRoleNavigationAccess)
	rbac.Delete("/remove-navigation-access", ctrRbac.RemoveNavigationAccess)

	//AUTH PUBLIC
	authPublic := apiV1.Group("/auth")
	authPublic.Post("/register", crtlAuth.RegisterUser)
	authPublic.Post("/login", crtlAuth.LoginUser)
	authPublic.Post("/forgot-password", crtlAuth.ForgotPassword)
	authPublic.Post("/verify-reset-token", crtlAuth.VerifyResetToken)
	authPublic.Post("/change-password", crtlAuth.ChangeTempPassword)

	//AUTH PRIVATE
	authProtected := apiV1.Group("/auth", auth.AuthMiddleware)
	authProtected.Post("/logout", crtlAuth.LogoutUser)
	authProtected.Post("/delete-user", crtlAuth.DeleteUser)
	authProtected.Post("/update-user/:username", crtlAuth.UpdateUser)
}
