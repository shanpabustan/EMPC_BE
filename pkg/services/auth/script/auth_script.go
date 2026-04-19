package scpAuth

import (
	mdlAuth "EMPC_BE/pkg/services/auth/model"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	config "EMPC_BE/pkg/config"
)

// func RegisterUser(data *mdlAuth.RegisterStaffResult) (*mdlAuth.RegisterStaffResult, error) {
// 	dataJSON, _ := json.Marshal(data)
// 	var result mdlAuth.RegisterStaffResult
// 	var dbResult string

// 	err := config.DBConnList[0].Debug().Raw(
// 		`SELECT register_user($1)`, string(dataJSON),
// 	).Scan(&dbResult).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := json.Unmarshal([]byte(dbResult), &result); err != nil {
// 		return nil, err
// 	}

// 	return &result, nil
// }

func RegisterUser(data *mdlAuth.RegisterStaffResult) (*mdlAuth.RegisterStaffResult, error) {
	var dbResult string

	sqlDB, err := config.DBConnList[0].DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	err = sqlDB.QueryRow(`
		SELECT register_user(
			$1,  -- username
			$2,  -- staff_id
			$3,  -- first_name
			$4,  -- middle_name
			$5,  -- last_name
			$6,  -- email
			$7,  -- phone_no
			$8,  -- birthdate
			$9,  -- password
			$10, -- institution_id
			$11, -- institution_code
			$12  -- institution_name
		)`,
		data.Username,
		data.StaffID,
		data.FirstName,
		data.MiddleName,
		data.LastName,
		data.Email,
		data.PhoneNo,
		data.Birthdate,
		data.Password,
		data.InstitutionID,
		data.InstitutionCode,
		data.InstitutionName,
	).Scan(&dbResult)
	if err != nil {
		return nil, fmt.Errorf("register_user failed: %w", err)
	}

	var result mdlAuth.RegisterStaffResult
	if err := json.Unmarshal([]byte(dbResult), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal db result: %w", err)
	}

	return &result, nil
}

func LoginUser(data *mdlAuth.LoginResult) error {
	// Plain SQL query

	query := `
		UPDATE users
		SET last_login = NOW(),
		    is_active = $1
		WHERE id = $2
	`

	return config.DBConnList[0].Exec(
		query,
		true, // set active to true
		data.UserID,
	).Error
}

// func LogoutUser(userId int) error {
// 	// Plain SQL query
// 	query := `
// 		UPDATE users
// 		SET is_active = $1
// 		WHERE id = $2
// 	`

// 	return config.DBConnList[0].Exec(
// 		query,
// 		false,
// 		userId,
// 	).Error
// }

// script/auth.go
func LogoutUser(userID int) error {
	query := `
		UPDATE users
		SET is_active = false,
		    updated_at = NOW()
		WHERE id = $1
	`

	result := config.DBConnList[0].Exec(query, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to update logout state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	return nil
}

// script/auth.go
func GetUserByIdentity(identity string) (*mdlAuth.User, error) {
	var user mdlAuth.User

	// Try to find user by email, username, or staff_id
	query := `
		SELECT 
			u.id, 
			u.username, 
			u.staff_id, 
			u.first_name, 
			u.middle_name, 
			u.last_name, 
			u.email, 
			u.phone_no, 
			u.birthdate,
			u.role_id,
			u.is_active,
			u.requires_password_reset,
			u.institution_id,
			u.institution_code,
			u.institution_name,
			COALESCE(r.role_name, '') as role_name
		FROM users u
		LEFT JOIN sys_roles r ON u.role_id = r.id
		WHERE (u.email = $1 OR u.username = $1 OR u.staff_id = $1) 
		AND u.deleted_at IS NULL
		LIMIT 1
	`

	err := config.DBConnList[0].Raw(query, identity).Scan(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user by identity: %w", err)
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user not found with identity: %s", identity)
	}

	return &user, nil
}

// script/auth.go - Update this function

func ChangeTempPassword(userID int, hashedPassword string) error {
	db := config.DBConnList[0]

	query := `
		UPDATE users 
		SET password = $1, 
		    requires_password_reset = false, 
		    last_password_reset = NOW(),
		    updated_at = NOW()
		WHERE id = $2
	`

	result := db.Exec(query, hashedPassword, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	return nil
}

func DeleteUserByIdentity(userIdentity string) error {
	query := `
		UPDATE users
		SET deleted_at = NOW(),
		    is_active = false,
		    updated_at = NOW()
		WHERE email = $1
		   OR username = $1
	`

	return config.DBConnList[0].Exec(
		query,
		userIdentity,
	).Error
}

func UpdateUser(data *mdlAuth.UpdateUserResult) error {
	query := `
		UPDATE users
		SET
			username = $1,
			staff_id = $2,
			first_name = $3,
			middle_name = $4,
			last_name = $5,
			email = $6,
			phone_no = $7,
			birthdate = $8,
			institution_code = $9,
			updated_at = NOW()
		WHERE id = $10
	`

	return config.DBConnList[0].Debug().Exec(
		query,
		data.Username,
		data.StaffID,
		data.FirstName,
		data.MiddleName,
		data.LastName,
		data.Email,
		data.PhoneNo,
		data.Birthdate,
		data.InstitutionCode,
		data.UserID,
	).Error
}

// =================================================
// FORGOT PASSWORD
// =================================================

func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SaveResetToken stores the reset token in database
func SaveResetToken(email, token string) error {
	db := &config.DBConnList[0]
	// Set token expiration to 5 minutes
	expiresAt := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")

	query := `
        INSERT INTO public.password_reset_tokens (email, token, expires_at)
        VALUES (?, ?, ?)
    `

	if err := db.Exec(query, email, token, expiresAt).Error; err != nil {
		return fmt.Errorf("failed to save reset token: %v", err)
	}

	return nil
}

// IsResetTokenValid checks if token is valid and not expired (returns bool)
func IsResetTokenValid(token string) bool {
	db := &config.DBConnList[0]

	var count int

	query := `
        SELECT COUNT(*) 
        FROM public.password_reset_tokens 
        WHERE token = $1 
        AND used_at IS NULL
        AND expires_at > NOW()
    `

	if err := db.Raw(query, token).Scan(&count).Error; err != nil {
		log.Printf("Database error checking token: %v", err)
		return false
	}

	return count > 0
}

// GetEmailFromToken retrieves email from valid token
func GetEmailFromToken(token string) (string, error) {
	db := &config.DBConnList[0]

	var email string
	query := `
        SELECT email 
        FROM public.password_reset_tokens 
        WHERE token = ? AND used_at IS NULL
        LIMIT 1
    `

	if err := db.Raw(query, token).Scan(&email).Error; err != nil {
		return "", fmt.Errorf("failed to get email from token: %v", err)
	}

	if email == "" {
		return "", fmt.Errorf("invalid token or token already used")
	}

	return email, nil
}

// MarkTokenAsUsed invalidates the token after use
func MarkTokenAsUsed(token string) error {
	db := &config.DBConnList[0]

	query := `
        UPDATE public.password_reset_tokens 
        SET used_at = NOW() 
        WHERE token = ?
    `

	if err := db.Exec(query, token).Error; err != nil {
		return fmt.Errorf("failed to mark token as used: %v", err)
	}

	return nil
}

// GetUserIdByEmail retrieves user ID by email to verify existence
func GetUserIdByEmail(email string) (int, error) {
	db := &config.DBConnList[0]

	var userID int
	query := `
        SELECT id 
        FROM public.users 
        WHERE email = ? 
        LIMIT 1
    `

	if err := db.Raw(query, email).Scan(&userID).Error; err != nil {
		return 0, fmt.Errorf("failed to find user by email: %v", err)
	}

	if userID == 0 {
		return 0, fmt.Errorf("user not found")
	}

	return userID, nil
}

// GetUserDetailsByEmail retrieves username and institution code for password reset
func GetUserDetailsByEmail(email string) (string, string, error) {
	db := &config.DBConnList[0]

	var user struct {
		Username        string `json:"username"`
		InstitutionCode string `json:"institution_code"`
	}

	query := `
        SELECT username, institution_code 
        FROM public.users 
        WHERE email = ? 
        LIMIT 1
    `

	if err := db.Raw(query, email).Scan(&user).Error; err != nil {
		return "", "", fmt.Errorf("failed to find user by email: %v", err)
	}

	if user.Username == "" {
		return "", "", fmt.Errorf("user not found")
	}

	return user.Username, user.InstitutionCode, nil
}

////////////////////////////////////
// HELPER FUNCTIONS
////////////////////////////////////

// func GetUserIDByEmail(email string) (int, error) {
// 	var userID int
// 	query := `
// 		SELECT id
// 		FROM users
// 		WHERE email = ? AND deleted_at IS NULL
// 		LIMIT 1
// 	`
// 	err := config.DBConnList[0].Raw(query, email).Scan(&userID).Error
// 	if err != nil {
// 		return 0, err
// 	}

// 	return userID, nil
// }

// Shortest possible version
func GetUserByUsername(username string) (*mdlAuth.UserWithPermissions, error) {
	db := config.DBConnList[0]

	var jsonStr string
	if err := db.Raw(`SELECT get_user_by_username($1)::text`, username).Scan(&jsonStr).Error; err != nil {
		return nil, err
	}

	var user mdlAuth.UserWithPermissions
	if err := json.Unmarshal([]byte(jsonStr), &user); err != nil {
		return nil, err
	}

	if user.RoleName == "super_admin" {
		user.Permissions = []string{"Full Access"}
	}

	return &user, nil
}

func GetUserWithNavigation(username string) (*mdlAuth.UserWithNavigationResponse, error) {
	db := config.DBConnList[0]

	var jsonStr string
	if err := db.Raw(`SELECT get_user_with_navigation($1)::text`, username).Scan(&jsonStr).Error; err != nil {
		return nil, fmt.Errorf("failed to get user with navigation: %w", err)
	}

	if jsonStr == "" || jsonStr == "null" {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	var user mdlAuth.UserWithNavigationResponse
	if err := json.Unmarshal([]byte(jsonStr), &user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &user, nil
}

func GetOrCreateUserFromLogin(loginDetails *mdlAuth.LoginResult) (*mdlAuth.User, error) {
	db := config.DBConnList[0]

	var user mdlAuth.User

	// Try to find user by staff_id or email
	query := `
		SELECT id, staff_id, username, email, first_name, middle_name, last_name, role_id
		FROM users 
		WHERE staff_id = $1 OR email = $2
		LIMIT 1
	`

	err := db.Raw(query, loginDetails.StaffID, loginDetails.Email).Scan(&user).Error
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// If user exists, return it
	if user.ID > 0 {
		return &user, nil
	}

	// Create new user from login details
	insertQuery := `
		INSERT INTO users (
			staff_id, username, email, first_name, middle_name, last_name,
			institution_id, institution_code, institution_name, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW(), NOW())
		RETURNING id, staff_id, username, email, first_name, middle_name, last_name, role_id
	`

	err = db.Raw(insertQuery,
		loginDetails.StaffID,
		loginDetails.Username,
		loginDetails.Email,
		loginDetails.FirstName,
		loginDetails.MiddleName,
		loginDetails.LastName,
		loginDetails.InstitutionID,
		loginDetails.InstitutionCode,
		loginDetails.InstitutionName,
	).Scan(&user).Error

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// UpdateUserLogin updates user's login timestamp
func UpdateUserLogin(userID int) error {
	db := config.DBConnList[0]

	query := `
		UPDATE users 
		SET last_login = NOW(), 
		    is_active = true,
		    updated_at = NOW()
		WHERE id = $1
	`

	result := db.Exec(query, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to update user login: %w", result.Error)
	}

	return nil
}

// GetUserIDByEmail gets user ID by email (keep for backward compatibility)
func GetUserIDByEmail(email string) (int, error) {
	db := config.DBConnList[0]

	var userID int
	query := `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`

	err := db.Raw(query, email).Scan(&userID).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}

	if userID == 0 {
		return 0, fmt.Errorf("user not found with email: %s", email)
	}

	return userID, nil
}

// script/auth.go - Add this function

func UpdateUserFromLogin(loginDetails *mdlAuth.LoginResult) error {
	db := config.DBConnList[0]

	query := `
		UPDATE users 
		SET first_name = $1,
		    middle_name = $2,
		    last_name = $3,
		    email = $4,
		    phone_no = $5,
		    institution_id = $6,
		    institution_code = $7,
		    institution_name = $8,
		    username = $9,
		    updated_at = NOW()
		WHERE staff_id = $10 OR email = $4
	`

	result := db.Exec(query,
		loginDetails.FirstName,
		loginDetails.MiddleName,
		loginDetails.LastName,
		loginDetails.Email,
		loginDetails.PhoneNo,
		loginDetails.InstitutionID,
		loginDetails.InstitutionCode,
		loginDetails.InstitutionName,
		loginDetails.Username,
		loginDetails.StaffID,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	return nil
}

// script/auth.go - Add this function

func UpdateUserLoginByUsername(username string) error {
	db := config.DBConnList[0]

	query := `
		UPDATE users 
		SET last_login = NOW(), 
		    is_active = true,
		    requires_password_reset = false,
		    updated_at = NOW()
		WHERE username = $1
	`

	result := db.Exec(query, username)
	if result.Error != nil {
		return fmt.Errorf("failed to update user login: %w", result.Error)
	}

	return nil
}
