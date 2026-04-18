package hlpAuth

import (
	"fmt"
	"net/smtp"

	utils_v1 "github.com/FDSAP-Git-Org/hephaestus/utils/v1"
)

func SendTempPasswordEmail(toEmail, username, instiCode, tempPassword string) error {
	from := utils_v1.GetEnv("SMTP_USER")
	smtpHost := utils_v1.GetEnv("SMTP_HOST")
	password := utils_v1.GetEnv("SMTP_PASS")
	smtpPort := utils_v1.GetEnv("SMTP_PORT")

	subject := "Subject: Welcome to iProvidence - Your Account is Ready\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#f4f6f8;font-family:Arial,Helvetica,sans-serif;">
  <div style="max-width:600px;margin:0 auto;background:#ffffff;border-radius:8px;overflow:hidden;">
    
    <div style="background:#2563eb;color:#ffffff;padding:20px;text-align:center;">
      <h1 style="margin:0;font-size:22px;">Welcome to iProvidence</h1>
    </div>

    <div style="padding:20px;color:#111827;font-size:14px;line-height:1.6;">
      <p>Hi <strong>%s</strong>,</p>

      <p>Welcome to <strong>iProvidence</strong>, your platform for KPI tracking and performance management. Your account has been successfully created and is ready for use.</p>

      <p>Below are your temporary login credentials:</p>

      <div style="background:#f3f4f6;padding:16px;border-radius:6px;margin:15px 0;border-left:4px solid #2563eb;">
        <p style="margin:0;"><strong>Username:</strong> %s</p>
        <p style="margin:0;"><strong>Institution Code:</strong> %s</p>
		  <p style="margin:0;"><strong>Temporary Password:</strong> %s</p>
      </div>

      <div style="background:#fef3c7;padding:12px;border-radius:4px;margin:15px 0;border:1px solid #f59e0b;">
        <p style="margin:0;color:#92400e;">
          <strong>Security Notice:</strong> For your security, please change your temporary password immediately after your first login.
        </p>
      </div>

      <p style="margin-top:20px;">Thank you,<br><strong>The iProvidence Team</strong></p>
    </div>

    <div style="background:#f9fafb;text-align:center;padding:15px;font-size:12px;color:#6b7280;">
      <p style="margin:0;">© 2025 iProvidence. Streamlining performance management.</p>
    </div>

  </div>
</body>
</html>
`, username, username, instiCode, tempPassword)

	message := []byte(subject + mime + htmlBody)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{toEmail},
		message,
	)
}

// SendPasswordResetEmail sends password reset link email
func SendPasswordResetEmail(toEmail, resetToken string) error {
	from := utils_v1.GetEnv("SMTP_USER")
	smtpHost := utils_v1.GetEnv("SMTP_HOST")
	password := utils_v1.GetEnv("SMTP_PASS")
	smtpPort := utils_v1.GetEnv("SMTP_PORT")

	appBaseURL := utils_v1.GetEnv("APP_BASE_URL")
	if appBaseURL == "" {
		appBaseURL = "http://localhost:3000" // fallback
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", appBaseURL, resetToken)

	subject := "Subject: iProvidence - Password Reset Request\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#f4f6f8;font-family:Arial,Helvetica,sans-serif;">
  <div style="max-width:600px;margin:0 auto;background:#ffffff;border-radius:8px;overflow:hidden;">
    
    <div style="background:#dc2626;color:#ffffff;padding:20px;text-align:center;">
      <h1 style="margin:0;font-size:22px;">Password Reset Request</h1>
    </div>

    <div style="padding:20px;color:#111827;font-size:14px;line-height:1.6;">
      <p>We received a request to reset your password for your iProvidence account.</p>

      <p>Click the button below to reset your password:</p>

      <div style="text-align:center;margin:25px 0;">
        <a href="%s" style="background-color:#dc2626;color:#ffffff;padding:12px 30px;text-decoration:none;border-radius:6px;font-weight:bold;display:inline-block;">
          Reset Password
        </a>
      </div>

      <p>Or copy and paste this link in your browser:</p>
      <div style="background:#f3f4f6;padding:12px;border-radius:4px;margin:15px 0;word-break:break-all;">
        <code style="color:#374151;font-size:12px;">%s</code>
      </div>

      <div style="background:#fef3c7;padding:12px;border-radius:4px;margin:15px 0;border:1px solid #f59e0b;">
        <p style="margin:0;color:#92400e;">
          <strong>Important:</strong> This password reset link will expire in <strong>5 minutes</strong> for security reasons.
        </p>
      </div>

      <div style="background:#f0fdf4;padding:12px;border-radius:4px;margin:15px 0;border:1px solid #22c55e;">
        <p style="margin:0;color:#166534;">
          <strong>Security Tip:</strong> If you didn't request this reset, please ignore this email. Your account remains secure.
        </p>
      </div>

      <p style="margin-top:20px;">Thank you,<br><strong>The iProvidence Team</strong></p>
    </div>

    <div style="background:#f9fafb;text-align:center;padding:15px;font-size:12px;color:#6b7280;">
      <p style="margin:0;">© 2025 iProvidence. Streamlining performance management.</p>
    </div>

  </div>
</body>
</html>
`, resetLink, resetLink)

	message := []byte(subject + mime + htmlBody)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{toEmail},
		message,
	)
}

