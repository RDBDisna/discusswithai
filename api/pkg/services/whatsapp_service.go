package services

import (
	"context"
	"fmt"

	"github.com/NdoleStudio/discusswithai/pkg/entities"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/NdoleStudio/discusswithai/pkg/whatsapp"
	"github.com/palantir/stacktrace"
)

// WhatsappService is responsible for managing whatsapp events
type WhatsappService struct {
	logger         telemetry.Logger
	tracer         telemetry.Tracer
	client         *whatsapp.Client
	openAPIService *OpenAPIService
}

// NewWhatsappService creates a new WhatsappService
func NewWhatsappService(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	client *whatsapp.Client,
	openAPIService *OpenAPIService,
) (s *WhatsappService) {
	return &WhatsappService{
		logger:         logger.WithService(fmt.Sprintf("%T", s)),
		tracer:         tracer,
		client:         client,
		openAPIService: openAPIService,
	}
}

// WhatsappReceiveParams represents a whatsapp message
type WhatsappReceiveParams struct {
	From        string
	To          string
	MessageText string
	Name        string
	Type        string
	MessageID   string
}

// Receive an incoming event from whatsapp
func (service *WhatsappService) Receive(ctx context.Context, params *WhatsappReceiveParams) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	if params.Type != whatsapp.MessageWebhookMessageTypeText {
		service.handleInvalidMessage(ctx, params)
		return
	}

	responseText, err := service.openAPIService.GetChatCompletion(ctx, &OpenAPICompletionParams{
		Channel:   entities.ChannelWhatsapp,
		ChannelID: params.From,
		Name:      params.Name,
		Message:   params.MessageText,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot get completion for user [%s] and channel [%s]", params.From, entities.ChannelWhatsapp)
		responseSMS := "We could not generate the completion using chatGPT. Please try again later."
		service.handleCompletionError(ctx, stacktrace.Propagate(err, msg), responseSMS, params)
		return
	}

	response, _, err := service.client.Message.Send(ctx, &whatsapp.MessageSendParams{
		From:              params.To,
		To:                params.From,
		PreviousMessageID: &params.MessageID,
		Body:              responseText,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot send whatsapp to user [%s] with response [%s]", params.From, responseText)
		ctxLogger.Error(service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg)))
		return
	}

	ctxLogger.Info(fmt.Sprintf("sent response via whatsapp with id [%s] to [%s] with [%d] characters", response.Messages[0].ID, params.To, len(responseText)))
}

func (service *WhatsappService) handleCompletionError(ctx context.Context, err error, message string, params *WhatsappReceiveParams) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	ctxLogger.Error(stacktrace.Propagate(err, message))

	response, _, err := service.client.Message.Send(ctx, &whatsapp.MessageSendParams{
		From:              params.To,
		To:                params.From,
		PreviousMessageID: &params.MessageID,
		Body:              message,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot send completion error SMS to [%s]", params.From)
		ctxLogger.Error(service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg)))
		return
	}

	ctxLogger.Info(fmt.Sprintf("sent completion error message id [%s] to [%s]", response.Messages[0].ID, params.From))
}

func (service *WhatsappService) handleInvalidMessage(ctx context.Context, params *WhatsappReceiveParams) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	response, _, err := service.client.Message.Send(ctx, &whatsapp.MessageSendParams{
		From:              params.To,
		To:                params.From,
		PreviousMessageID: &params.MessageID,
		Body:              fmt.Sprintf("We only support text messages at the moment we plan to support %s content in the future.", params.Type),
	})
	if err != nil {
		ctxLogger.Error(service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot send invalid whatsapp reply to [%s]", params.From))))
		return
	}

	ctxLogger.Info(fmt.Sprintf("sent invalid content whatsappp with id [%s] to [%s] becasue the content type was [%s]", response.Messages[0].ID, params.From, params.Type))
}
