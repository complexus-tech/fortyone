package workerbootstrap

import "github.com/complexus-tech/projects-api/pkg/logger"

var (
	workerConfigurationParseFailure = logger.MustDefineError(
		"worker.configuration.parse_failed",
		"Worker environment configuration could not be parsed",
	)
	workerConfigurationInvalidFailure = logger.MustDefineError(
		"worker.configuration.invalid",
		"Worker runtime configuration is invalid",
	)
	workerDatabaseUnavailableFailure = logger.MustDefineError(
		"worker.postgres.unavailable",
		"Worker could not connect to PostgreSQL",
	)
	workerRedisUnavailableFailure = logger.MustDefineError(
		"worker.redis.unavailable",
		"Worker could not connect to Redis",
	)
	workerSchedulerInitializationFailure = logger.MustDefineError(
		"worker.scheduler.initialization_failed",
		"Worker schedules could not be registered",
	)
	workerSlackInitializationFailure = logger.MustDefineError(
		"worker.slack.initialization_failed",
		"Slack event processing could not be initialized",
	)
	workerSlackSigningSecretMissingFailure = logger.MustDefineError(
		"worker.slack.signing_secret_missing",
		"SLACK_SIGNING_SECRET is not configured",
	)
	workerHTTPInitializationFailure = logger.MustDefineError(
		"worker.http.initialization_failed",
		"Worker health endpoints could not be initialized",
	)
	workerHTTPListenFailure = logger.MustDefineError(
		"worker.http.listen_failed",
		"Worker health server could not bind its configured address",
	)
	workerSchedulerStartFailure = logger.MustDefineError(
		"worker.scheduler.start_failed",
		"Worker scheduler could not start",
	)
	workerTaskServerStartFailure = logger.MustDefineError(
		"worker.queue.start_failed",
		"Worker task server could not start",
	)
)
