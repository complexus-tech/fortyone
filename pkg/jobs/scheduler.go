package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"
)

type Scheduler struct {
	scheduler gocron.Scheduler
	db        *sqlx.DB
	log       *logger.Logger
}

func NewScheduler(db *sqlx.DB, log *logger.Logger) (*Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		scheduler: scheduler,
		db:        db,
		log:       log,
	}, nil
}

// Start initializes and starts all jobs
func (s *Scheduler) Start() {
	s.log.Info(context.Background(), "Starting background job scheduler")

	// Register all jobs
	s.registerDeleteStoriesJob()
	s.registerTokenCleanupJob()
	s.registerWebhookCleanupJob()

	s.scheduler.Start()
}

// Shutdown gracefully stops the scheduler
func (s *Scheduler) Shutdown() error {
	s.log.Info(context.Background(), "Shutting down background job scheduler")
	return s.scheduler.Shutdown()
}

// registerDeleteStoriesJob sets up the job to purge stories deleted for 30+ days
func (s *Scheduler) registerDeleteStoriesJob() {
	ctx := context.Background()
	task := func() {
		s.log.Info(ctx, "Running job: Purge deleted stories")
		if err := s.purgeDeletedStories(ctx); err != nil {
			s.log.Error(ctx, "Failed to purge deleted stories", "error", err)
		}
	}

	// Run daily at 12:00 AM
	job, err := s.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(20, 30, 0),
		)),
		gocron.NewTask(task),
	)

	if err != nil {
		s.log.Error(ctx, "Failed to register purge deleted stories job", "error", err)
		return
	}

	s.log.Info(ctx, "Registered purge deleted stories job", "jobID", job.ID())
}

// registerTokenCleanupJob sets up the job to purge verification tokens older than 7 days
func (s *Scheduler) registerTokenCleanupJob() {
	ctx := context.Background()
	task := func() {
		s.log.Info(ctx, "Running job: Purge expired verification tokens")
		if err := s.purgeExpiredTokens(ctx); err != nil {
			s.log.Error(ctx, "Failed to purge expired verification tokens", "error", err)
		}
	}

	// Run weekly at 12:00 AM on Sunday
	job, err := s.scheduler.NewJob(
		gocron.WeeklyJob(
			1,
			gocron.NewWeekdays(time.Sunday),
			gocron.NewAtTimes(gocron.NewAtTime(20, 35, 0)),
		),
		gocron.NewTask(task),
	)

	if err != nil {
		s.log.Error(ctx, "Failed to register token cleanup job", "error", err)
		return
	}

	s.log.Info(ctx, "Registered token cleanup job", "jobID", job.ID())
}

// registerWebhookCleanupJob sets up the job to purge old webhook events
func (s *Scheduler) registerWebhookCleanupJob() {
	ctx := context.Background()
	task := func() {
		s.log.Info(ctx, "Running job: Purge old webhook events")
		if err := s.purgeOldStripeWebhookEvents(ctx); err != nil {
			s.log.Error(ctx, "Failed to purge old webhook events", "error", err)
		}
	}

	// Run weekly at 12:30 AM on Sunday
	job, err := s.scheduler.NewJob(
		gocron.WeeklyJob(
			1,
			gocron.NewWeekdays(time.Sunday),
			gocron.NewAtTimes(gocron.NewAtTime(20, 40, 0)),
		),
		gocron.NewTask(task),
	)

	if err != nil {
		s.log.Error(ctx, "Failed to register webhook cleanup job", "error", err)
		return
	}

	s.log.Info(ctx, "Registered webhook cleanup job", "jobID", job.ID())
}
