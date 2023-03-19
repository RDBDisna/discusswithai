package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MessageService is the API client for the `/messages` endpoint
type MessageService service

// Send a whatsapp message to a user
//
// API Docs: https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
func (service *MessageService) Send(ctx context.Context, params *MessageSendParams) (*MessageSendResponse, *Response, error) {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                params.To,
		"type":              MessageWebhookMessageTypeText,
		"text": map[string]any{
			"body": params.Body,
		},
	}

	if params.PreviousMessageID != nil {
		payload["context"] = map[string]string{
			"message_id": *params.PreviousMessageID,
		}
	}

	request, err := service.client.newRequest(ctx, http.MethodPost, fmt.Sprintf("/v16.0/%s/messages", params.From), payload)
	if err != nil {
		return nil, nil, err
	}

	response, err := service.client.do(request)
	if err != nil {
		return nil, response, err
	}

	message := new(MessageSendResponse)
	if err = json.Unmarshal(*response.Body, message); err != nil {
		return nil, response, err
	}

	return message, response, nil
}
