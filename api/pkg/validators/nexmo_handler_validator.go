package validators

import (
	"context"
	"fmt"
	"net/url"

	"github.com/NdoleStudio/discusswithai/pkg/requests"
	"github.com/thedevsaddam/govalidator"

	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
)

// NexmoHandlerValidator validates models used in handlers.NexmoHandler
type NexmoHandlerValidator struct {
	govalidator.Validator
	logger telemetry.Logger
	tracer telemetry.Tracer
}

// NewNexmoHandlerValidator creates a new handlers.NexmoHandler validator
func NewNexmoHandlerValidator(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
) (v *NexmoHandlerValidator) {
	return &NexmoHandlerValidator{
		logger: logger.WithService(fmt.Sprintf("%T", v)),
		tracer: tracer,
	}
}

// ValidateReceive checks that an event is coming from Nexmo
func (validator *NexmoHandlerValidator) ValidateReceive(ctx context.Context, request requests.NexmoReceiveRequest) url.Values {
	_, span := validator.tracer.Start(ctx)
	defer span.End()

	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"to": []string{
				"required",
				phoneNumberRule,
			},
			"msisdn": []string{
				"required",
				phoneNumberRule,
			},
			"text": []string{
				"required",
				"min:1",
				"max:1024",
			},
		},
	})

	return v.ValidateStruct()
}
