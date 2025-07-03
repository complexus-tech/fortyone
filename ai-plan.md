# AI Backend Implementation Plan

## Complexus Projects API - Intelligent Insights & Predictions

### 🎯 Overview

Transform the Complexus project management platform by adding intelligent backend AI capabilities that provide predictive insights, automated health scoring, and performance analytics through background processing and existing infrastructure.

## 🏗️ Core AI Features

### 1. Sprint Completion Predictions

- **What**: Predict likelihood of sprint completion based on historical data
- **When**: Daily analysis during active sprints
- **Output**: Completion probability (0-100%), risk factors, recommendations
- **Value**: Early warning system for project managers

### 2. Objective Health Scoring

- **What**: Automatically calculate and update objective health status
- **When**: Weekly analysis of all active objectives
- **Output**: Health scores (At Risk, On Track, Off Track), trend analysis
- **Value**: Proactive objective management and resource allocation

### 3. Team Performance Analytics

- **What**: Analyze team velocity, capacity, and burnout indicators
- **When**: Weekly team performance reviews
- **Output**: Capacity recommendations, burnout alerts, optimization suggestions
- **Value**: Improved team management and workload distribution

### 4. Smart Story Clustering

- **What**: Identify related stories and missing dependencies
- **When**: Nightly background processing
- **Output**: Story relationship suggestions, dependency recommendations
- **Value**: Better sprint planning and reduced blockers

### 5. Intelligent Digest Generation

- **What**: AI-generated weekly/monthly project summaries
- **When**: Weekly on Mondays, Monthly on 1st
- **Output**: Executive summaries, trend insights, action items
- **Value**: Automated reporting for stakeholders

### 6. AI-Powered Story Assignments

- **What**: Intelligent story assignment recommendations and auto-assignment during creation
- **When**: Real-time during story creation, on-demand for existing stories, bulk sprint planning
- **Output**: Optimal assignee suggestions with confidence scores, reasoning, and workload context
- **Value**: Optimal task distribution, reduced PM overhead, improved team utilization, skill-based matching

## 📋 Technical Architecture

### New Dependencies

```go
// Add to go.mod
github.com/sashabaranov/go-openai v1.20.4
```

### Configuration Updates

```go
// Add to cmd/worker/main.go WorkerConfig struct
type WorkerConfig struct {
    // ... existing config fields ...
    AI struct {
        OpenAIAPIKey string `env:"APP_OPENAI_API_KEY"`
        Model        string `default:"gpt-4o-mini" env:"APP_AI_MODEL"`
        Enabled      bool   `default:"true" env:"APP_AI_ENABLED"`
        MaxTokens    int    `default:"2000" env:"APP_AI_MAX_TOKENS"`
    }
    // Update queues to include ai-insights
    Queues map[string]int `default:"critical:6,default:3,low:1,onboarding:5,cleanup:2,notifications:4,automation:3,ai-insights:4"`
}

// Add to cmd/api/main.go Config struct (similar AI struct)
```

### New Package Structure

```
pkg/
  ai/                    # AI service package
    config.go           # AI configuration
    service.go          # Main AI service with OpenAI client
    predictions.go      # Sprint & objective prediction logic
    analytics.go        # Team performance analysis
    clustering.go       # Story relationship analysis
    assignments.go      # Story assignment algorithms
    prompts.go          # AI prompt templates

internal/
  core/
    aiinsights/         # New core domain
      models.go         # Core AI prediction models
      insights.go       # Business logic for AI insights
      service.go        # AI insights service

  handlers/
    aiinsightsgrp/      # New handler group
      insights.go       # API endpoints for AI insights
      models.go         # API response models
      routes.go         # Route definitions

  repo/
    aiinsightsrepo/     # New repository
      models.go         # Database models
      repo.go           # Data access layer
```

### Database Schema

```sql
-- Sprint completion predictions
CREATE TABLE sprint_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id),
    completion_probability DECIMAL(5,2) NOT NULL,
    predicted_completion_date TIMESTAMP,
    confidence_score DECIMAL(5,2),
    risk_factors JSONB,
    recommendations JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(sprint_id, DATE(created_at))
);

-- Objective health predictions
CREATE TABLE objective_health_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    objective_id UUID NOT NULL REFERENCES objectives(id),
    predicted_health TEXT NOT NULL CHECK (predicted_health IN ('At Risk', 'On Track', 'Off Track')),
    health_score DECIMAL(5,2) NOT NULL,
    risk_factors JSONB,
    recommendations JSONB,
    confidence_score DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(objective_id, DATE(created_at))
);

-- Team performance insights
CREATE TABLE team_performance_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    workspace_id UUID NOT NULL,
    insight_type TEXT NOT NULL CHECK (insight_type IN ('velocity', 'capacity', 'burnout_risk', 'efficiency')),
    insight_data JSONB NOT NULL,
    recommendations JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(team_id, insight_type, DATE(created_at))
);

-- User skill profiles (AI-generated)
CREATE TABLE user_skill_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID NOT NULL REFERENCES teams(id),
    skill_data JSONB NOT NULL, -- {"react": 0.85, "backend": 0.92, "ui": 0.45}
    story_completion_stats JSONB NOT NULL, -- average times, success rates by story type
    last_analyzed TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, team_id)
);

-- Assignment recommendations and outcomes
CREATE TABLE assignment_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id),
    recommended_user_id UUID NOT NULL REFERENCES users(id),
    confidence_score DECIMAL(5,2) NOT NULL,
    reasoning TEXT NOT NULL,
    estimated_completion_days DECIMAL(4,2),
    workload_factor JSONB, -- current workload context
    created_at TIMESTAMP DEFAULT NOW()
);

-- Assignment outcomes for learning
CREATE TABLE assignment_outcomes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id),
    recommended_user_id UUID, -- Who AI recommended
    actual_user_id UUID NOT NULL REFERENCES users(id), -- Who actually got assigned
    completion_time_hours DECIMAL(8,2), -- Actual completion time
    user_override BOOLEAN DEFAULT FALSE, -- Did user override AI recommendation
    user_feedback_rating INTEGER CHECK (user_feedback_rating BETWEEN 1 AND 5),
    completion_quality TEXT CHECK (completion_quality IN ('excellent', 'good', 'acceptable', 'poor')),
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 🔄 Background Job Implementation

### New Task Types

```go
// pkg/tasks/aipredict.go
const (
    TypeSprintPrediction      = "ai:predict:sprint_completion"
    TypeObjectiveHealthCheck  = "ai:predict:objective_health"
    TypeTeamInsights         = "ai:analyze:team_performance"
    TypeStoryRelationships   = "ai:analyze:story_relationships"
    TypeDigestGeneration     = "ai:generate:digest"
    TypeUserSkillAnalysis    = "ai:analyze:user_skills"
    TypeBulkAssignment       = "ai:assign:bulk_stories"
)
```

### Scheduled Jobs

```go
// Add to cmd/worker/main.go scheduler

// Daily sprint predictions at 9 AM
_, err = scheduler.Register(
    "0 9 * * *",
    asynq.NewTask(tasks.TypeSprintPrediction, nil),
    asynq.Queue("ai-insights"),
)

// Weekly objective health checks on Monday 8 AM
_, err = scheduler.Register(
    "0 8 * * 1",
    asynq.NewTask(tasks.TypeObjectiveHealthCheck, nil),
    asynq.Queue("ai-insights"),
)

// Weekly user skill analysis on Sunday 11 PM
_, err = scheduler.Register(
    "0 23 * * 0",
    asynq.NewTask(tasks.TypeUserSkillAnalysis, nil),
    asynq.Queue("ai-insights"),
)
```

### Worker Task Handler Registration

```go
// Add to cmd/worker/main.go after existing handler setup
aiHandlers := taskhandlers.NewAIHandlers(log, db, cfg.AI)

// Register AI task handlers in mux
mux.HandleFunc(tasks.TypeSprintPrediction, aiHandlers.HandleSprintPrediction)
mux.HandleFunc(tasks.TypeObjectiveHealthCheck, aiHandlers.HandleObjectiveHealthCheck)
mux.HandleFunc(tasks.TypeTeamInsights, aiHandlers.HandleTeamInsights)
mux.HandleFunc(tasks.TypeUserSkillAnalysis, aiHandlers.HandleUserSkillAnalysis)
mux.HandleFunc(tasks.TypeBulkAssignment, aiHandlers.HandleBulkAssignment)
```

## 🌐 API Endpoints

### Handler Group Registration

```go
// Add to internal/handlers/handlers.go BuildAllRoutes function
aiinsightsgrp.Routes(aiinsightsgrp.Config{
    DB:        cfg.DB,
    Log:       cfg.Log,
    SecretKey: cfg.SecretKey,
    AIService: cfg.AIService, // New AI service dependency
}, app)
```

### AI Insights Handler Group Structure

```go
// internal/handlers/aiinsightsgrp/routes.go
package aiinsightsgrp

type Config struct {
    DB        *sqlx.DB
    Log       *logger.Logger
    SecretKey string
    AIService *ai.Service // New dependency
}

func Routes(cfg Config, app *web.App) {
    aiInsightsService := aiinsights.New(cfg.Log, aiinsightsrepo.New(cfg.Log, cfg.DB), cfg.AIService)
    auth := mid.Auth(cfg.Log, cfg.SecretKey)
    gzip := mid.Gzip(cfg.Log)

    h := New(aiInsightsService, cfg.Log)

    // Sprint Predictions
    app.Get("/workspaces/{workspaceId}/ai/predictions/sprints/{sprintId}", h.GetSprintPrediction, auth, gzip)

    // Objective Health Predictions
    app.Get("/workspaces/{workspaceId}/ai/predictions/objectives/{objectiveId}", h.GetObjectivePrediction, auth, gzip)

    // Team Performance Insights
    app.Get("/workspaces/{workspaceId}/ai/insights/teams/{teamId}", h.GetTeamInsights, auth, gzip)

    // Manual prediction refresh
    app.Post("/workspaces/{workspaceId}/ai/predictions/refresh", h.RefreshPredictions, auth)

    // Story Assignment Endpoints
    app.Post("/workspaces/{workspaceId}/ai/assignments/suggest-realtime", h.SuggestAssignmentRealtime, auth)
    app.Post("/workspaces/{workspaceId}/ai/assignments/stories/{storyId}/suggest", h.SuggestAssignment, auth, gzip)
    app.Post("/workspaces/{workspaceId}/ai/assignments/stories/{storyId}/auto-assign", h.AutoAssignStory, auth)
    app.Post("/workspaces/{workspaceId}/ai/assignments/sprints/{sprintId}/bulk-assign", h.BulkAssignSprint, auth)

    // Team Skills and Workload
    app.Get("/workspaces/{workspaceId}/ai/assignments/teams/{teamId}/skills", h.GetTeamSkills, auth, gzip)

    // Assignment Feedback
    app.Post("/workspaces/{workspaceId}/ai/assignments/feedback", h.SubmitAssignmentFeedback, auth)
}

// Handler methods follow your existing pattern:
// func (h *Handlers) GetSprintPrediction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//     sprintIdParam := web.Params(r, "sprintId")
//     workspaceIdParam := web.Params(r, "workspaceId")
//     // ... parse and validate UUIDs like your existing handlers
//     // ... call service methods
//     // web.Respond(ctx, w, response, http.StatusOK)
//     return nil
// }
```

### Actual API Endpoints (Following Your Workspace Pattern)

```go
// Sprint Predictions
GET /workspaces/{workspaceId}/ai/predictions/sprints/{sprintId}

// Objective Health Predictions
GET /workspaces/{workspaceId}/ai/predictions/objectives/{objectiveId}

// Team Performance Insights
GET /workspaces/{workspaceId}/ai/insights/teams/{teamId}

// Manual Refresh
POST /workspaces/{workspaceId}/ai/predictions/refresh

// Assignment Endpoints
POST /workspaces/{workspaceId}/ai/assignments/suggest-realtime
POST /workspaces/{workspaceId}/ai/assignments/stories/{storyId}/suggest
POST /workspaces/{workspaceId}/ai/assignments/stories/{storyId}/auto-assign
POST /workspaces/{workspaceId}/ai/assignments/sprints/{sprintId}/bulk-assign
GET  /workspaces/{workspaceId}/ai/assignments/teams/{teamId}/skills
POST /workspaces/{workspaceId}/ai/assignments/feedback
```

## 📱 Consumption Methods

### 1. Dashboard Integration (Primary)

- **Sprint Pages**: Completion probability widget with risk indicators
- **Objective Pages**: Health trend charts and recommendations
- **Team Dashboards**: Performance insights and capacity suggestions
- **Stories**: Related story suggestions and dependency alerts
- **Story Creation**: Auto-assign toggle with real-time AI suggestions
- **Sprint Planning**: Bulk assignment recommendations and auto-assign options

### 2. Email Digests (Secondary)

```go
// Weekly AI summary emails using existing Brevo integration
templates/ai/
  weekly-team-insights.html      # Team performance summary
  sprint-risk-alerts.html        # At-risk sprint warnings
  objective-health-report.html   # Objective health updates
  monthly-digest.html            # Executive monthly summary
```

### 3. Real-time Notifications (Alerts)

```go
// Critical AI alerts through existing notification system
// - Sprint completion probability drops below 60%
// - Objective health changes to "At Risk"
// - Team burnout indicators detected
// - High-confidence story dependencies found
```

### 4. AI Chat Enhancement (Future)

```go
// Feed AI insights into existing frontend chat
// - "How likely is Sprint ABC-123 to complete?"
// - "Which objectives need attention this week?"
// - "Show me team performance insights"
// - "Who should I assign this story to?"
// - "Which team member has the lightest workload?"
// - "Auto-assign all unassigned stories in this sprint"
```

## 🚀 Implementation Phases

### Phase 1: Foundation (Week 1-2)

**Goal**: Set up AI infrastructure and basic predictions

**Tasks**:

- [ ] Add OpenAI dependency and configuration
- [ ] Create AI service package (`pkg/ai/`)
- [ ] Create database tables for predictions
- [ ] Implement basic AI service with OpenAI client
- [ ] Create first background task (Sprint Predictions)
- [ ] Add AI insights core domain
- [ ] Create basic API endpoint for sprint predictions

**Deliverables**:

- Working sprint completion predictions
- Database schema for AI insights
- AI insights handler group (`aiinsightsgrp`)
- Basic API endpoint: `GET /workspaces/{workspaceId}/ai/predictions/sprints/{sprintId}`

### Phase 2: Core Predictions (Week 3-4)

**Goal**: Implement all core prediction features

**Tasks**:

- [ ] Implement Objective Health Scoring
- [ ] Add Team Performance Analytics
- [ ] Create comprehensive task handlers
- [ ] Add remaining API endpoints
- [ ] Implement prediction refresh mechanisms
- [ ] Add confidence scoring and risk factors

**Deliverables**:

- All core AI predictions working
- Complete API endpoint coverage
- Scheduled background jobs running
- Performance insights generating

### Phase 2.5: AI Assignments (Week 4-5)

**Goal**: Implement intelligent story assignment capabilities

**Tasks**:

- [ ] Create user skill profiling from historical story data
- [ ] Build assignment recommendation engine
- [ ] Implement real-time assignment API for story creation
- [ ] Add assignment database tables and models
- [ ] Create assignment task handlers for skill analysis
- [ ] Build auto-assignment logic with confidence scoring
- [ ] Add assignment feedback collection system
- [ ] Implement bulk assignment for sprint planning

**Deliverables**:

- Real-time assignment suggestions during story creation
- Auto-assign toggle functionality
- Bulk sprint assignment capabilities
- User skill profiles and workload analysis
- Assignment feedback loop for continuous improvement

### Phase 3: Integration & Notifications (Week 5-6)

**Goal**: Integrate with existing systems and add alerts

**Tasks**:

- [ ] Integrate with existing notification system
- [ ] Create email digest templates
- [ ] Add dashboard widgets/components
- [ ] Implement real-time alerts for critical predictions
- [ ] Add AI insights to existing analytics pages
- [ ] Create user preference settings

**Deliverables**:

- Email digest system
- Dashboard integration
- Real-time notification alerts
- Complete user-facing experience

## 📊 Success Metrics

### Technical Metrics

- **Prediction Accuracy**: >80% for sprint completions, >75% for objective health
- **Assignment Accuracy**: >85% user satisfaction with AI assignment suggestions
- **Performance**: AI tasks complete within 5 minutes
- **Reliability**: 99.5% uptime for AI background jobs
- **Response Time**: API endpoints respond within 200ms, real-time assignments < 500ms

### Business Metrics

- **User Engagement**: 60%+ of teams actively viewing AI insights
- **Assignment Adoption**: 70%+ of new stories use AI assignment suggestions
- **Early Warning**: Identify 90% of at-risk sprints before failure
- **Time Savings**: 2+ hours saved per week per project manager, 30+ minutes saved per day on assignments
- **Decision Quality**: 40% improvement in resource allocation decisions
- **Workload Balance**: 25% improvement in team workload distribution

## 🎯 Quick Start Checklist

### Environment Setup

- [ ] Add `APP_OPENAI_API_KEY` to environment variables
- [ ] Configure AI settings in application config
- [ ] Add OpenAI Go SDK dependency
- [ ] Create AI service configuration

### Database Setup

- [ ] Run database migrations for AI tables
- [ ] Create indexes for performance
- [ ] Verify table relationships and constraints
- [ ] Set up backup strategy for AI data

### Development Setup

- [ ] Create AI service package structure
- [ ] Implement basic OpenAI client wrapper
- [ ] Create first prediction algorithm
- [ ] Set up background job for AI tasks

## 💡 Getting Started

1. **Review this plan** with your team and stakeholders
2. **Set up OpenAI account** and obtain API keys
3. **Begin with Phase 1** - Foundation implementation
4. **Start with Sprint Predictions** as the first feature
5. **Iterate based on user feedback** and accuracy metrics

This plan provides a comprehensive roadmap for implementing intelligent AI features that will transform how teams manage projects, predict outcomes, and optimize performance within the Complexus platform.
