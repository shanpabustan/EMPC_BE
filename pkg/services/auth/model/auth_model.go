package mdlAuth

// ==========================
// REGISTER STAFF
// ==========================
type StaffRegistrationApiRequest struct {
	Username        string `json:"username"`         // optional
	StaffID         string `json:"staff_id"`         // required
	FirstName       string `json:"first_name"`       // optional
	MiddleName      string `json:"middle_name"`      // optional
	LastName        string `json:"last_name"`        // optional
	Email           string `json:"email"`            // optional
	PhoneNo         string `json:"phone_no"`         // optional
	Birthdate       string `json:"birthdate"`        // optional
	InstitutionCode string `json:"institution_code"` // required
}

type StaffRegistrationAPIResponse struct {
	RetCode string                    `json:"retCode"`
	Message string                    `json:"message"`
	Data    *StaffRegistrationAPIData `json:"data,omitempty"`
}

type StaffRegistrationAPIData struct {
	Message   string               `json:"message"`
	IsSuccess bool                 `json:"isSuccess"`
	Error     interface{}          `json:"error"`
	Details   *RegisterStaffResult `json:"details,omitempty"`
}

type RegisterStaffRequest struct {
	StaffID         string `json:"staff_id"`         // required
	InstitutionCode string `json:"institution_code"` // required
	Birthdate       string `json:"birthdate"`        // required
}

type RegisterStaffResult struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	StaffID         string `json:"staff_id"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNo         string `json:"phone_no"`
	Birthdate       string `json:"birthdate"`
	InstitutionID   int    `json:"institution_id"`
	InstitutionCode string `json:"institution_code"`
	InstitutionName string `json:"institution_name"`
	Password        string `json:"password"` // temporary password
}

// ==========================
// LOGIN STAFF
// ==========================
type LoginRequest struct {
	UserIdentity    string `json:"user_identity"`    // required
	Password        string `json:"password"`         // required
	InstitutionCode string `json:"institution_code"` // optional
}

type LoginAPIResponse struct {
	RetCode string        `json:"retCode"`
	Message string        `json:"message"`
	Data    *LoginAPIData `json:"data,omitempty"`
}

type LoginAPIData struct {
	Message   string       `json:"message"`
	IsSuccess bool         `json:"isSuccess"`
	Error     interface{}  `json:"error"`
	Details   *LoginResult `json:"details,omitempty"`
}

type LoginResult struct {
	UserID                int    `json:"user_id"`
	Username              string `json:"username"`
	StaffID               string `json:"staff_id"`
	FirstName             string `json:"first_name"`
	MiddleName            string `json:"middle_name"`
	LastName              string `json:"last_name"`
	Email                 string `json:"email"`
	PhoneNo               string `json:"phone_no"`
	LastLogin             string `json:"last_login"`
	IsLoggedIn            bool   `json:"is_logged_in"`
	InstitutionID         int    `json:"institution_id"`
	InstitutionCode       string `json:"institution_code"`
	InstitutionName       string `json:"institution_name"`
	RequiresPasswordReset bool   `json:"requires_password_reset"`
	LastPasswordReset     string `json:"last_password_reset"`
	Token                 string `json:"token"`
	Is2FARequired         bool   `json:"is_2fa_required"`
}

// ==========================
// CHANGE TEMPORARY PASSWORD
// ==========================

type ChangePasswordRequest struct {
	Username        string `json:"username"`
	NewPassword     string `json:"new_password"`
	InstitutionCode string `json:"institution_code"`
}

type ChangePasswordAPIResponse struct {
	RetCode string                         `json:"retCode"`
	Message string                         `json:"message"`
	Data    *ChangePasswordAPIResponseData `json:"data,omitempty"`
}

type ChangePasswordAPIResponseData struct {
	Message   string                `json:"message"`
	IsSuccess bool                  `json:"isSuccess"`
	Error     interface{}           `json:"error"`
	Details   *ChangePasswordResult `json:"details,omitempty"`
}

type ChangePasswordResult struct {
	StaffID         string `json:"staff_id"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNo         string `json:"phone_no"`
	Birthdate       string `json:"birthdate"`
	InstitutionName string `json:"institution_name"`
	Password        string `json:"password"`
}

// ==========================
// DELETE USER
// ==========================
type DeleteUserRequest struct {
	UserIdentity    string `json:"user_identity"`
	InstitutionCode string `json:"institution_code"`
}

type DeleteUserAPIResponse struct {
	RetCode string             `json:"retCode"`
	Message string             `json:"message"`
	Data    *DeleteUserAPIData `json:"data,omitempty"`
}

type DeleteUserAPIData struct {
	Message   string      `json:"message"`
	IsSuccess bool        `json:"isSuccess"`
	Details   interface{} `json:"details,omitempty"`
}

// ==========================
// UPDATE USER
// ==========================
type UpdateUserRequest struct {
	Username        string `json:"username"`
	StaffID         string `json:"staff_id"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNo         string `json:"phone_no"`
	InstitutionCode string `json:"institution_code"`
	Birthdate       string `json:"birthdate"`
}

type UpdateUserAPIResponse struct {
	RetCode string          `json:"retCode"`
	Message string          `json:"message"`
	Data    *UpdateUserData `json:"data,omitempty"`
}

type UpdateUserData struct {
	Message   string            `json:"message"`
	IsSuccess bool              `json:"isSuccess"`
	Error     interface{}       `json:"error"`
	Details   *UpdateUserResult `json:"details,omitempty"`
}

type UpdateUserResult struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	StaffID         string `json:"staff_id"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNo         string `json:"phone_no"`
	Birthdate       string `json:"birthdate"`
	LastLogin       string `json:"last_login"`
	InstitutionID   int    `json:"institution_id"`
	InstitutionCode string `json:"institution_code"`
	InstitutionName string `json:"institution_name"`
}

// =================================================
// FORGOT PASSWORD MODELS
// =================================================

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type VerifyResetToken struct {
	Token string `json:"token"`
}

type ResetPasswordTokenRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ==========================
// LOGOUT STAFF
// ==========================
type LogoutRequest struct {
	UserIdentity    string `json:"user_identity"`    // required
	InstitutionCode string `json:"institution_code"` // optional (0000 if non-member)
}

type LogoutAPIResponse struct {
	RetCode string         `json:"retCode"`
	Message string         `json:"message"`
	Data    *LogoutAPIData `json:"data,omitempty"`
}

type LogoutAPIData struct {
	Message   string        `json:"message"`
	IsSuccess bool          `json:"isSuccess"`
	Error     interface{}   `json:"error"`
	Details   *LogoutResult `json:"details,omitempty"`
}

type LogoutResult struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	StaffID         string `json:"staff_id"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNo         string `json:"phone_no"`
	LastLogin       string `json:"last_login"`
	InstitutionID   int    `json:"institution_id"`
	InstitutionCode string `json:"institution_code"`
	InstitutionName string `json:"institution_name"`
}

// ==========================
// VALIDATE TOKEN
// ==========================

type ValidateTokenAPIResponse struct {
	RetCode string             `json:"retCode"`
	Message string             `json:"message"`
	Data    *ValidateTokenData `json:"data"`
}

type ValidateTokenData struct {
	Message   string                `json:"message"`
	IsSuccess bool                  `json:"isSuccess"`
	Error     interface{}           `json:"error"`
	Details   *ValidateTokenDetails `json:"details"`
}

type ValidateTokenDetails struct {
	Username  string `json:"username"`
	InstiCode string `json:"insti_code"`
	InstiName string `json:"insti_name"`
	AppCode   string `json:"app_code"`
	AppName   string `json:"app_name"`
}

type UserWithPermissions struct {
	ID          int64    `json:"id"`
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	StaffID     string   `json:"staff_id"`
	FirstName   string   `json:"first_name"`
	MiddleName  string   `json:"middle_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	RoleID      *int     `json:"role_id"` // Pointer to handle NULL
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
}
