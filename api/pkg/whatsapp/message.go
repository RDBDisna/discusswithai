package whatsapp

// MessageSendParams are parameters for sending a whatsapp message
type MessageSendParams struct {
	From              string  `json:"from"`
	To                string  `json:"to"`
	PreviousMessageID *string `json:"previous_message_id"`
	Body              string  `json:"body"`
}

// MessageSendResponse is the response after a message is sent
type MessageSendResponse struct {
	MessagingProduct string           `json:"messaging_product"`
	Contacts         []MessageContact `json:"contacts"`
	Messages         []struct {
		ID string `json:"id"`
	} `json:"messages"`
}
