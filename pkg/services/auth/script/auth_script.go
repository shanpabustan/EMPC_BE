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

func RegisterUser(data *mdlAuth.RegisterStaffResult) (*mdlAuth.RegisterStaffResult, error) {
	dataJSON, _ := json.Marshal(data)
	var result mdlAuth.RegisterStaffResult
	var dbResult string

	err := config.DBConnList[0].Debug().Raw(
		`SELECT register_user($1)`, string(dataJSON),
	).Scan(&dbResult).Error
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(dbResult), &result); err != nil {
		return nil, err
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

func LogoutUser(userId int) error {
	// Plain SQL query
	query := `
		UPDATE users
		SET is_active = $1
		WHERE id = $2
	`

	return config.DBConnList[0].Exec(
		query,
		false,
		userId,
	).Error
}

func ChangeTempPassword(data *mdlAuth.ChangePasswordResult) error {
	query := `
		UPDATE users
		SET password = $1,
		    requires_password_reset = false,
		    last_password_reset = NOW(),
		    updated_at = NOW()
		WHERE email = $2
	`

	return config.DBConnList[0].Exec(
		query,
		data.Password,
		data.Email,
	).Error
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

func GetUserIDByEmail(email string) (int, error) {
	var userID int
	query := `
		SELECT id
		FROM users
		WHERE email = ? AND deleted_at IS NULL
		LIMIT 1
	`
	err := config.DBConnList[0].Raw(query, email).Scan(&userID).Error
	if err != nil {
		return 0, err
	}

	return userID, nil
}

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
