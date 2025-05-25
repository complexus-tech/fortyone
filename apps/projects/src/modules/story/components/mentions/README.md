# Mentions Feature

This directory contains the implementation of the mentions feature for the comment input system.

## How it works

1. **Trigger**: Type `@` followed by a username or display name to trigger the mentions dropdown
2. **Filtering**: The list filters team members based on their display name or username as you type
3. **Navigation**: Use arrow keys (up/down) to navigate through suggestions
4. **Selection**: Press Enter or click on a team member to select them
5. **Styling**: Mentions appear with custom info color styling that matches the design system

## Components

### `list.tsx`

- Main suggestion list component that displays filtered team members
- Handles keyboard navigation (arrow keys, enter, escape)
- Shows user avatars (or initials if no avatar) and usernames
- Implements proper TypeScript types for the suggestion system
- Uses custom color palette from Tailwind config

## Current Implementation

- **Dynamic data**: Fetches real team members using `useTeamMembers(teamId)` hook
- **Type-safe**: Uses proper `Member` type from `@/types` for data mapping
- Integrates with TipTap editor via the Mention extension
- Uses Tippy.js for proper popup positioning
- Supports both light and dark themes
- **Custom styling**: Uses the project's color palette (info color as primary accent)
- **Enhanced UX**: Smooth transitions, border indicators, and improved contrast

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

- **Mentions in editor**: Subtle info color background with border and proper contrast
- **Suggestion dropdown**:
  - Selected items have left border indicator in info color
  - Hover states with subtle info color backgrounds
  - Avatar rings for better visual hierarchy
  - Fallback initials use info color theme

## Future Enhancements

- [ ] Add user roles/titles in the dropdown
- [ ] Implement user search debouncing for better performance
- [ ] Add notification system when users are mentioned
- [ ] Store mention data in comments for proper serialization
- [ ] Add user status indicators (online/offline)
- [ ] Support for team-based filtering
- [ ] Mention analytics and tracking

## Usage

Simply type `@` in any comment input and start typing a team member's name or username. The suggestion dropdown will appear automatically showing all team members that match your query.

**Note**: The component requires a `teamId` prop to fetch the appropriate team members. This prop flows through:

- Story → Activities → Comments → CommentInput
- Story → Activities → CommentInput (standalone)
