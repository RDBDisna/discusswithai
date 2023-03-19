package services

import (
	"context"
	"fmt"
	"time"

	"github.com/NdoleStudio/discusswithai/pkg/cache"
	"github.com/NdoleStudio/discusswithai/pkg/entities"
	"github.com/NdoleStudio/discusswithai/pkg/nexmo"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/palantir/stacktrace"
)

const (
	smsCharacterLimit = 160 * 5
)

// NexmoService is responsible for managing nexmo events
type NexmoService struct {
	logger         telemetry.Logger
	tracer         telemetry.Tracer
	client         *nexmo.Client
	openAPIService *OpenAPIService
	cache          cache.Cache
}

// NewNexmoService creates a new NexmoService
func NewNexmoService(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	client *nexmo.Client,
	cache cache.Cache,
	openAPIService *OpenAPIService,
) (s *NexmoService) {
	return &NexmoService{
		logger:         logger.WithService(fmt.Sprintf("%T", s)),
		tracer:         tracer,
		client:         client,
		cache:          cache,
		openAPIService: openAPIService,
	}
}

// NexmoReceiveParams represents a nexmo SMS message
type NexmoReceiveParams struct {
	From        string
	To          string
	Message     string
	IsMultipart bool
	Reference   string
}

// Receive an incoming SMS from nexmo
func (service *NexmoService) Receive(ctx context.Context, params *NexmoReceiveParams) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	if params.IsMultipart {
		service.handleMultipartSMS(ctx, params)
		return
	}

	responseText, err := service.openAPIService.GetChatCompletion(ctx, &OpenAPICompletionParams{
		Channel:   entities.ChannelSMS,
		ChannelID: params.From,
		Message:   params.Message,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot get completion for user [%s] and channel [%s]", params.From, entities.ChannelSMS)
		responseSMS := "We could not generate the completion using chatGPT. Please try again later."
		service.handleCompletionError(ctx, stacktrace.Propagate(err, msg), responseSMS, params)
		return
	}

	if len(responseText) > smsCharacterLimit {
		msg := fmt.Sprintf("The response text [%s] to [%s] contains [%d] characters which is more than [%d] chracter limit", responseText, params.From, len(responseText), smsCharacterLimit)
		responseSMS := fmt.Sprintf("The response text contains %d characters. Contact us at arnold@discusswithai.com to receive responses with more than %d characters via sms.", len(responseText), smsCharacterLimit)
		service.handleCompletionError(ctx, stacktrace.NewError(msg), responseSMS, params)
		return
	}

	response, _, err := service.client.Sms.Send(ctx, &nexmo.SmsSendParams{
		From: params.To,
		To:   params.From,
		Text: responseText,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot send SMS to user [%s] with response [%s]", params.From, responseText)
		ctxLogger.Error(service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg)))
		return
	}

	ctxLogger.Info(fmt.Sprintf("sent response content SMS with id [%s] to [%s] with [%d] characters", response.Messages[0].MessageID, params.To, len(responseText)))
}

func (service *NexmoService) handleCompletionError(ctx context.Context, err error, message string, params *NexmoReceiveParams) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	ctxLogger.Error(stacktrace.Propagate(err, message))

	response, _, err := service.client.Sms.Send(ctx, &nexmo.SmsSendParams{
		From: params.To,
		To:   params.From,
		Text: message,
	})
	if err != nil {
		msg := fmt.Sprintf("cannot send completion error SMS to [%s]", params.From)
		ctxLogger.Error(service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, msg)))
		return
	}

	ctxLogger.Info(fmt.Sprintf("sent completion error message id [%s] to [%s] becasue the text [%s] was [%d] characters", response.Messages[0].MessageID, params.To, params.Message, len(params.Message)))
}

func (service *NexmoService) handleMultipartSMS(ctx context.Context, params *NexmoReceiveParams) {
	ctx, span, ctxLogger := service.tracer.StartWithLogger(ctx, service.logger)
	defer span.End()

	key := fmt.Sprintf("sms.multipart.%s:%s:%s", params.From, params.To, params.Reference)
	if _, err := service.cache.Get(ctx, key); err == nil {
		return
	}

	response, _, err := service.client.Sms.Send(ctx, &nexmo.SmsSendParams{
		From: params.To,
		To:   params.From,
		Text: "We don't yet support text prompts with more than 160 characters.",
	})
	if err != nil {
		ctxLogger.Error(service.tracer.WrapErrorSpan(span, stacktrace.Propagate(err, fmt.Sprintf("cannot multipart content SMS to [%s]", params.From))))
		return
	}

	if err = service.cache.Set(ctx, key, "", time.Hour); err != nil {
		ctxLogger.Error(stacktrace.Propagate(err, fmt.Sprintf("cannot set item in redis with key [%s]", key)))
	}

	ctxLogger.Info(fmt.Sprintf("sent invalid content SMS with id [%s] to [%s] becasue the text [%s] was [%d] characters", response.Messages[0].MessageID, params.To, params.Message, len(params.Message)))
}
