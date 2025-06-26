# AI Tools Architecture

## Overview

The AI assistant now uses a tools-based architecture with the latest AI SDK, providing comprehensive project management capabilities beyond just navigation. This architecture allows Maya to perform specific actions and provide detailed insights about your projects.

## Architecture Components

### 1. Tools System

Located in `src/lib/ai/tools/`, each tool provides specific functionality:

#### Navigation Tool (`navigation.ts`)

- **Purpose**: Handle navigation to different application sections
- **Actions**: Navigate to my-work, summary, analytics, objectives, sprints, etc.
- **Usage**: "Go to my stories", "Show me the dashboard"

#### Stories Tool (`stories.ts`)

- **Purpose**: Manage and query stories/tasks
- **Actions**:
  - `list-assigned`: Show stories assigned to current user
  - `list-created`: Show stories created by current user
  - `list-all`: Show all stories in current view
  - `get-details`: Get detailed information about a specific story
  - `create`: Create a new story
  - `search`: Search stories by query
- **Usage**: "Show me my assigned stories", "Create a new story", "Search for authentication stories"

#### Sprints Tool (`sprints.ts`)

- **Purpose**: Manage sprints and provide sprint insights
- **Actions**:
  - `list-active`: Show currently active sprints
  - `list-all`: Show all sprints
  - `get-summary`: Get detailed sprint summary with progress
  - `get-burndown`: Get burndown chart data
  - `get-velocity`: Get team velocity data
  - `plan-next`: Get recommendations for next sprint planning
  - `get-details`: Get detailed sprint information
- **Usage**: "Get current sprint summary", "What's the team velocity?", "Show me the burndown chart"

### 2. API Integration

- **Route**: `/api/chat`
- **Framework**: Latest AI SDK with tools support
- **Model**: OpenAI GPT-4o-mini
- **Features**: Streaming responses, tool invocation, error handling

### 3. Chat Component

- **Location**: `src/components/ui/chat/index.tsx`
- **Features**:
  - Real-time chat interface
  - Tool invocation handling
  - Automatic navigation after tool usage
  - Message persistence

## Usage Examples

### Navigation

```
User: "Go to my stories"
AI: "I'll take you to your work page where you can see all your assigned stories."
→ Navigates to /my-work
```

### Story Management

```
User: "Show me my assigned stories"
AI: "Here are the stories assigned to you:
- Implement user authentication (in-progress, high priority)
- Design dashboard layout (todo, medium priority)"
```

### Sprint Insights

```
User: "Get current sprint summary"
AI: "Here's the summary for Sprint 15 - User Authentication:
- Progress: 65% complete
- Remaining points: 7 out of 21
- Days remaining: 3
- Current velocity: 2.3 points/day"
```

### Story Creation

```
User: "Create a new story for implementing user preferences"
AI: "I'll help you create a new story. What would you like to title it?"
→ Guides through story creation process
```

## Technical Implementation

### Tool Structure

Each tool follows this pattern:

```typescript
export const toolName = {
  name: "tool-name",
  description: "Tool description",
  parameters: z.object({
    // Zod schema for parameters
  }),
  execute: ({ params }) => {
    // Tool logic
    return { success: true, data: result };
  },
};
```

### Integration Points

1. **API Route**: Registers tools with AI SDK
2. **Chat Component**: Handles tool invocations and responses
3. **Navigation**: Automatic routing based on tool results
4. **Error Handling**: Graceful fallbacks for tool failures

## Future Enhancements

### Planned Tools

1. **Objectives Tool**: Manage OKRs and strategic goals
2. **Analytics Tool**: Generate reports and insights
3. **Team Tool**: Manage team members and permissions
4. **Notifications Tool**: Handle alerts and updates

### Advanced Features

1. **Context Awareness**: Tools that consider user's current location
2. **Permission Checking**: Tools that respect user roles
3. **Real-time Data**: Integration with actual application data
4. **Custom Actions**: Workspace-specific tool configurations

## Setup and Configuration

### Environment Variables

```bash
OPENAI_API_KEY=your_openai_api_key_here
```

### Dependencies

- `ai` (Vercel AI SDK) - Latest version
- `@ai-sdk/openai` - OpenAI provider
- `@ai-sdk/react` - React hooks
- `zod` - Schema validation

### Adding New Tools

1. Create tool file in `src/lib/ai/tools/`
2. Define tool structure with name, description, parameters, and execute function
3. Export from `src/lib/ai/tools/index.ts`
4. Update system prompt in API route if needed

## Testing

### Tool Testing

```bash
# Test navigation
"Go to my stories"

# Test story management
"Show me my assigned stories"
"Create a new story"

# Test sprint insights
"Get current sprint summary"
"What's the team velocity?"
```

### Integration Testing

1. Verify tool invocation in chat
2. Check navigation after tool usage
3. Validate error handling
4. Test with different user inputs

## Troubleshooting

### Common Issues

1. **Tool not invoked**: Check tool name and parameters in system prompt
2. **Navigation not working**: Verify tool result structure
3. **API errors**: Check OpenAI API key and rate limits
4. **Type errors**: Ensure tool parameters match Zod schema

### Debug Mode

Enable debug logging to see tool invocations and responses in browser console.

## Performance Considerations

1. **Tool Execution**: Tools are executed server-side for security
2. **Caching**: Consider caching for frequently accessed data
3. **Rate Limiting**: Implement rate limiting for tool usage
4. **Error Recovery**: Graceful fallbacks for tool failures

This architecture provides a solid foundation for expanding Maya's capabilities while maintaining clean separation of concerns and easy extensibility.
