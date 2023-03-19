package requests

import "github.com/NdoleStudio/discusswithai/pkg/services"

// NexmoReceiveRequest is an incoming sms message from the nexmo API
type NexmoReceiveRequest struct {
	request
	APIKey           string `json:"api-key"`
	Msisdn           string `json:"msisdn"`
	To               string `json:"to"`
	MessageID        string `json:"messageId"`
	Text             string `json:"text"`
	Type             string `json:"type"`
	Keyword          string `json:"keyword"`
	MessageTimestamp string `json:"message-timestamp"`
	Timestamp        string `json:"timestamp"`
	Nonce            string `json:"nonce"`
	Concat           string `json:"concat"`
	ConcatRef        string `json:"concat-ref"`
	ConcatTotal      string `json:"concat-total"`
	ConcatPart       string `json:"concat-part"`
	Data             string `json:"data"`
	Udh              string `json:"udh"`
}

// Sanitize sets defaults to MessageReceive
func (request *NexmoReceiveRequest) Sanitize() NexmoReceiveRequest {
	request.To = request.sanitizePhoneNumber(request.To)
	request.Text = request.sanitizeString(request.Text)
	request.Msisdn = request.sanitizePhoneNumber(request.Msisdn)
	return *request
}

// ToReceiveParams converts NexmoReceiveRequest to services.NexmoReceiveParams
func (request *NexmoReceiveRequest) ToReceiveParams() *services.NexmoReceiveParams {
	return &services.NexmoReceiveParams{
		From:        request.Msisdn,
		To:          request.To,
		Message:     request.Text,
		IsMultipart: request.Concat == "true",
		Reference:   request.ConcatRef,
	}
}
