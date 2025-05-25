# Mentions Feature

This directory contains the implementation of the mentions feature for the comment input system.

## How it works

1. **Trigger**: Type `@` followed by a username or display name to trigger the mentions dropdown
2. **Filtering**: The list filters team members based on their display name or username as you type
3. **Navigation**: Use arrow keys (up/down) to navigate through suggestions
4. **Selection**: Press Enter or click on a team member to select them
5. **Styling**: Mentions appear with custom info color styling that matches the design system
6. **Extraction**: Extract mentioned users from saved comments for notifications/analytics

## Components

### `list.tsx`

- **Uses Command component**: Built with the same `Command` component from your UI library (consistent with assignees menu)
- Handles keyboard navigation (arrow keys, enter, escape)
- Shows user avatars and usernames with proper Typography components
- Implements proper TypeScript types for the suggestion system
- **Design system consistency**: Matches the styling of other Command-based components

## Current Implementation

- **Dynamic data**: Fetches real team members using `useTeamMembers(teamId)` hook
- **Type-safe**: Uses proper `Member` type from `@/types` for data mapping
- **Design system integration**: Uses `Command`, `Avatar`, `Text`, and `Flex` components from UI library
- Integrates with TipTap editor via the Mention extension
- Uses Tippy.js for proper popup positioning
- Supports both light and dark themes
- **Custom styling**: Uses the project's color palette (info color as primary accent)
- **Enhanced UX**: Smooth transitions, border indicators, and improved contrast
- **Data attributes**: Mentions include `data-id` and `data-label` for extraction

## Extracting Mentioned Users

### Using the `useMentions` Hook

```typescript
import { useMentions } from "@/lib/hooks/use-mentions";

const MyComponent = ({ comment }: { comment: Comment }) => {
  const {
    mentions,           // Array of { id, label } objects
    mentionedUserIds,   // Array of user IDs
    hasMentions,        // Boolean - true if any mentions exist
    mentionCount,       // Number of mentions
    isUserMentioned     // Function to check if specific user is mentioned
  } = useMentions(comment.content);

  // Check if current user is mentioned
  const currentUserMentioned = isUserMentioned(session?.user?.id);

  // Get all mentioned user IDs for notifications
  const userIdsToNotify = mentionedUserIds;

  return (
    <div>
      {hasMentions && (
        <p>This comment mentions {mentionCount} users</p>
      )}
      {currentUserMentioned && (
        <Badge>You were mentioned</Badge>
      )}
    </div>
  );
};
```

### Using Utility Functions Directly

```typescript
import {
  extractMentionsFromHTML,
  extractMentionedUserIds,
  isUserMentioned,
} from "@/lib/utils/mentions";

// Extract all mention data
const mentions = extractMentionsFromHTML(comment.content);
// [{ id: "user123", label: "John Doe" }, { id: "user456", label: "Jane Smith" }]

// Extract just user IDs
const userIds = extractMentionedUserIds(comment.content);
// ["user123", "user456"]

// Check if specific user is mentioned
const mentioned = isUserMentioned(comment.content, "user123");
// true
```

## Use Cases

### 1. **Notifications**

```typescript
const handleCommentSubmit = async (commentContent: string) => {
  // Save comment
  const comment = await saveComment(commentContent);

  // Extract mentioned users
  const mentionedUserIds = extractMentionedUserIds(commentContent);

  // Send notifications
  if (mentionedUserIds.length > 0) {
    await sendMentionNotifications(mentionedUserIds, comment.id);
  }
};
```

### 2. **Comment Analytics**

```typescript
const CommentStats = ({ comment }: { comment: Comment }) => {
  const { mentionCount, mentions } = useMentions(comment.content);

  return (
    <div>
      <Text>Mentions: {mentionCount}</Text>
      {mentions.map(mention => (
        <Badge key={mention.id}>{mention.label}</Badge>
      ))}
    </div>
  );
};
```

### 3. **User Mention History**

```typescript
const getUserMentions = (comments: Comment[], userId: string) => {
  return comments.filter((comment) => isUserMentioned(comment.content, userId));
};
```

## Data Mapping

The component maps team member data using the proper `Member` type:

```typescript
type Member = {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  fullName: string;
  avatarUrl: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};
```

**Mapping to MentionItem:**

- **ID**: `member.id`
- **Label**: `member.fullName`
- **Username**: `member.username`
- **Avatar**: `member.avatarUrl`

## Integration Points

- **CommentInput**: Requires `teamId` prop to fetch team members
- **Comments**: Updated to pass `teamId` to nested CommentInput components
- **Activities**: Passes `teamId` from story context to both Comments and CommentInput

## Styling Features

- **Mentions in editor**: Subtle background with border and proper contrast using design system colors
- **Suggestion dropdown**:
  - Uses `Command.List` with proper styling from design system
  - `Command.Item` with active states and hover effects
  - `Avatar` component with consistent sizing
  - `Text` components with proper color variants
  - Matches the look and feel of assignees menu and other Command-based components

## Future Enhancements

- [ ] Add user roles/titles in the dropdown
- [ ] Implement user search debouncing for better performance
- [x] Add notification system when users are mentioned
- [x] Store mention data in comments for proper serialization
- [ ] Add user status indicators (online/offline)
- [ ] Support for team-based filtering
- [ ] Mention analytics and tracking

## Usage

Simply type `@` in any comment input and start typing a team member's name or username. The suggestion dropdown will appear automatically showing all team members that match your query.

**Note**: The component requires a `teamId` prop to fetch the appropriate team members. This prop flows through:

- Story → Activities → Comments → CommentInput
- Story → Activities → CommentInput (standalone)
