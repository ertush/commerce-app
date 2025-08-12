package notifications

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailService handles email notifications
type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	AdminEmail   string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@ecommerce.com"),
		AdminEmail:   getEnv("ADMIN_EMAIL", "admin@ecommerce.com"),
	}
}

// SendEmail sends an email
func (e *EmailService) SendEmail(to, subject, body string) error {
	// For development, just log the email
	if e.SMTPUsername == "" || e.SMTPPassword == "" {
		fmt.Printf("EMAIL (DEV MODE): To: %s, Subject: %s, Body: %s\n", to, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", e.FromEmail, to, subject, body)

	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	err := smtp.SendMail(addr, auth, e.FromEmail, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// SendOrderNotificationToAdmin sends an order notification to the administrator
func (e *EmailService) SendOrderNotificationToAdmin(orderID, customerName, customerEmail, customerPhone string, total float64, items []string) error {
	subject := fmt.Sprintf("New Order Placed - Order #%s", orderID)

	body := fmt.Sprintf(`
New order has been placed!

Order Details:
- Order ID: %s
- Customer Name: %s
- Customer Email: %s
- Customer Phone: %s
- Total Amount: $%.2f

Order Items:
%s

Please process this order as soon as possible.
`, orderID, customerName, customerEmail, customerPhone, total, strings.Join(items, "\n"))

	return e.SendEmail(e.AdminEmail, subject, body)
}
