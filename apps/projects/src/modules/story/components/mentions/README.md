# Mentions Feature

This directory contains the implementation of the mentions feature for the comment input system.

## How it works

1. **Trigger**: Type `@` followed by a username or display name to trigger the mentions dropdown
2. **Filtering**: The list filters users based on their display name or username as you type
3. **Navigation**: Use arrow keys (up/down) to navigate through suggestions
4. **Selection**: Press Enter or click on a user to select them
5. **Styling**: Mentions appear with custom info color styling that matches the design system

## Components

### `list.tsx`

- Main suggestion list component that displays filtered users
- Handles keyboard navigation (arrow keys, enter, escape)
- Shows user avatars (or initials if no avatar) and usernames
- Implements proper TypeScript types for the suggestion system
- Uses custom color palette from Tailwind config

## Current Implementation

- Uses static user data (defined in `comment-input.tsx`)
- Integrates with TipTap editor via the Mention extension
- Uses Tippy.js for proper popup positioning
- Supports both light and dark themes
- **Custom styling**: Uses the project's color palette (info color as primary accent)
- **Enhanced UX**: Smooth transitions, border indicators, and improved contrast

## Styling Features

- **Mentions in editor**: Subtle info color background with border and proper contrast
- **Suggestion dropdown**:
  - Selected items have left border indicator in info color
  - Hover states with subtle info color backgrounds
  - Avatar rings for better visual hierarchy
  - Fallback initials use info color theme

## Future Enhancements

- [ ] Connect to real user API instead of static data
- [ ] Add user roles/titles in the dropdown
- [ ] Implement user search debouncing for better performance
- [ ] Add notification system when users are mentioned
- [ ] Store mention data in comments for proper serialization

## Usage

Simply type `@` in any comment input and start typing a user's name or username. The suggestion dropdown will appear automatically.

Example users available:

- John Doe (@johndoe)
- Jane Smith (@janesmith)
- Mike Johnson (@mikejohnson)
- Sarah Wilson (@sarahwilson)
- Tom Brown (@tombrown)
