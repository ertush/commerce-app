package notifications

import (
	"fmt"
	"os"

	"github.com/AndroidStudyOpenSource/africastalking-go/sms"
)

// SMSService handles SMS notifications
type SMSService struct {
	APIKey    string
	Username  string
	SenderID  string
	Recipient string
}

// NewSMSService creates a new SMS service instance
func NewSMSService() *SMSService {
	var (
		apiKey    string = getEnv("AFRICASTALKING_API_KEY", "sandbox")
		username  string = getEnv("AFRICASTALKING_USERNAME", "sandbox")
		senderId  string = getEnv("AFRICASTALKING_SENDER_ID", "africas-talking")
		recipient string = getEnv("AFRICASTALKING_RECIPIENT", "+254700000000")
	)

	return &SMSService{
		APIKey:    apiKey,
		Username:  username,
		SenderID:  senderId,
		Recipient: recipient,
	}
}

// SendOrderNotification sends an SMS notification for a new order
func (s *SMSService) SendOrderNotification(message string) error {

	env := "sandbox"

	smsService := sms.NewService(s.Username, s.APIKey, env)

	smsResponse, err := smsService.Send(s.SenderID, s.Recipient, message)
	if err != nil {
		fmt.Printf("Failed to send SMS notification: %v\n", err)
		return err
	}

	fmt.Printf("SMS notification sent: %v\n", smsResponse)

	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
