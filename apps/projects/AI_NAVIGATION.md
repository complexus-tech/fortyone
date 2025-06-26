# AI Navigation Feature

## Overview

The AI Navigation feature allows users to navigate through the Complexus application using natural language commands through the Maya AI assistant. Users can ask Maya to take them to different parts of the application, and she will intelligently route them to the appropriate pages.

## How It Works

### 1. Chat Interface

- Users can access Maya through the "Ask Maya" button in the bottom-right corner of the application
- The chat interface provides a conversational way to interact with the AI

### 2. Navigation Detection

The system detects navigation intent through two methods:

#### Direct Pattern Matching

- **My Work**: "my work", "my stories", "my tasks", "assigned to me"
- **Summary**: "summary", "dashboard", "overview", "home"
- **Analytics**: "analytics", "reports", "metrics", "insights"
- **Objectives**: "objectives", "okrs", "goals", "key results"
- **Sprints**: "sprints", "sprint board", "active sprint"
- **Notifications**: "notifications", "inbox", "alerts"
- **Settings**: "settings", "preferences", "configuration"
- **Teams**: "teams", "team management", "team members"
- **Roadmaps**: "roadmap", "roadmaps", "product roadmap"

#### Command Patterns

- "Go to [destination]"
- "Navigate to [destination]"
- "Take me to [destination]"
- "Show me [destination]"
- "Open [destination]"

### 3. Available Routes

- `/my-work` - My Stories/Work
- `/summary` - Dashboard/Summary
- `/analytics` - Analytics page
- `/objectives` - Objectives/OKRs
- `/sprints` - Sprint management
- `/notifications` - Notifications/Inbox
- `/settings` - Settings
- `/teams` - Team management
- `/roadmaps` - Roadmap view

## Usage Examples

### Direct Navigation

- "my stories" → Navigates to `/my-work`
- "dashboard" → Navigates to `/summary`
- "analytics" → Navigates to `/analytics`

### Command-Based Navigation

- "Go to my stories" → Navigates to `/my-work`
- "Take me to the dashboard" → Navigates to `/summary`
- "Show me analytics" → Navigates to `/analytics`
- "Navigate to objectives" → Navigates to `/objectives`

## Technical Implementation

### API Route

- **Location**: `/api/chat`
- **Method**: POST
- **Features**:
  - Uses Vercel AI SDK with OpenAI GPT-4o-mini
  - Detects navigation intent from user messages
  - Provides intelligent responses with navigation context
  - Includes fallback handling for API errors

### Chat Component

- **Location**: `src/components/ui/chat/index.tsx`
- **Features**:
  - Real-time chat interface
  - Navigation intent detection
  - Automatic routing based on detected intent
  - Suggested prompts for common navigation requests

### Error Handling

- Graceful fallback when AI service is unavailable
- Clear error messages for users
- Continues to work with basic navigation even without AI

## Future Enhancements

1. **Context-Aware Navigation**: Consider user's current location and permissions
2. **Smart Suggestions**: AI-powered suggestions based on user behavior
3. **Voice Commands**: Voice-to-text integration for hands-free navigation
4. **Custom Routes**: Allow workspace-specific navigation shortcuts
5. **Analytics Integration**: Track navigation patterns for insights

## Setup Requirements

### Environment Variables

The feature requires an OpenAI API key to be configured:

```
OPENAI_API_KEY=your_openai_api_key_here
```

### Dependencies

- `ai` (Vercel AI SDK) - Already installed
- `@ai-sdk/openai` - Already installed
- `@ai-sdk/react` - Already installed

## Testing

To test the navigation feature:

1. Open the application
2. Click the "Ask Maya" button in the bottom-right corner
3. Try navigation commands like:
   - "Go to my stories"
   - "Show me the dashboard"
   - "Take me to analytics"
4. Verify that the application navigates to the correct page

## Troubleshooting

### Common Issues

1. **Navigation not working**: Check if OpenAI API key is configured
2. **Slow responses**: Verify internet connection and API service status
3. **Wrong route**: Check the navigation mapping in the code

### Debug Mode

Enable debug logging by checking the browser console for navigation detection logs.
