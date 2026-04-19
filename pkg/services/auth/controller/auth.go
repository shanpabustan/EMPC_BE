package crtlAuth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"EMPC_BE/pkg/config"
	helper "EMPC_BE/pkg/global/json_response"
	httpRequestV1 "EMPC_BE/pkg/middleware/httpRequest/v1"
	utilityV1 "EMPC_BE/pkg/middleware/utility/v1"
	hlpAuth "EMPC_BE/pkg/services/auth/helper"
	mdlAuth "EMPC_BE/pkg/services/auth/model"
	scpAuth "EMPC_BE/pkg/services/auth/script"

	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

func CheckUserExists(username, email, staffID string) (bool, error) {
	var count int64

	err := config.DBConnList[0].Debug().Raw(`
		SELECT COUNT(*) FROM public.users
		WHERE deleted_at IS NULL
		AND (
			staff_id = ?
			OR (username = ? AND ? != '')
			OR (email = ? AND ? != '')
		)`,
		staffID,
		username, username,
		email, email,
	).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}

func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(bytes), err
}

func RegisterUser(c fiber.Ctx) error {
	// 1. Parse request body
	var req mdlAuth.RegisterStaffRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	fmt.Println(req)

	// 3. Check if user already exists
	exists, err := CheckUserExists(
		strings.TrimSpace(req.Username),
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.StaffID),
	)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to check user existence", err, http.StatusInternalServerError)
	}
	if exists {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"User already exists", nil, http.StatusConflict)
	}
	fmt.Println(req.StaffID)
	fmt.Println(req.InstitutionCode)
	// 3. Build external API request
	apiReq := mdlAuth.StaffRegistrationApiRequest{
		StaffID:         req.StaffID,
		InstitutionCode: req.InstitutionCode,
		Username:        req.Username,
		FirstName:       req.FirstName,
		MiddleName:      req.MiddleName,
		LastName:        req.LastName,
		Email:           req.Email,
		PhoneNo:         req.PhoneNo,
		Birthdate:       req.Birthdate,
	}

	fmt.Println(apiReq)

	// 4. Marshal request body
	body, err := json.Marshal(apiReq)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Failed to marshal request body", err, http.StatusInternalServerError)
	}

	// 5. Build HTTP request
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/v1/auth/user-management/register-new-user/staff"
	
	httpReq, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Failed to create HTTP request", err, http.StatusInternalServerError)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", utilityV1.GetEnv("CAGABAY_API_KEY"))

	// 6. Execute HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"External API call failed", err, http.StatusInternalServerError)
	}
	defer httpResp.Body.Close()

	// 7. Read raw response
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to read external API response", err, http.StatusInternalServerError)
	}
	log.Printf("External API Status: %d | Response: %s", httpResp.StatusCode, string(respBytes))

	fmt.Sprintf(">>>>>>>>>>>", (httpResp.Body))

	// 8. Unmarshal response
	var apiResp mdlAuth.StaffRegistrationAPIResponse

	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}
	fmt.Println(apiResp)

	// Reformat birthdate from "2000-12-27T00:00:00Z" → "2000-12-27"
	if apiResp.Data != nil && apiResp.Data.Details != nil {
		if apiResp.Data.Details.Birthdate != "" {
			t, err := time.Parse(time.RFC3339, apiResp.Data.Details.Birthdate)
			if err == nil {
				apiResp.Data.Details.Birthdate = t.Format("2006-01-02")
			}
		}
	}

	// 9. Check business logic RetCode
	if apiResp.RetCode != "200" && apiResp.RetCode != "203" {
		msg := "External Registration Failed"
		if apiResp.Data != nil {
			msg = apiResp.Data.Message
		}
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode, msg, nil, http.StatusBadRequest)
	}

	// 10. Guard against missing details
	if apiResp.Data == nil || apiResp.Data.Details == nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"External API returned success but no user details", nil, http.StatusInternalServerError)
	}

	insertreq := apiResp.Data.Details

	// Save plain password before hashing
	plainPassword := insertreq.Password

	if insertreq.Password != "" {
		hashed, err := HashPassword(insertreq.Password)
		if err != nil {
			return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
				"Failed to hash password", err, http.StatusInternalServerError)
		}
		insertreq.Password = hashed
	}

	// 11. Save to internal DB (now stores hashed password)
	result, err := scpAuth.RegisterUser(insertreq)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303,
			"Inserting data failed", err, http.StatusInternalServerError)
	}

	// 12. Send temp password email (async) — use plain password
	if apiResp.Data.Details.Email != "" {
		emailDetails := *apiResp.Data.Details
		go func(d mdlAuth.RegisterStaffResult, pwd string) {
			if err := hlpAuth.SendTempPasswordEmail(
				d.Email, d.Username, d.InstitutionCode, pwd,
			); err != nil {
				log.Printf("Async Email Failed: %v", err)
			}
		}(emailDetails, plainPassword)
	}

	return helper.JSONResponseWithDataV1(c, apiResp.RetCode, apiResp.Message, result, http.StatusCreated)
}

// func LoginUser(c fiber.Ctx) error {
// 	var req mdlAuth.LoginRequest
// 	if err := c.Bind().Body(&req); err != nil {
// 		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
// 			"Parsing request body failed", err, http.StatusBadRequest)
// 	}

// 	// Call external login API
// 	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/v1/auth/user-logs/login"

// 	headers := map[string]string{
// 		"Content-Type": "application/json",
// 		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
// 	}

// 	body, _ := json.Marshal(req)
// 	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
// 	if err != nil {
// 		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
// 			"Request to external API failed", err, http.StatusInternalServerError)
// 	}

// 	// Unmarshal to typed struct
// 	var apiResp mdlAuth.LoginAPIResponse
// 	respBytes, _ := json.Marshal(resp)
// 	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
// 		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
// 			"Failed to parse external API response", err, http.StatusInternalServerError)
// 	}

// 	if apiResp.RetCode != "201" {
// 		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode,
// 			apiResp.Data.Message, nil, http.StatusBadRequest)
// 	}

// 	// Check if user exists in DB
// 	userID, err := scpAuth.GetUserIDByEmail(apiResp.Data.Details.Email)
// 	if err != nil || userID == 0 {
// 		return helper.JSONResponseWithDataV1(c, respcode.ERR_CODE_404, "User not found in DB", nil, http.StatusNotFound)
// 	}
// 	apiResp.Data.Details.UserID = userID

// 	// Update internal DB (last_login, is_active)
// 	if err := scpAuth.LoginUser(apiResp.Data.Details); err != nil {
// 		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303,
// 			"Failed to update login state", err, http.StatusInternalServerError)
// 	}

// 	// Success: return user login details
// 	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
// 		apiResp.Data.Message, apiResp.Data.Details, http.StatusOK)
// }

func LoginUser(c fiber.Ctx) error {
	var req mdlAuth.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// Validate required fields
	if req.UserIdentity == "" || req.Password == "" {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"User identity and password are required", nil, http.StatusBadRequest)
	}

	// First, get user from local database to get institution_code
	user, err := scpAuth.GetUserByIdentity(req.UserIdentity)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_404,
			"User not found", err, http.StatusNotFound)
	}

	// Check if user account is active
	if user.IsActive {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"User is already logged in", nil, http.StatusConflict)
	}

	// Build login request with institution_code from local DB
	loginReq := mdlAuth.LoginRequest{
		UserIdentity:    req.UserIdentity,
		Password:        req.Password,
		InstitutionCode: user.InstitutionCode,
	}

	// Call external login API
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/v1/auth/user-logs/login"

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, err := json.Marshal(loginReq)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to marshal login request", err, http.StatusInternalServerError)
	}

	log.Printf("Login Request URL: %s", apiURL)
	log.Printf("Login Request Body: %s", string(body))

	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_405,
			"Request to external API failed", err, http.StatusInternalServerError)
	}

	// Unmarshal to typed struct
	var apiResp mdlAuth.LoginAPIResponse
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to marshal external API response", err, http.StatusInternalServerError)
	}

	log.Printf("External API Response: %s", string(respBytes))

	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"Failed to parse external API response", err, http.StatusInternalServerError)
	}

	if apiResp.RetCode != "201" {
		errMsg := "Login failed"
		if apiResp.Data != nil && apiResp.Data.Message != "" {
			errMsg = apiResp.Data.Message
		}
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode, errMsg, nil, http.StatusBadRequest)
	}

	// Safely check if Data and Details exist
	if apiResp.Data == nil || apiResp.Data.Details == nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_310,
			"External API returned success but no user details", nil, http.StatusInternalServerError)
	}

	// Check if password is temporary (contains "T3mpP")
	if strings.Contains(req.Password, "T3mpP") {
		log.Printf("Temporary password detected for user: %s", apiResp.Data.Details.Username)

		return helper.JSONResponseWithDataV1(c, "PWD_RESET_REQUIRED",
			"Temporary password detected. Please change your password.",
			map[string]interface{}{
				"username":                apiResp.Data.Details.Username,
				"staff_id":                apiResp.Data.Details.StaffID,
				"email":                   apiResp.Data.Details.Email,
				"token":                   apiResp.Data.Details.Token,
				"requires_password_reset": true,
			}, http.StatusOK)
	}

	// Update user information from login response
	if err := scpAuth.UpdateUserFromLogin(apiResp.Data.Details); err != nil {
		log.Printf("Failed to update user info: %v", err)
	}

	// Update login state in local DB
	if err := scpAuth.UpdateUserLogin(user.ID); err != nil {
		log.Printf("Failed to update login state: %v", err)
	}

	// Fetch user with navigation using the username
	userWithNav, err := scpAuth.GetUserWithNavigation(apiResp.Data.Details.Username)
	if err != nil {
		log.Printf("Failed to fetch user with navigation: %v", err)
		userWithNav = &mdlAuth.UserWithNavigationResponse{
			Email:      apiResp.Data.Details.Email,
			FirstName:  apiResp.Data.Details.FirstName,
			MiddleName: apiResp.Data.Details.MiddleName,
			LastName:   apiResp.Data.Details.LastName,
			StaffID:    apiResp.Data.Details.StaffID,
			RoleID:     user.RoleID,
			RoleName:   "",
			Navigation: []interface{}{},
		}
	}

	// Prepare login response
	loginResponse := &mdlAuth.LoginResponseData{
		Token: apiResp.Data.Details.Token,
		User:  userWithNav,
	}

	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		"Login successful", loginResponse, http.StatusOK)
}

// func LogoutUser(c fiber.Ctx) error {
// 	var req mdlAuth.LogoutRequest
// 	if err := c.Bind().Body(&req); err != nil {
// 		return helper.JSONResponseWithErrorV1(
// 			c,
// 			respcode.ERR_CODE_301,
// 			"Parsing request body failed",
// 			err,
// 			http.StatusBadRequest,
// 		)
// 	}

// 	// Call external logout API
// 	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") +
// 		"/soteria-go/api/public/v1/auth/user-logs/logout"

// 	headers := map[string]string{
// 		"Content-Type": "application/json",
// 		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
// 	}

// 	body, _ := json.Marshal(req)
// 	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
// 	if err != nil {
// 		return helper.JSONResponseWithErrorV1(
// 			c,
// 			respcode.ERR_CODE_405,
// 			"Request to external API failed",
// 			err,
// 			http.StatusInternalServerError,
// 		)
// 	}

// 	// Parse external API response
// 	var apiResp mdlAuth.LogoutAPIResponse
// 	respBytes, _ := json.Marshal(resp)
// 	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
// 		return helper.JSONResponseWithErrorV1(
// 			c,
// 			respcode.ERR_CODE_310,
// 			"Failed to parse external API response",
// 			err,
// 			http.StatusInternalServerError,
// 		)
// 	}

// 	// Handle non-success logout
// 	if apiResp.RetCode != "202" {
// 		return helper.JSONResponseWithErrorV1(
// 			c,
// 			apiResp.RetCode,
// 			apiResp.Data.Message,
// 			nil,
// 			http.StatusBadRequest,
// 		)
// 	}

// 	// Get internal user ID by email
// 	userID, err := scpAuth.GetUserIDByEmail(apiResp.Data.Details.Email)
// 	if err != nil || userID == 0 {
// 		return helper.JSONResponseWithDataV1(
// 			c,
// 			respcode.ERR_CODE_404,
// 			"User not found in DB",
// 			nil,
// 			http.StatusNotFound,
// 		)
// 	}

// 	apiResp.Data.Details.UserID = userID

// 	// Update internal DB (set inactive, clear login state)
// 	if err := scpAuth.LogoutUser(userID); err != nil {
// 		return helper.JSONResponseWithErrorV1(
// 			c,
// 			respcode.ERR_CODE_303,
// 			"Failed to update logout state",
// 			err,
// 			http.StatusInternalServerError,
// 		)
// 	}

// 	// Success response
// 	return helper.JSONResponseWithDataV1(
// 		c,
// 		apiResp.RetCode,
// 		apiResp.Data.Message,
// 		apiResp.Data.Details,
// 		http.StatusOK,
// 	)
// }

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

	// Validate required fields
	if req.UserIdentity == "" {
		return helper.JSONResponseWithErrorV1(
			c,
			respcode.ERR_CODE_400,
			"User identity (email/username/staff_id) is required",
			nil,
			http.StatusBadRequest,
		)
	}

	// Get user from local DB first to get email for external API
	user, err := scpAuth.GetUserByIdentity(req.UserIdentity)
	if err != nil {
		return helper.JSONResponseWithErrorV1(
			c,
			respcode.ERR_CODE_404,
			"User not found in local database",
			err,
			http.StatusNotFound,
		)
	}

	// Prepare logout request for external API
	logoutReq := mdlAuth.LogoutRequest{
		UserIdentity:    user.Email, // Use email for external API
		InstitutionCode: req.InstitutionCode,
	}

	// Call external logout API
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") +
		"/soteria-go/api/public/v1/auth/user-logs/logout"

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, _ := json.Marshal(logoutReq)

	log.Printf("Logout Request URL: %s", apiURL)
	log.Printf("Logout Request Body: %s", string(body))

	resp, err := httpRequestV1.SendRequest(apiURL, "POST", nil, body, headers, nil, 30)
	if err != nil {
		log.Printf("External API Error: %v", err)
		// Continue with local logout even if external API fails
	}

	// Parse external API response if available
	var apiResp mdlAuth.LogoutAPIResponse
	if resp != nil {
		respBytes, _ := json.Marshal(resp)
		log.Printf("External API Response: %s", string(respBytes))

		if err := json.Unmarshal(respBytes, &apiResp); err == nil {
			// Check if external logout was successful
			if apiResp.RetCode != "202" {
				log.Printf("External logout failed with code: %s, message: %s", apiResp.RetCode, apiResp.Message)
				// Continue with local logout anyway
			}
		}
	}

	// Update internal DB (set inactive, clear login state)
	if err := scpAuth.LogoutUser(user.ID); err != nil {
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
		respcode.SUC_CODE_200,
		"Logout successful",
		map[string]interface{}{
			"user_id":   user.ID,
			"username":  user.Username,
			"staff_id":  user.StaffID,
			"email":     user.Email,
			"logout_at": time.Now().Format(time.RFC3339),
		},
		http.StatusOK,
	)
}

func ChangeTempPassword(c fiber.Ctx) error {
	var req mdlAuth.ChangePasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Parsing request body failed", err, http.StatusBadRequest)
	}

	// Validate required fields
	if req.Username == "" || req.NewPassword == "" {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"Username and new password are required", nil, http.StatusBadRequest)
	}

	// Get user from local database to get institution_code
	user, err := scpAuth.GetUserByIdentity(req.Username)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_404,
			"User not found in local database", err, http.StatusNotFound)
	}

	// Build change password request with institution_code
	changeReq := mdlAuth.ChangePasswordRequest{
		Username:        req.Username,
		NewPassword:     req.NewPassword,
		InstitutionCode: user.InstitutionCode,
	}

	// External API call
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/v1/auth/security-management/change-password"

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-api-key":    utilityV1.GetEnv("CAGABAY_API_KEY"),
	}

	body, _ := json.Marshal(changeReq)

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
		errMsg := "Password change failed"
		if apiResp.Data != nil && apiResp.Data.Message != "" {
			errMsg = apiResp.Data.Message
		}
		return helper.JSONResponseWithErrorV1(c, apiResp.RetCode, errMsg, nil, http.StatusBadRequest)
	}

	// Hash the new password before storing locally
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to hash password", err, http.StatusInternalServerError)
	}

	// Update local DB with hashed password
	if err := scpAuth.ChangeTempPassword(user.ID, string(hashedPassword)); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_303,
			"Failed to update password locally", err, http.StatusInternalServerError)
	}

	// Return success without auto-login
	return helper.JSONResponseWithDataV1(c, apiResp.RetCode,
		"Password changed successfully. Please login with your new password.",
		map[string]interface{}{
			"username": req.Username,
		}, http.StatusOK)
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
	apiURL := utilityV1.GetEnv("CAGABAY_BASE_URL") + "/soteria-go/api/public/v1/auth/user-management/delete-user"
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
		"/soteria-go/api/public/v1/auth/user-management/update-user/staff/" + username

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
