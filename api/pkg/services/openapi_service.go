package services

import (
	"context"
	"fmt"

	"github.com/NdoleStudio/discusswithai/pkg/entities"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/davecgh/go-spew/spew"
	"github.com/palantir/stacktrace"
	"github.com/sashabaranov/go-openai"
)

// OpenAPIService is responsible for managing openapi events
type OpenAPIService struct {
	logger telemetry.Logger
	tracer telemetry.Tracer
	client *openai.Client
}

// NewOpenAPIService creates a new OpenAPIService
func NewOpenAPIService(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	client *openai.Client,
) (s *OpenAPIService) {
	return &OpenAPIService{
		logger: logger.WithService(fmt.Sprintf("%T", s)),
		tracer: tracer,
		client: client,
	}
}

// OpenAPICompletionParams are parameters for calling the completion api
type OpenAPICompletionParams struct {
	ChannelID string
	Channel   entities.Channel
	Message   string
}

// GetChatCompletion returns the chat completion using GPT
func (service *OpenAPIService) GetChatCompletion(ctx context.Context, params *OpenAPICompletionParams) (string, error) {
	ctx, span, _ := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	response, err := service.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 200,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "As a user chatting with an AI language model",
			},
			{
				Role:    "user",
				Content: params.Message,
			},
		},
	})
	if err != nil {
		msg := fmt.Sprintf("cannot create completion for prompt [%s]", params.Message)
		return "", service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	spew.Dump(response)

	return response.Choices[0].Message.Content, nil
}
