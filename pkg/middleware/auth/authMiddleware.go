package middleware

import (
	mdlAuth "EMPC_BE/pkg/services/auth/model"
	scpAuth "EMPC_BE/pkg/services/auth/script"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/FDSAP-Git-Org/hephaestus/helper/v1"
	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	utils_v1 "github.com/FDSAP-Git-Org/hephaestus/utils/v1"
	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(c fiber.Ctx) error {
	// 1. Extract Authorization header

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return v1.JSONResponseWithError(
			c,
			respcode.ERR_CODE_401,
			"Authorization header missing",
			nil,
			http.StatusUnauthorized,
		)
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader || tokenString == "" {
		return v1.JSONResponseWithError(
			c,
			respcode.ERR_CODE_401,
			"Invalid token format",
			nil,
			http.StatusUnauthorized,
		)
	}

	// 2. Call Cagabay validate-header API
	apiURL := utils_v1.GetEnv("CAGABAY_BASE_URL") +
		"/soteria-go/api/public/v1/auth/security-management/validate-header"

	headers := map[string]string{
		"Authorization": "Bearer " + tokenString,
		"x-api-key":     utils_v1.GetEnv("CAGABAY_API_KEY"),
		"Content-Type":  "application/json",
	}

	resp, err := utils_v1.SendRequest(apiURL, "GET", nil, headers, 10)
	if err != nil {
		return v1.JSONResponseWithError(
			c,
			respcode.ERR_CODE_401,
			"Token validation failed",
			err,
			http.StatusUnauthorized,
		)
	}

	// 3. Parse Cagabay response
	var apiResp mdlAuth.ValidateTokenAPIResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return v1.JSONResponseWithError(
			c,
			respcode.ERR_CODE_310,
			"Failed to parse token validation response",
			err,
			http.StatusInternalServerError,
		)
	}

	// 4. Handle validation result
	switch apiResp.RetCode {
	case "215":
		// success → continue
	case "109":
		return v1.JSONResponseWithError(
			c,
			respcode.ERR_CODE_401,
			"Token has been terminated",
			nil,
			http.StatusUnauthorized,
		)
	default:
		return v1.JSONResponseWithError(
			c,
			respcode.ERR_CODE_401,
			"Token validation failed",
			nil,
			http.StatusUnauthorized,
		)
	}

	// 5. Store validated data in context
	if apiResp.Data != nil && apiResp.Data.Details != nil {
		c.Locals("username", apiResp.Data.Details.Username)
		c.Locals("institution_code", apiResp.Data.Details.InstiCode)
		c.Locals("institution_name", apiResp.Data.Details.InstiName)
		c.Locals("app_code", apiResp.Data.Details.AppCode)
		c.Locals("app_name", apiResp.Data.Details.AppName)

		// Fetch user with navigation using the new function
		userWithNav, err := scpAuth.GetUserWithNavigation(apiResp.Data.Details.Username)
		if err != nil {
			return v1.JSONResponseWithError(c,
				respcode.ERR_CODE_500,
				"Failed to fetch User with navigation",
				err,
				http.StatusInternalServerError,
			)
		}
		c.Locals("user", userWithNav)
		c.Locals("navigation", userWithNav.Navigation)

		fmt.Printf("User %s authenticated with role: %s\n",
			apiResp.Data.Details.Username, userWithNav.RoleName)
	}

	return c.Next()
}
