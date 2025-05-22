# AsyncQ Migration Implementation

This document outlines the migration from `gocron` to `asyncq` for background job processing.

## What Was Implemented

### 1. Standalone Cleanup Functions (`pkg/jobs/cleanup_functions.go`)

- Created reusable cleanup functions that can be called by both gocron and asyncq
- Functions: `PurgeExpiredTokens`, `PurgeDeletedStories`, `PurgeOldStripeWebhookEvents`
- These functions accept `context.Context`, `*sqlx.DB`, and `*logger.Logger` as parameters

### 2. Updated Existing gocron Methods

- Modified existing scheduler methods to call the new standalone functions
- This allows both systems to run in parallel during migration

### 3. AsyncQ Task Definitions (`pkg/tasks/cleanup_tasks.go`)

- Added three new task types:
  - `TypeTokenCleanup` - cleanup expired verification tokens
  - `TypeDeleteStories` - cleanup deleted stories
  - `TypeWebhookCleanup` - cleanup old stripe webhook events
- Each has an enqueue method that adds tasks to the "cleanup" queue

### 4. AsyncQ Handlers (`internal/taskhandlers/cleanup_handlers.go`)

- Created handlers that process the cleanup tasks
- Handlers call the standalone cleanup functions from `pkg/jobs/`
- Include proper error handling and logging

### 5. Updated Worker (`cmd/worker/main.go`)

- Added database configuration and connectivity
- Set up asyncq scheduler for periodic task scheduling
- Registered cleanup handlers with the worker
- Added "cleanup" queue with priority 2
- Configured periodic schedules:
  - Token cleanup: Weekly on Sunday at 20:45
  - Delete stories: Daily at 20:40
  - Webhook cleanup: Weekly on Sunday at 20:50

## Migration Strategy

### Phase 1: Parallel Operation (Current)

- Both gocron (in main app) and asyncq (in worker) are running
- This allows you to verify asyncq is working correctly
- Tasks will run from both systems during this phase

### Phase 2: Remove gocron (Later)

- Once you're confident asyncq is working correctly
- Remove the gocron initialization from `cmd/api/main.go`
- Delete or comment out these lines:
  ```go
  // jobScheduler, err := jobs.NewScheduler(db, log)
  // go jobScheduler.Start()
  // defer jobScheduler.Shutdown()
  ```

## Testing the Implementation

### 1. Start Redis

Make sure Redis is running on localhost:6379

### 2. Run the Worker

```bash
cd cmd/worker
go run main.go
```

You should see:

- Database connection established
- Cleanup scheduler started
- Asynq worker server started
- Periodic tasks registered

### 3. Check Logs

The worker will log when:

- Periodic tasks are scheduled/enqueued
- Tasks are processed by handlers
- Database cleanup operations complete

### 4. Manual Task Testing

You can manually enqueue tasks to test the handlers:

```go
// In your main app or a test script
tasksService.EnqueueTokenCleanup()
tasksService.EnqueueDeleteStories()
tasksService.EnqueueWebhookCleanup()
```

### 5. Monitor with AsyncQMon (Optional)

AsyncQ provides a web UI for monitoring:

```bash
docker run --rm --name asynqmon -p 8080:8080 hibiken/asynqmon
```

Then visit http://localhost:8080

## Configuration

### Worker Configuration

The worker uses environment variables for configuration:

- `APP_DB_HOST`, `APP_DB_PORT`, etc. - Database connection
- `APP_REDIS_HOST`, `APP_REDIS_PORT`, etc. - Redis connection

### Queue Priorities

Current queue configuration in worker:

- `critical`: 6 (60% of worker time)
- `default`: 3 (30% of worker time)
- `low`: 1 (10% of worker time)
- `onboarding`: 5 (25% of worker time)
- `cleanup`: 2 (10% of worker time)

## Benefits of This Migration

1. **No Duplicate Execution**: Only one worker instance handles scheduling
2. **Distributed Processing**: Multiple worker instances can process tasks
3. **Better Monitoring**: AsyncQ provides web UI and CLI tools
4. **Scalability**: Easy to scale workers independently
5. **Reliability**: Tasks are persisted in Redis with automatic retries
6. **Flexibility**: Can easily add new task types and handlers

## Next Steps

1. **Test thoroughly** - Run both systems in parallel and verify tasks execute correctly
2. **Monitor performance** - Check that cleanup tasks complete successfully
3. **Scale if needed** - Add more worker instances if processing is slow
4. **Remove gocron** - Once confident, remove the gocron scheduler from main app
