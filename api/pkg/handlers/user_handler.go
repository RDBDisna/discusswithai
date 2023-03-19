package handlers

//
//import (
//	"fmt"
//	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
//	"github.com/NdoleStudio/discusswithai/pkg/validators"
//	"github.com/palantir/stacktrace"
//)
//
//// UserHandler handles user http requests.
//type UserHandler struct {
//	handler
//	logger    telemetry.Logger
//	tracer    telemetry.Tracer
//	validator *validators.UserHandlerValidator
//}
//
//// NewUserHandler creates a new UserHandler
//func NewUserHandler(
//	logger telemetry.Logger,
//	tracer telemetry.Tracer,
//	validator *validators.UserHandlerValidator,
//) (h *UserHandler) {
//	return &UserHandler{
//		logger:    logger.WithService(fmt.Sprintf("%T", h)),
//		tracer:    tracer,
//		validator: validator,
//	}
//}
//
//// RegisterRoutes registers the routes for the MessageHandler
//func (h *UserHandler) RegisterRoutes(app *fiber.App, middlewares []fiber.Handler) {
//	router := app.Group("/v1/users")
//	router.Get("/me", h.computeRoute(middlewares, h.me)...)
//}
//
//// me returns the currently authenticated entities.User
//// @Summary      Currently authenticated user
//// @Description  Fetches the currently authenticated user. This method creates the user if one doesn't exist
//// @Security	 BearerAuth
//// @Tags         Users
//// @Produce      json
//// @Failure      400		{object}	responses.BadRequest
//// @Failure 	 401    	{object}	responses.Unauthorized
//// @Failure      422		{object}	responses.UnprocessableEntity
//// @Failure      500		{object}	responses.InternalServerError
//// @Router       /users/me 	[get]
//func (h *UserHandler) me(c *fiber.Ctx) error {
//	ctx, span := h.tracer.StartFromFiberCtx(c)
//	defer span.End()
//
//	ctxLogger := h.tracer.CtxLogger(h.logger, span)
//
//	authUser := h.userFromContext(c)
//
//	user, err := h.service.Get(ctx, c.OriginalURL(), authUser)
//	if err != nil {
//		msg := fmt.Sprintf("cannot get user with ID [%s]", authUser.ID)
//		ctxLogger.Error(stacktrace.Propagate(err, msg))
//		return h.responseInternalServerError(c)
//	}
//
//	return h.responseOK(c, "user fetched successfully", user)
//}
