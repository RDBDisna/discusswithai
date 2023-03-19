package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/NdoleStudio/discusswithai/pkg/entities"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
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
	Name      string
	Message   string
}

// GetChatCompletion returns the chat completion using GPT
func (service *OpenAPIService) GetChatCompletion(ctx context.Context, params *OpenAPICompletionParams) (string, error) {
	ctx, span, _ := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	name := "a user"
	if params.Name != "" {
		name = params.Name
	}

	response, err := service.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 3000,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: fmt.Sprintf("As %s chatting with the OpenAI language model via %s.", name, params.Channel),
			},
			{
				Role:    "user",
				Name:    params.Name,
				Content: params.Message,
			},
		},
	})
	if err != nil {
		msg := fmt.Sprintf("cannot create completion for prompt [%s]", params.Message)
		return "", service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg))
	}

	return strings.TrimRight(response.Choices[0].Message.Content, "\n"), nil
}
