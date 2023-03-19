package handlers

import (
	"fmt"

	"github.com/NdoleStudio/discusswithai/pkg/requests"
	"github.com/NdoleStudio/discusswithai/pkg/services"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/NdoleStudio/discusswithai/pkg/validators"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v2"
	"github.com/palantir/stacktrace"
)

// NexmoHandler handles nexmo events
type NexmoHandler struct {
	handler
	logger    telemetry.Logger
	tracer    telemetry.Tracer
	service   *services.NexmoService
	validator *validators.NexmoHandlerValidator
}

// NewNexmoHandler creates a new NexmoHandler
func NewNexmoHandler(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
	service *services.NexmoService,
	validator *validators.NexmoHandlerValidator,
) (h *NexmoHandler) {
	return &NexmoHandler{
		logger:    logger.WithService(fmt.Sprintf("%T", h)),
		tracer:    tracer,
		service:   service,
		validator: validator,
	}
}

// RegisterRoutes registers the routes for the NexmoHandler
func (h *NexmoHandler) RegisterRoutes(app *fiber.App, middlewares ...fiber.Handler) {
	router := app.Group("/v1/nexmo")
	router.Post("/receive", h.computeRoute(middlewares, h.Receive)...)
}

// Receive receives a new SMS message from the Nexmo API
// @Summary      Receive a new SMS message from the nexmo API
// @Description  Add a new message received from the nexmo API
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        payload   body requests.NexmoReceiveRequest  true  "Received message request payload"
// @Success      202  {object}  responses.Accepted
// @Failure      400  {object}  responses.BadRequest
// @Failure      422  {object}  responses.UnprocessableEntity
// @Failure      500  {object}  responses.InternalServerError
// @Router       /nexmo/receive [post]
func (h *NexmoHandler) Receive(c *fiber.Ctx) error {
	ctx, span, ctxLogger := h.tracer.StartFromFiberCtxWithLogger(c, h.logger)
	defer span.End()

	var request requests.NexmoReceiveRequest
	if err := c.BodyParser(&request); err != nil {
		msg := fmt.Sprintf("cannot marshall [%s] into %T", c.Body(), request)
		ctxLogger.Warn(stacktrace.Propagate(err, msg))
		return h.responseBadRequest(c, err)
	}

	if errors := h.validator.ValidateReceive(ctx, request.Sanitize()); len(errors) != 0 {
		msg := fmt.Sprintf("validation errors [%s], while receiving message from nexmo [%s]", spew.Sdump(errors), c.Body())
		ctxLogger.Warn(stacktrace.NewError(msg))
		return h.responseUnprocessableEntity(c, errors, "validation errors while receiving message")
	}

	h.service.Receive(ctx, request.ToReceiveParams())

	return h.responseAccepted(c, "message received successfully")
}
