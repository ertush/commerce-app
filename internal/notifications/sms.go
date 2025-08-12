package notifications

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SMSService handles SMS notifications
type SMSService struct {
	APIKey   string
	Username string
	BaseURL  string
}

// NewSMSService creates a new SMS service instance
func NewSMSService() *SMSService {
	apiKey := getEnv("AFRICASTALKING_API_KEY", "sandbox")
	username := getEnv("AFRICASTALKING_USERNAME", "sandbox")
	baseURL := getEnv("AFRICASTALKING_BASE_URL", "https://api.sandbox.africastalking.com")

	return &SMSService{
		APIKey:   apiKey,
		Username: username,
		BaseURL:  baseURL,
	}
}

// SendSMS sends an SMS message
func (s *SMSService) SendSMS(phoneNumber, message string) error {
	// Remove any non-digit characters from phone number
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")

	// Add country code if not present
	if !strings.HasPrefix(phoneNumber, "254") {
		phoneNumber = "254" + phoneNumber
	}

	data := url.Values{}
	data.Set("username", s.Username)
	data.Set("to", phoneNumber)
	data.Set("message", message)

	req, err := http.NewRequest("POST", s.BaseURL+"/version1/messaging", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending SMS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS API returned status: %d", resp.StatusCode)
	}

	return nil
}

// SendOrderNotification sends an SMS notification for a new order
func (s *SMSService) SendOrderNotification(phoneNumber, orderID, customerName string, total float64) error {
	message := fmt.Sprintf("Hello %s! Your order #%s has been placed successfully. Total: $%.2f. Thank you for shopping with us!",
		customerName, orderID, total)

	return s.SendSMS(phoneNumber, message)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
