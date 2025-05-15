package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"
)

// Scheduler manages all background jobs
type Scheduler struct {
	scheduler gocron.Scheduler
	db        *sqlx.DB
	log       *logger.Logger
}

// NewScheduler creates a new job scheduler
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
	s.registerDeleteStoriesJob()
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
			gocron.NewAtTime(0, 0, 0),
		)),
		gocron.NewTask(task),
	)

	if err != nil {
		s.log.Error(ctx, "Failed to register purge deleted stories job", "error", err)
		return
	}

	s.log.Info(ctx, "Registered purge deleted stories job", "jobID", job.ID())
}
