package main

import (
	"fmt"

	"github.com/Neroframe/AuthService/internal/domain"
	gomail "gopkg.in/mail.v2"
)

func main() {
	// Create a new message
	message := gomail.NewMessage()

	// Set email headers
	message.SetHeader("From", "aidyn.kazhakhmet@nu.edu.kz")
	message.SetHeader("To", "neroframe@proton.me")
	message.SetHeader("Subject", "This is an email sent via Gomail and Gmail SMTP")

	// Set email body to HTML format
	message.SetBody("text/html", buildEmailBody(domain.PurposeEmailVerification, "123123"))

	// In case receipent doesnt support html
	message.AddAlternative("text/plain", "Hello!")

	// Set up the SMTP dialer
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "aidyn.kazhakhmet@nu.edu.kz", "nghu xcow ezsv pkkl")

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		fmt.Println("Error:", err)
		panic(err)
	} else {
		fmt.Println("Email sent successfully!")
	}
}

func buildEmailBody(purpose, code string) string {
	switch purpose {
	case domain.PurposeEmailVerification:
		return fmt.Sprintf(`
			<html>
				<body>
					<h2>Email Verification</h2>
					<p>Use the following code to verify your email:</p>
					<h3>%s</h3>
					<p>This code will expire in 10 minutes.</p>
				</body>
			</html>`, code)

	case domain.PurposeResetPassword:
		return fmt.Sprintf(`
			<html>
				<body>
					<h2>Password Reset</h2>
					<p>Use the following code to reset your password:</p>
					<h3>%s</h3>
					<p>If you didn't request this, ignore the email.</p>
				</body>
			</html>`, code)

	default:
		return "<html><body><p>Invalid email purpose.</p></body></html>"
	}
}
