package handlers

import (
	"fmt"
	"strings"

	"github.com/NdoleStudio/discusswithai/pkg/requests"
	"github.com/NdoleStudio/discusswithai/pkg/whatsapp"
	"github.com/palantir/stacktrace"

	"github.com/NdoleStudio/discusswithai/pkg/services"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/gofiber/fiber/v2"
)

// WhatsappHandler handles nexmo events
type WhatsappHandler struct {
	handler
	logger  telemetry.Logger
	tracer  telemetry.Tracer
	service *services.WhatsappService
}

// NewWhatsappHandler creates a new WhatsappHandler
func NewWhatsappHandler(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	service *services.WhatsappService,
) (h *WhatsappHandler) {
	return &WhatsappHandler{
		logger:  logger.WithService(fmt.Sprintf("%T", h)),
		tracer:  tracer,
		service: service,
	}
}

// RegisterRoutes registers the routes for the NexmoHandler
func (h *WhatsappHandler) RegisterRoutes(app *fiber.App, middlewares ...fiber.Handler) {
	router := app.Group("/v1/whatsapp")
	router.Post("/events", h.computeRoute(middlewares, h.Event)...)
	router.Get("/events", h.computeRoute(middlewares, h.Verify)...)
}

// Verify receives a verification request from the whatsapp API
// @Summary      Receive a verification request from the whatsapp API
// @Description  Receive a new verification request from the whatsapp API
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.NoContent
// @Failure      400  {object}  responses.BadRequest
// @Failure      422  {object}  responses.UnprocessableEntity
// @Failure      500  {object}  responses.InternalServerError
// @Router       /whatsapp/events [get]
func (h *WhatsappHandler) Verify(c *fiber.Ctx) error {
	_, span, _ := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request requests.WhatsappVerifyRequest
	parts := strings.Split(strings.Replace(c.OriginalURL(), "/v1/whatsapp/events?", "", 1), "&")
	for _, part := range parts {
		if strings.Contains(part, "hub.mode=") {
			request.Mode = strings.Replace(part, "hub.mode=", "", 1)
		}
		if strings.Contains(part, "hub.challenge=") {
			request.Challenge = strings.Replace(part, "hub.challenge=", "", 1)
		}
		if strings.Contains(part, "hub.verify_token=") {
			request.VerifyToken = strings.Replace(part, "hub.verify_token=", "", 1)
		}
	}
	return c.Send([]byte(request.Challenge))
}

// Event receives an event from the whatsapp API
// @Summary      Receive an event from the whatsapp API
// @Description  Receive an event from the whatsapp API
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        payload   body requests.NexmoReceiveRequest  true  "Received message request payload"
// @Success      202  {object}  responses.Accepted
// @Failure      400  {object}  responses.BadRequest
// @Failure      422  {object}  responses.UnprocessableEntity
// @Failure      500  {object}  responses.InternalServerError
// @Router       /whatsapp/events [get]
func (h *WhatsappHandler) Event(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request whatsapp.MessageWebhookRequest
	if err := c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall params [%s] into %T", c.OriginalURL(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	messages := request.Entry[0].Changes[0].Value.Messages
	if messages == nil {
		if request.Entry[0].Changes[0].Value.Statuses != nil {
			messageId := (*request.Entry[0].Changes[0].Value.Statuses)[0].ID
			ctxLogger.Info(fmt.Sprintf("request [%s] is a status update for messageID [%s]", request.Entry[0].ID, messageId))
			return h.responseAccepted(c, "Status update processed successfully")
		}
		ctxLogger.Error(stacktrace.NewError(fmt.Sprintf("cannot parse webhook request [%s]", c.Body())))
		return h.responseAccepted(c, "Could not process webhook request. exception swallowed")
	}

	text := ""
	if (*messages)[0].Type == whatsapp.MessageWebhookMessageTypeText {
		text = (*messages)[0].Text.Body
	}

	h.service.Receive(ctx, &services.WhatsappReceiveParams{
		From:        (*messages)[0].From,
		To:          request.Entry[0].Changes[0].Value.Metadata.PhoneNumberID,
		MessageText: text,
		Type:        (*messages)[0].Type,
		MessageID:   (*messages)[0].ID,
	})

	return h.responseAccepted(c, "whatsapp message processed successfully")
}
