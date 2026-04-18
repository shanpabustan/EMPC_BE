package crtlAuth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	helper "EMPC_BE/pkg/global/json_response"
	httpRequestV1 "EMPC_BE/pkg/middleware/httpRequest/v1"
	utilityV1 "EMPC_BE/pkg/middleware/utility/v1"
	hlpAuth "EMPC_BE/pkg/services/auth/helper"
	mdlAuth "EMPC_BE/pkg/services/auth/model"
	scpAuth "EMPC_BE/pkg/services/auth/script"

	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	"github.com/gofiber/fiber/v3"
)

// ============================================
// STAFF REGISTRATION ENDPOINT
// ============================================
func RegisterUser(c fiber.Ctx) error {
	var req mdlAuth.RegisterStaffRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// Build API request with defaults
	apiReq := mdlAuth.StaffRegistrationApiRequest{
		StaffID:         req.StaffID,
		InstitutionCode: req.InstitutionCode,
		Birthdate:       req.Birthdate,
		Username:        "",
		FirstName:       "first_name",
		MiddleName:      "middle_name",
		LastName:        "last_name",
		PhoneNo:         "09123456789",
		Email:           "email@gmail.com",
	}

	// Call external staff registration endpoint
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/helper/auth/user-management/register-new-user/staff"
	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, _ := json.Marshal(apiReq)
	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Request to external API failed", err, http.StatusInternalServerError)
	}

	// Unmarshal to typed struct
	var apiResp mdlAuth.StaffRegistrationAPIResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}

	if apiResp.RetCode != "203" {
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode,
			apiResp.Data.Message, nil, http.StatusBadRequest)
	}

	// Hash password
	// apiResp.Data.Details.Password, err = utilityV1.HashData(apiResp.Data.Details.Password)
	// if err != nil {
	// 	return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
	// 		"Failed to Hash Password", err, http.StatusInternalServerError)
	// }

	// Success: save to internal DB
	result, err := scpAuth.RegisterUser(apiResp.Data.Details)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303,
			"Inserting data failed", err, http.StatusInternalServerError)
	}

	// ✅ Send temp password email (with error handling)
	go func() {
		if err := hlpAuth.SendTempPasswordEmail(
			apiResp.Data.Details.Email,
			apiResp.Data.Details.Username,
			apiResp.Data.Details.InstitutionCode,
			apiResp.Data.Details.Password,
		); err != nil {
			// Optional: log the error (don't break user flow)
			fmt.Println("Failed to send temp password email:", err)
		}

	}()

	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		apiResp.Data.Message, result, http.StatusCreated)
}

func LoginUser(c fiber.Ctx) error {
	var req mdlAuth.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// Call external login API
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/helper/auth/user-logs/login"
	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, _ := json.Marshal(req)
	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Request to external API failed", err, http.StatusInternalServerError)
	}

	// Unmarshal to typed struct
	var apiResp mdlAuth.LoginAPIResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}

	if apiResp.RetCode != "201" {
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode,
			apiResp.Data.Message, nil, http.StatusBadRequest)
	}

	// Check if user exists in DB
	userID, err := scpAuth.GetUserIDByEmail(apiResp.Data.Details.Email)
	if err != nil || userID == 0 {
		return helper.JSONResponseWithDataV1(c, respcode.ERR_CODE_404, "User not found in DB", nil, http.StatusNotFound)
	}
	apiResp.Data.Details.UserID = userID

	// Update internal DB (last_login, is_active)
	if err := scpAuth.LoginUser(apiResp.Data.Details); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303,
			"Failed to update login state", err, http.StatusInternalServerError)
	}

	// Success: return user login details
	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		apiResp.Data.Message, apiResp.Data.Details, http.StatusOK)
}

func LogoutUser(c fiber.Ctx) error {
	var req mdlAuth.LogoutRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(
			c,
			respcode.ERR_CODE_301,
			"Parsing request body failed",
			err,
			http.StatusBadRequest,
		)
	}

	// Call external logout API
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") +
		"/soteria-go/api/public/helper/auth/user-logs/logout"

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, _ := json.Marshal(req)
	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(
			c,
			respcode.ERR_CODE_405,
			"Request to external API failed",
			err,
			http.StatusInternalServerError,
		)
	}

	// Parse external API response
	var apiResp mdlAuth.LogoutAPIResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(
			c,
			respcode.ERR_CODE_310,
			"Failed to parse external API response",
			err,
			http.StatusInternalServerError,
		)
	}

	// Handle non-success logout
	if apiResp.RetCode != "202" {
		return helper.JSONResponseWithErrorV1(
			c,
			apiResp.RetCode,
			apiResp.Data.Message,
			nil,
			http.StatusBadRequest,
		)
	}

	// Get internal user ID by email
	userID, err := scpAuth.GetUserIDByEmail(apiResp.Data.Details.Email)
	if err != nil || userID == 0 {
		return helper.JSONResponseWithDataV1(
			c,
			respcode.ERR_CODE_404,
			"User not found in DB",
			nil,
			http.StatusNotFound,
		)
	}

	apiResp.Data.Details.UserID = userID

	// Update internal DB (set inactive, clear login state)
	if err := scpAuth.LogoutUser(userID); err != nil {
		return helper.JSONResponseWithErrorV1(
			c,
			respcode.ERR_CODE_303,
			"Failed to update logout state",
			err,
			http.StatusInternalServerError,
		)
	}

	// Success response
	return helper.JSONResponseWithDataV1(
		c,
		apiResp.RetCode,
		apiResp.Data.Message,
		apiResp.Data.Details,
		http.StatusOK,
	)
}

func ChangeTempPassword(c fiber.Ctx) error {
	var req mdlAuth.ChangePasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// External API call
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/helper/auth/security-management/change-password"
	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, _ := json.Marshal(req)
	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Request to external API failed", err, http.StatusInternalServerError)
	}

	// Parse external response
	var apiResp mdlAuth.ChangePasswordAPIResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}

	if apiResp.RetCode != "203" {
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode,
			apiResp.Data.Message, nil, http.StatusBadRequest)
	}

	// Update local DB
	if err := scpAuth.ChangeTempPassword(apiResp.Data.Details); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303,
			"Failed to update password locally", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		apiResp.Data.Message, apiResp.Data.Details, http.StatusOK)
}

func DeleteUser(c fiber.Ctx) error {
	var req mdlAuth.DeleteUserRequest

	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// Get Bearer Token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return helper.JSONResponseWithErrorV1(c, "401",
			"Missing Authorization token", nil, http.StatusUnauthorized)
	}

	// Call external API
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/helper/auth/user-management/delete-user"
	headers := map[string]string{
		"Content-Type":  "application/json",
		"x-api-key":     utilityV1.GetEnv("CAGABAY_API_KEY"),
		"Authorization": authHeader,
	}

	body, _ := json.Marshal(req)
	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Request to external API failed", err, http.StatusInternalServerError)
	}

	// Parse response
	var apiResp mdlAuth.DeleteUserAPIResponse
	respBytes, _ := json.Marshal(resp)

	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}

	if apiResp.RetCode != "210" {
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode,
			apiResp.Data.Message, nil, http.StatusBadRequest)
	}

	// Delete internally
	if err := scpAuth.DeleteUserByIdentity(req.UserIdentity); err != nil {
		return helper.JSONResponseWithErrorV1(c, "314",
			"Deleting Data Failed", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		apiResp.Data.Message, nil, http.StatusOK)
}

func UpdateUser(c fiber.Ctx) error {
	username := c.Params("username")
	var req mdlAuth.UpdateUserRequest

	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// Get Bearer Token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return helper.JSONResponseWithErrorV1(c, "401",
			"Missing Authorization token", nil, http.StatusUnauthorized)
	}

	// Call external API
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") +
		"/soteria-go/api/public/helper/auth/user-management/update-user/staff/" + username

	headers := map[string]string{
		"Content-Type":  "application/json",
		"x-api-key":     utilityV1.GetEnv("CAGABAY_API_KEY"),
		"Authorization": authHeader,
	}

	body, _ := json.Marshal(req)
	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Request to external API failed", err, http.StatusInternalServerError)
	}

	// Unmarshal external response
	var apiResp mdlAuth.UpdateUserAPIResponse
	respBytes, _ := json.Marshal(resp)

	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}

	// API failure
	if apiResp.RetCode != "203" && apiResp.RetCode != "204" {
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode,
			apiResp.Data.Message, nil, http.StatusBadRequest)
	}

	// Get user ID from DB
	userID, err := scpAuth.GetUserIDByEmail(apiResp.Data.Details.Email)
	if err != nil || userID == 0 {
		return helper.JSONResponseWithDataV1(c, "404", "User not found in DB", nil, http.StatusNotFound)
	}
	apiResp.Data.Details.UserID = userID

	// Update internal DB
	if err := scpAuth.UpdateUser(apiResp.Data.Details); err != nil {
		return helper.JSONResponseWithErrorV1(c, "304",
			"Updating Data Failed", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		apiResp.Data.Message, apiResp.Data.Details, http.StatusOK)
}

// ============================================
// FORGOT PASSWORD ENDPOINT
// ============================================
func ForgotPassword(c fiber.Ctx) error {
	var req mdlAuth.ForgotPasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301, "Invalid request body", err, http.StatusBadRequest)
	}

	// Validate email
	if req.Email == "" {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_401, "Email is required", nil, http.StatusBadRequest)
	}

	// Check if user exists with this email
	_, err := scpAuth.GetUserIdByEmail(req.Email)
	if err != nil {
		// For security, don't reveal if email exists or not
		log.Printf("User not found for email: %s", req.Email)
		return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200, "If the email exists, a reset link has been sent", nil, http.StatusOK)
	}

	// Generate reset token
	token, err := scpAuth.GenerateResetToken()
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_305, "Failed to generate reset token", err, http.StatusInternalServerError)
	}

	// Save token to database
	if err := scpAuth.SaveResetToken(req.Email, token); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303, "Failed to save reset token", err, http.StatusInternalServerError)
	}

	// Send reset email (async)
	go func() {
		if err := hlpAuth.SendPasswordResetEmail(req.Email, token); err != nil {
			log.Printf("Failed to send reset email: %v", err)
		}
	}()

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200, "Reset link has been sent", map[string]any{"token": token}, http.StatusOK)
}

func VerifyResetToken(c fiber.Ctx) error {
	req := mdlAuth.VerifyResetToken{}
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301, "Invalid request body", err, http.StatusBadRequest)
	}
	token := req.Token
	if token == "" {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_401, "Reset token is required", nil, http.StatusBadRequest)
	}

	// Validate token using boolean function
	isValid := scpAuth.IsResetTokenValid(token)
	if !isValid {
		log.Printf("Invalid reset token attempted: %s", token)
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_104, "Invalid or expired reset token", nil, http.StatusBadRequest)
	}

	// Get email from token to return in response (optional)
	email, err := scpAuth.GetEmailFromToken(token)
	if err != nil {
		log.Printf("Valid token but failed to get email for token %s: %v", token, err)
		// Still return success since token is valid, just without email
		return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200, "Token is valid", nil, http.StatusOK)
	}

	// Get user details for the response
	username, _, err := scpAuth.GetUserDetailsByEmail(email)
	if err != nil {
		log.Printf("Valid token but user not found for email %s: %v", email, err)
		// Still return success since token is valid
		return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200, "Token is valid", nil, http.StatusOK)
	}

	log.Printf("Reset token validated successfully for user: %s", username)

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200, "Token is valid", nil, http.StatusOK)
}
