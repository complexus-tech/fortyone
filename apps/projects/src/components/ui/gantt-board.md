# Gantt Board Component Documentation

## Overview

The Gantt Board component is a fully interactive Gantt chart implementation for project management with stories/tasks. It supports multiple zoom levels, drag-and-drop functionality, and real-time updates.

## Architecture

### Component Structure

```
GanttBoard (main container)
├── Stories (left sidebar with story details)
│   ├── Header (zoom controls, today button)
│   └── Story rows (title, assignee, priority, status, duration)
└── Chart (right side with timeline and gantt bars)
    ├── TimelineHeader (date columns, periods)
    └── Story bars (interactive drag/resize bars)
```

### File Location

```
apps/projects/src/components/ui/gantt-board.tsx
```

## Key Features

### 1. Multi-Zoom Timeline

- **Weeks view**: Daily columns with week grouping
- **Months view**: Monthly columns with start/end days
- **Quarters view**: Quarterly columns with start/end months

### 2. Interactive Gantt Bars

- Drag to move entire date range
- Resize handles to adjust start/end dates
- Real-time visual feedback during drag
- Optimistic updates with API sync

### 3. Story Management

- Click story titles to navigate to detail page
- Change priority, status, assignee via dropdowns
- Context menu for additional actions
- Hover prefetching for performance

### 4. Date Calculations

- Automatic positioning based on zoom level
- Smart snapping to appropriate time units
- Proper handling of month/quarter boundaries

### 5. Performance Optimizations

- Memoized date calculations
- Efficient viewport scrolling
- Lazy loading and prefetching
- Local storage for zoom preferences

## Component Breakdown

### GanttBoard (Main Container)

**File**: `gantt-board.tsx` (main export)

**Responsibilities**:

- Manages global state (zoom level, date range)
- Handles story mutations and updates
- Coordinates scrolling behavior
- Fetches and manages story/team data
- Filters stories to only show those with dates

**Props**:

```typescript
type GanttBoardProps = {
  stories: Story[];
  className?: string;
};
```

**Usage**:

```tsx
<GanttBoard stories={storiesArray} className="optional-styling" />
```

### Stories (Left Sidebar)

**Responsibilities**:

- Displays story metadata (ID, title, assignee, etc.)
- Provides interactive controls for story properties
- Handles navigation and prefetching

**Features**:

- Team code + sequence ID (e.g., "PROJ-123")
- Assignee avatar with tooltip showing full details
- Priority icon with dropdown editor
- Status icon with dropdown editor
- Clickable title linking to story detail page
- Duration display (when dates are available)
- Context menu for additional actions
- Permission-based editing (guests cannot modify)

### Chart (Right Timeline)

**Responsibilities**:

- Renders timeline header and story bars
- Manages grid layout and today highlighting
- Coordinates with Bar components for interactions

**Features**:

- Background grid aligned with header columns
- Today highlighting across all story rows
- Weekend highlighting in weeks view
- Interactive gantt bars positioned over grid
- Responsive minimum width based on timeline extent

### Bar (Individual Gantt Bar)

**Responsibilities**:

- Handles drag and resize interactions
- Manages optimistic UI updates
- Calculates positioning for different zoom levels

**Interaction Types**:

- `"move"`: Drag entire bar to shift both start and end dates
- `"resize-start"`: Drag left edge to adjust start date only
- `"resize-end"`: Drag right edge to adjust end date only

**Drag Phases**:

1. **mouseDown**: Record initial state and drag type
2. **mouseMove**: Calculate new position and update visual state
3. **mouseUp**: Finalize dates and sync with API

### TimelineHeader (Date Headers)

**Responsibilities**:

- Renders appropriate headers for each zoom level
- Handles today highlighting
- Manages responsive column widths

**Zoom Level Behaviors**:

- **Weeks**: Two rows (month/week, day/dayname)
- **Months**: Two rows (month/year, start/end day)
- **Quarters**: Two rows (quarter/year, start/end month)

### Header (Control Header)

**Responsibilities**:

- Zoom level selection menu
- Today button for quick navigation
- Sticky positioning for always-visible controls

## Types and Interfaces

### ZoomLevel

```typescript
type ZoomLevel = "weeks" | "months" | "quarters";
```

**Affects**:

- Column width (64px/120px/180px)
- Time unit calculations (days/months/quarters)
- Header display format
- Drag snap behavior

### Story

Main story data type from API containing:

- `id`, `title`, `startDate`, `endDate`
- `assigneeId`, `priority`, `statusId`
- `teamId`, `sequenceId`

### DetailedStory

Extended story type for mutations and updates.

## Helper Functions

### getTimePeriodsForZoom()

```typescript
const getTimePeriodsForZoom = (
  dateRange: { start: Date; end: Date },
  zoomLevel: ZoomLevel,
) => Date[]
```

**Purpose**: Generate an array of time periods for the given date range and zoom level.

**Examples**:

- weeks: `[day1, day2, day3, ...]`
- months: `[month1, month2, month3, ...]`
- quarters: `[quarter1, quarter2, quarter3, ...]`

### getColumnWidth()

```typescript
const getColumnWidth = (zoomLevel: ZoomLevel) => number;
```

**Purpose**: Get the pixel width for each column based on zoom level.

**Returns**:

- weeks: `64px`
- months: `120px`
- quarters: `180px`

### getWeekSpans()

```typescript
const getWeekSpans = (days: Date[]) => WeekSpan[]
```

**Purpose**: Calculate week spans for the weeks view header. Groups consecutive days by week and calculates their visual span.

### getVisibleDateRange()

```typescript
const getVisibleDateRange = (centerDate: Date, viewportDays = 365) => {
  start: Date;
  end: Date;
};
```

**Purpose**: Calculate the visible date range centered on a given date. Creates a viewport window to avoid rendering excessive dates.

## Zoom Level Details

### Weeks View (64px columns)

- **Column Unit**: 1 day
- **Header Rows**:
  - Top: Month/year + week number
  - Bottom: Day number + day name
- **Special Features**: Weekend highlighting, Today highlighting
- **Drag Behavior**: Snaps to daily increments

### Months View (120px columns)

- **Column Unit**: 1 month
- **Header Rows**:
  - Top: Month name + year
  - Bottom: First day (1) + last day (31) of month
- **Drag Behavior**: Snaps to monthly increments
- **Bar Spanning**: Full months or month ranges

### Quarters View (180px columns)

- **Column Unit**: 1 quarter
- **Header Rows**:
  - Top: Quarter (Q1) + year
  - Bottom: First month (Jan) + last month (Mar) of quarter
- **Drag Behavior**: Snaps to quarterly increments
- **Bar Spanning**: Full quarters or quarter ranges

## Drag & Drop System

The interactive drag system operates in three distinct phases:

### Phase 1: MouseDown - Capture Initial State

```typescript
const handleMouseDown = (e: React.MouseEvent, type: "move" | "resize-start" | "resize-end")
```

**Actions**:

- Records original dates and mouse position
- Sets drag type for appropriate behavior
- Prevents default browser behaviors
- Begins tracking drag state

### Phase 2: MouseMove - Calculate New Position

```typescript
const handleMouseMove = (e: MouseEvent)
```

**Actions**:

- Converts pixel movement to time units based on zoom level
- Updates visual position optimistically
- Respects zoom level constraints and boundaries
- Provides real-time visual feedback

### Phase 3: MouseUp - Finalize Changes

```typescript
const handleMouseUp = ()
```

**Actions**:

- Calculates final dates from visual position
- Only calls API if dates actually changed (prevents unnecessary requests)
- Provides visual feedback during server synchronization
- Handles optimistic update cleanup

## State Management

### Local State

- **zoomLevel**: Stored in localStorage, persists across sessions
- **dragPosition**: Temporary visual state during drag operations
- **hasScrolledRef**: Prevents multiple auto-scroll attempts

### Server State (via React Query)

- **stories**: Main story data from API
- **teams**: Team data for code lookups
- **members**: Team member data for assignee dropdowns

### Mutations

- **useUpdateStoryMutation**: Handles all story property updates
- **Optimistic updates**: Immediate UI feedback
- **Error handling**: Rollback on failure

## Performance Considerations

### 1. Memoization

- Date calculations memoized to prevent recalculation
- Stable callback references to prevent re-renders

### 2. Virtualization

- Only renders visible date range (1 year window)
- Efficient scrolling without full date range

### 3. Prefetching

- Story details prefetched on hover
- Route prefetching for instant navigation

### 4. Local Storage

- Zoom level persisted to maintain user preferences

### 5. Optimistic Updates

- Immediate UI feedback during drag operations
- Server sync happens asynchronously

## Integration Points

### APIs

- **useUpdateStoryMutation**: Story property updates
- **useTeams**: Team data for code display
- **useTeamMembers**: Assignee options and avatar display

### Navigation

- **Link**: Story detail pages with SEO-friendly URLs
- **Router**: Prefetching for performance

### UI Components

- **PrioritiesMenu, StatusesMenu, AssigneesMenu**: Property editors
- **StoryContextMenu**: Right-click actions
- **Avatar, Tooltip**: User interface elements

### Hooks

- **useLocalStorage**: Persistent zoom level
- **useUserRole**: Permission-based UI behavior
- **useSession**: Authentication and API access

## Accessibility Features

- Keyboard navigation support
- Screen reader labels and descriptions
- Focus management during interactions
- Semantic HTML structure
- Color contrast compliance
- ARIA attributes for interactive elements

## Usage Examples

### Basic Usage

```tsx
import { GanttBoard } from "@/components/ui/gantt-board";

function ProjectPage() {
  const { data: stories } = useStories();

  return (
    <div className="h-full">
      <GanttBoard stories={stories} />
    </div>
  );
}
```

### With Custom Styling

```tsx
<GanttBoard stories={filteredStories} className="custom-gantt-styling" />
```

### With Story Filtering

```tsx
const storiesWithDates = stories.filter(
  (story) => story.startDate && story.endDate,
);

<GanttBoard stories={storiesWithDates} />;
```

## Troubleshooting

### Common Issues

**Stories not displaying**:

- Ensure stories have both `startDate` and `endDate`
- Check that stories array is not empty
- Verify team data is loaded for team code display

**Drag not working**:

- Check user permissions (guests cannot edit)
- Verify `useUpdateStoryMutation` is available
- Ensure story has valid date range

**Performance issues**:

- Check if too many stories are being rendered
- Verify memoization is working correctly
- Consider filtering stories to smaller dataset

**Scroll issues**:

- Ensure container has proper height constraints
- Check that `hasScrolledRef` is preventing multiple scrolls
- Verify date range calculations are correct

## Future Enhancements

### Planned Features

- Keyboard shortcuts for common actions
- Bulk operations (multi-select)
- Custom date ranges and filtering
- Export functionality (PDF, image)
- Real-time collaboration indicators
- Undo/redo functionality
- Advanced grouping options (by assignee, priority, etc.)

### Extension Points

- Custom zoom levels
- Additional story metadata columns
- Theme customization
- Custom drag behaviors
- Integration with external calendar systems

## Contributing

When modifying this component:

1. **Update documentation** when adding new features
2. **Test all zoom levels** for new functionality
3. **Verify accessibility** with screen readers
4. **Check performance** with large story datasets
5. **Test drag interactions** across all scenarios
6. **Validate responsive behavior** on different screen sizes

## Dependencies

### Required Packages

- `date-fns`: Date manipulation and formatting
- `@tanstack/react-query`: Server state management
- `next/navigation`: Routing and navigation
- `next-auth/react`: Authentication
- Custom UI components and hooks from the project

### Peer Dependencies

- React 18+
- Next.js 13+ (App Router)
- TypeScript 5+
