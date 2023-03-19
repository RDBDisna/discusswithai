package nexmo

import (
	"context"
	"encoding/json"
	"net/http"
)

// SMSService is the API client for the `/sms` endpoint
type SMSService service

// Send an outbound SMS from your Vonage account
//
// API Docs: https://developer.vonage.com/en/api/sms
func (service *SMSService) Send(ctx context.Context, params *SmsSendParams) (*SmsSendResponse, *Response, error) {
	request, err := service.client.newRequest(ctx, http.MethodPost, "/sms/json", map[string]string{
		"api_key":    service.client.apiKey,
		"api_secret": service.client.apiSecret,
		"to":         params.To,
		"from":       params.From,
		"text":       params.Text,
	})
	if err != nil {
		return nil, nil, err
	}

	response, err := service.client.do(request)
	if err != nil {
		return nil, response, err
	}

	status := new(SmsSendResponse)
	if err = json.Unmarshal(*response.Body, status); err != nil {
		return nil, response, err
	}

	return status, response, nil
}
