package whatsapp

const (
	// MessageWebhookMessageTypeText represents a text message
	MessageWebhookMessageTypeText = "text"
)

// MessageWebhookRequest is the webhook request from whatsapp when a new message is received
type MessageWebhookRequest struct {
	Object string                       `json:"object"`
	Entry  []MessageWebhookRequestEntry `json:"entry"`
}

type MessageWebhookRequestMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type UserProfile struct {
	Name string `json:"name"`
}

type MessageContact struct {
	Profile    UserProfile `json:"profile"`
	WhatsappID string      `json:"wa_id"`
}

type MessageWebhookText struct {
	Body string `json:"body"`
}

type MessageWebhookMessage struct {
	From      string              `json:"from"`
	ID        string              `json:"id"`
	Timestamp string              `json:"timestamp"`
	Text      *MessageWebhookText `json:"text"`
	Type      string              `json:"type"`
}

type MessageWebhookValue struct {
	MessagingProduct string                        `json:"messaging_product"`
	Metadata         MessageWebhookRequestMetadata `json:"metadata"`
	Contacts         *[]MessageContact             `json:"contacts"`
	Messages         *[]MessageWebhookMessage      `json:"messages"`
	Statuses         *[]MessageWebhookStatus       `json:"statuses"`
}

type MessageWebhookStatus struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	RecipientID  string `json:"recipient_id"`
	Conversation struct {
		ID                  string `json:"id"`
		ExpirationTimestamp string `json:"expiration_timestamp"`
		Origin              struct {
			Type string `json:"type"`
		} `json:"origin"`
	} `json:"conversation"`
	Pricing struct {
		Billable     bool   `json:"billable"`
		PricingModel string `json:"pricing_model"`
		Category     string `json:"category"`
	} `json:"pricing"`
}

type MessageWebhookRequestChanges struct {
	Value MessageWebhookValue `json:"value"`
	Field string              `json:"field"`
}

type MessageWebhookRequestEntry struct {
	ID      string                         `json:"id"`
	Changes []MessageWebhookRequestChanges `json:"changes"`
}
