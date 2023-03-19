package nexmo

// SmsSendParams are parameters for sending an SMS message
type SmsSendParams struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

// SmsSendResponse is the response after sending an SMS
type SmsSendResponse struct {
	MessageCount string `json:"message-count"`
	Messages     []struct {
		To               string `json:"to"`
		MessageID        string `json:"message-id"`
		Status           string `json:"status"`
		RemainingBalance string `json:"remaining-balance"`
		MessagePrice     string `json:"message-price"`
		Network          string `json:"network"`
		ClientRef        string `json:"client-ref"`
		AccountRef       string `json:"account-ref"`
	} `json:"messages"`
}
