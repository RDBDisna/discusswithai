package di

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NdoleStudio/discusswithai/pkg/whatsapp"

	cloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/NdoleStudio/discusswithai/pkg/cache"
	"github.com/NdoleStudio/discusswithai/pkg/entities"
	"github.com/NdoleStudio/discusswithai/pkg/handlers"
	"github.com/NdoleStudio/discusswithai/pkg/middlewares"
	"github.com/NdoleStudio/discusswithai/pkg/nexmo"
	"github.com/NdoleStudio/discusswithai/pkg/services"
	"github.com/NdoleStudio/discusswithai/pkg/telemetry"
	"github.com/NdoleStudio/discusswithai/pkg/validators"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/hirosassa/zerodriver"
	"github.com/palantir/stacktrace"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	openapi "github.com/sashabaranov/go-openai"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"gorm.io/driver/postgres"
	gormLogger "gorm.io/gorm/logger"

	"github.com/jinzhu/now"
	"gorm.io/gorm"
)

// Container is used to resolve services at runtime
type Container struct {
	projectID string
	db        *gorm.DB
	app       *fiber.App
	logger    telemetry.Logger
}

// NewContainer creates a new dependency injection container
func NewContainer(version string, projectID string) (container *Container) {
	// Set location to UTC
	now.DefaultConfig = &now.Config{
		TimeLocation: time.UTC,
	}

	container = &Container{
		projectID: projectID,
		logger:    logger(3).WithService(fmt.Sprintf("%T", container)),
	}

	container.InitializeTraceProvider(version, os.Getenv("GCP_PROJECT_ID"))

	container.RegisterNexmoRoutes()
	container.RegisterWhatsappRoutes()

	// this has to be last since it registers the /* route
	container.RegisterSwaggerRoutes()

	return container
}

// RegisterSwaggerRoutes registers routes for swagger

func (container *Container) RegisterSwaggerRoutes() {
	container.logger.Debug("registering swagger routes")
	container.App().Get("/*", swagger.HandlerDefault)
}

// RegisterNexmoRoutes registers routes for the /v1/nexmo prefix
func (container *Container) RegisterNexmoRoutes() {
	container.logger.Debug(fmt.Sprintf("registering %T routes", &handlers.NexmoHandler{}))
	container.NexmoHandler().RegisterRoutes(container.App())
}

// RegisterWhatsappRoutes registers routes for the /v1/whatsapp prefix
func (container *Container) RegisterWhatsappRoutes() {
	container.logger.Debug(fmt.Sprintf("registering %T routes", &handlers.NexmoHandler{}))
	container.WhatsappHandler().RegisterRoutes(container.App())
}

// NexmoHandlerValidator creates a new instance of validators.NexmoHandlerValidator
func (container *Container) NexmoHandlerValidator() (validator *validators.NexmoHandlerValidator) {
	container.logger.Debug(fmt.Sprintf("creating %T", validator))
	return validators.NewNexmoHandlerValidator(
		container.Logger(),
		container.Tracer(),
	)
}

// WhatsappHandler creates a new instance of handlers.WhatsappHandler
func (container *Container) WhatsappHandler() (handler *handlers.WhatsappHandler) {
	container.logger.Debug(fmt.Sprintf("creating %T", handler))
	return handlers.NewWhatsappHandler(
		container.Logger(),
		container.Tracer(),
		container.WhatsappService(),
	)
}

// NexmoHandler creates a new instance of handlers.NexmoHandler
func (container *Container) NexmoHandler() (handler *handlers.NexmoHandler) {
	container.logger.Debug(fmt.Sprintf("creating %T", handler))

	return handlers.NewNexmoHandler(
		container.Logger(),
		container.Tracer(),
		container.NexmoService(),
		container.NexmoHandlerValidator(),
	)
}

// WhatsappClient creates a new instance of whatsapp.Client
func (container *Container) WhatsappClient() (service *whatsapp.Client) {
	container.logger.Debug(fmt.Sprintf("creating %T", service))
	return whatsapp.New(
		whatsapp.WithHTTPClient(container.HTTPClient("whatsapp")),
		whatsapp.WithAccessToken(os.Getenv("WHATSAPP_ACCESS_TOKEN")),
	)
}

// NexmoClient creates a new instance of nexmo.Client
func (container *Container) NexmoClient() (service *nexmo.Client) {
	container.logger.Debug(fmt.Sprintf("creating %T", service))
	return nexmo.New(
		nexmo.WithHTTPClient(container.HTTPClient("nexmo")),
		nexmo.WithAPIKey(os.Getenv("NEXMO_API_KEY")),
		nexmo.WithAPISecret(os.Getenv("NEXMO_API_SECRET")),
	)
}

// OpenAPIClient creates a new instance of openapi.Client
func (container *Container) OpenAPIClient() (service *openapi.Client) {
	container.logger.Debug(fmt.Sprintf("creating %T", service))
	return openapi.NewClient(os.Getenv("OPENAPI_AUTH_TOKEN"))
}

// OpenAPIService creates a new instance of services.OpenAPIService
func (container *Container) OpenAPIService() (service *services.OpenAPIService) {
	container.logger.Debug(fmt.Sprintf("creating %T", service))
	return services.NewOpenAPIService(
		container.Logger(),
		container.Tracer(),
		container.OpenAPIClient(),
	)
}

// WhatsappService creates a new instance of services.WhatsappService
func (container *Container) WhatsappService() (service *services.WhatsappService) {
	container.logger.Debug(fmt.Sprintf("creating %T", service))
	return services.NewWhatsappService(
		container.Logger(),
		container.Tracer(),
		container.WhatsappClient(),
		container.OpenAPIService(),
	)
}

// NexmoService creates a new instance of services.NexmoService
func (container *Container) NexmoService() (service *services.NexmoService) {
	container.logger.Debug(fmt.Sprintf("creating %T", service))
	return services.NewNexmoService(
		container.Logger(),
		container.Tracer(),
		container.NexmoClient(),
		container.Cache(),
		container.OpenAPIService(),
	)
}

// HTTPClient creates a new http.Client
func (container *Container) HTTPClient(name string) *http.Client {
	container.logger.Debug(fmt.Sprintf("creating %s %T", name, http.DefaultClient))
	return &http.Client{
		Timeout: 60 * time.Second,
	}
}

//// HTTPRoundTripper creates an open telemetry http.RoundTripper
//func (container *Container) HTTPRoundTripper(name string) http.RoundTripper {
//	container.logger.Debug(fmt.Sprintf("Debug: initializing %s %T", name, http.DefaultTransport))
//	return otelroundtripper.New(
//		otelroundtripper.WithName(name),
//		otelroundtripper.WithParent(container.RetryHTTPRoundTripper()),
//		otelroundtripper.WithMeter(global.Meter(container.projectID)),
//		otelroundtripper.WithAttributes(container.OtelResources(container.version, container.projectID).Attributes()...),
//	)
//}

// App creates a new instance of fiber.App
func (container *Container) App() (app *fiber.App) {
	if container.app != nil {
		return container.app
	}

	container.logger.Debug(fmt.Sprintf("creating %T", app))

	app = fiber.New()

	if isLocal() {
		app.Use(fiberLogger.New())
	}
	app.Use(
		middlewares.OtelTraceContext(
			container.Tracer(),
			container.Logger(),
			"X-Cloud-Trace-Context",
			os.Getenv("GCP_PROJECT_ID"),
		),
	)
	app.Use(cors.New())

	container.app = app

	return app
}

// InitializeOtelResources initializes open telemetry resources
func (container *Container) InitializeOtelResources(version string, namespace string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(namespace),
		semconv.ServiceNamespaceKey.String(namespace),
		semconv.ServiceVersionKey.String(version),
		semconv.ServiceInstanceIDKey.String(hostName()),
		attribute.String("service.environment", os.Getenv("ENV")),
	)
}

// Cache creates a new instance of cache.Cache
func (container *Container) Cache() cache.Cache {
	container.logger.Debug("creating cache.Cache")
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		container.logger.Fatal(stacktrace.Propagate(err, fmt.Sprintf("cannot parse redis url [%s]", os.Getenv("REDIS_URL"))))
	}
	opt.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	return cache.NewRedisCache(container.Tracer(), redis.NewClient(opt))
}

// Logger creates a new instance of telemetry.Logger
func (container *Container) Logger(skipFrameCount ...int) telemetry.Logger {
	container.logger.Debug("creating telemetry.Logger")
	if len(skipFrameCount) > 0 {
		return logger(skipFrameCount[0])
	}
	return logger(3)
}

// GormLogger creates a new instance of gormLogger.Interface
func (container *Container) GormLogger() gormLogger.Interface {
	container.logger.Debug("creating gormLogger.Interface")
	return telemetry.NewGormLogger(
		container.Tracer(),
		container.Logger(6),
	)
}

// DB creates an instance of gorm.DB if it has not been created already
func (container *Container) DB() (db *gorm.DB) {
	if container.db != nil {
		return container.db
	}

	container.logger.Debug(fmt.Sprintf("creating %T", db))

	config := &gorm.Config{}
	if isLocal() {
		config = &gorm.Config{Logger: container.GormLogger()}
	}

	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_DSN")), config)
	if err != nil {
		container.logger.Fatal(err)
	}
	container.db = db

	container.logger.Debug(fmt.Sprintf("Running migrations for %T", db))

	if err = db.AutoMigrate(&entities.Message{}); err != nil {
		container.logger.Fatal(stacktrace.Propagate(err, fmt.Sprintf("cannot migrate %T", &entities.Message{})))
	}

	return container.db
}

// Tracer creates a new instance of telemetry.Tracer
func (container *Container) Tracer() (t telemetry.Tracer) {
	container.logger.Debug("creating telemetry.Tracer")
	return telemetry.NewOtelLogger(
		container.projectID,
		container.Logger(),
	)
}

// InitializeTraceProvider initializes the open telemetry trace provider
func (container *Container) InitializeTraceProvider(version string, namespace string) func() {
	if isLocal() {
		return container.initializeUptraceProvider(version, namespace)
	}
	return container.initializeGoogleTraceProvider(version, namespace)
}

func (container *Container) initializeGoogleTraceProvider(version string, namespace string) func() {
	container.logger.Debug("initializing google trace provider")

	exporter, err := cloudtrace.New(cloudtrace.WithProjectID(os.Getenv("GCP_PROJECT_ID")))
	if err != nil {
		container.logger.Fatal(stacktrace.Propagate(err, "cannot create cloud trace exporter"))
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(container.InitializeOtelResources(version, namespace)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		_ = exporter.Shutdown(context.Background())
	}
}

func (container *Container) initializeUptraceProvider(version string, namespace string) (flush func()) {
	container.logger.Debug("initializing uptrace provider")
	// Configure OpenTelemetry with sensible defaults.
	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(os.Getenv("UPTRACE_DSN")),
		uptrace.WithServiceName(namespace),
		uptrace.WithServiceVersion(version),
	)

	// Send buffered spans and free resources.
	return func() {
		err := uptrace.Shutdown(context.Background())
		if err != nil {
			container.logger.Error(err)
		}
	}
}

func logger(skipFrameCount int) telemetry.Logger {
	fields := map[string]string{
		"pid":      strconv.Itoa(os.Getpid()),
		"hostname": hostName(),
	}

	return telemetry.NewZerologLogger(
		os.Getenv("GCP_PROJECT_ID"),
		fields,
		logDriver(skipFrameCount),
		nil,
	)
}

func logDriver(skipFrameCount int) *zerodriver.Logger {
	if isLocal() {
		return consoleLogger(skipFrameCount)
	}
	return jsonLogger(skipFrameCount)
}

func jsonLogger(skipFrameCount int) *zerodriver.Logger {
	logLevel := zerolog.DebugLevel
	zerolog.SetGlobalLevel(logLevel)

	// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
	logLevelSeverity := map[zerolog.Level]string{
		zerolog.TraceLevel: "DEFAULT",
		zerolog.DebugLevel: "DEBUG",
		zerolog.InfoLevel:  "INFO",
		zerolog.WarnLevel:  "WARNING",
		zerolog.ErrorLevel: "ERROR",
		zerolog.PanicLevel: "CRITICAL",
		zerolog.FatalLevel: "CRITICAL",
	}

	zerolog.LevelFieldName = "severity"
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		return logLevelSeverity[l]
	}
	zerolog.TimestampFieldName = "time"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	zl := zerolog.New(os.Stderr).With().Timestamp().CallerWithSkipFrameCount(skipFrameCount).Logger()
	return &zerodriver.Logger{Logger: &zl}
}

func hostName() string {
	h, err := os.Hostname()
	if err != nil {
		h = strconv.Itoa(os.Getpid())
	}
	return h
}

func consoleLogger(skipFrameCount int) *zerodriver.Logger {
	l := zerolog.New(
		zerolog.ConsoleWriter{
			Out: os.Stderr,
		}).With().Timestamp().CallerWithSkipFrameCount(skipFrameCount).Logger()
	return &zerodriver.Logger{
		Logger: &l,
	}
}

func isLocal() bool {
	return os.Getenv("ENV") == "local"
}
