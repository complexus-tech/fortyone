# UI Components

## Row & Col Components

Flexible layout components for React Native with Tailwind-like API.

### Row Component

Horizontal flex container (flex-row).

```tsx
import { Row, Text } from '@/components/ui';

// Basic usage
<Row>
  <Text>Item 1</Text>
  <Text>Item 2</Text>
</Row>

// With alignment and spacing
<Row align="center" justify="between" gap={4}>
  <Text>Left</Text>
  <Text>Right</Text>
</Row>

// With wrapping
<Row wrap gap={2}>
  <Text>Item 1</Text>
  <Text>Item 2</Text>
  <Text>Item 3</Text>
</Row>
```

**Props:**

- `align`: "start" | "center" | "end" | "stretch" | "baseline" (default: "start")
- `justify`: "start" | "center" | "end" | "between" | "around" | "evenly" (default: "start")
- `gap`: 0 | 1 | 2 | 3 | 4 | 5 | 6 | 8 | 10 | 12 (default: 0)
- `wrap`: boolean (default: false)
- `className`: string (optional custom Tailwind classes)

### Col Component

Vertical flex container (flex-col).

```tsx
import { Col, Text } from '@/components/ui';

// Basic usage
<Col>
  <Text>Item 1</Text>
  <Text>Item 2</Text>
</Col>

// With alignment and spacing
<Col align="center" justify="center" gap={4} flex={1}>
  <Text>Centered content</Text>
</Col>

// Complex layouts
<Row gap={4} className="p-4">
  <Col flex={1} gap={2}>
    <Text fontWeight="bold">Column 1</Text>
    <Text>Content</Text>
  </Col>
  <Col flex={1} gap={2}>
    <Text fontWeight="bold">Column 2</Text>
    <Text>Content</Text>
  </Col>
</Row>
```

**Props:**

- `align`: "start" | "center" | "end" | "stretch" | "baseline" (default: "start")
- `justify`: "start" | "center" | "end" | "between" | "around" | "evenly" (default: "start")
- `gap`: 0 | 1 | 2 | 3 | 4 | 5 | 6 | 8 | 10 | 12 (default: 0)
- `flex`: 1 | "auto" | "initial" | "none" (optional)
- `className`: string (optional custom Tailwind classes)

### Common Patterns

#### Header with actions

```tsx
<Row align="center" justify="between" className="p-4 border-b border-gray-200">
  <Text fontSize="xl" fontWeight="bold">
    Title
  </Text>
  <Button>Action</Button>
</Row>
```

#### Card layout

```tsx
<Col gap={3} className="bg-white p-4 rounded-lg">
  <Row align="center" justify="between">
    <Text fontWeight="semibold">Card Title</Text>
    <Avatar src="..." size="sm" />
  </Row>
  <Text color="muted">Card content goes here</Text>
  <Row gap={2}>
    <Button size="sm">Primary</Button>
    <Button size="sm" variant="outline">
      Secondary
    </Button>
  </Row>
</Col>
```

#### Centered content

```tsx
<Col align="center" justify="center" flex={1} gap={4}>
  <Text fontSize="2xl" fontWeight="bold">
    Welcome
  </Text>
  <Text color="muted">Get started below</Text>
  <Button>Get Started</Button>
</Col>
```

#### List items

```tsx
<Col gap={2}>
  {items.map((item) => (
    <Row
      key={item.id}
      align="center"
      justify="between"
      className="p-3 bg-white rounded-lg"
    >
      <Row align="center" gap={3}>
        <Avatar name={item.name} size="sm" />
        <Text>{item.name}</Text>
      </Row>
      <Text color="muted">{item.status}</Text>
    </Row>
  ))}
</Col>
```

## Skeleton Components

Loading state components with smooth animations.

### StorySkeleton

Single story skeleton - matches the Story component structure.

```tsx
import { StorySkeleton } from "@/components/ui";

// Basic usage
<StorySkeleton />;
```

### StoriesSkeleton

Multiple stories skeleton (without sections) - for simple story lists.

```tsx
import { StoriesSkeleton } from '@/components/ui';

// Default (5 stories)
<StoriesSkeleton />

// Custom count
<StoriesSkeleton count={10} />
```

**Props:**

- `count`: number (default: 5) - Number of story skeletons to display

### StoriesListSkeleton

Stories grouped in sections skeleton - for stories organized by status, priority, or assignee.

```tsx
import { StoriesListSkeleton } from '@/components/ui';

// Default (3 sections, 3 stories each)
<StoriesListSkeleton />

// Custom sections and stories
<StoriesListSkeleton sectionsCount={5} storiesPerSection={4} />
```

**Props:**

- `sectionsCount`: number (default: 3) - Number of section groups
- `storiesPerSection`: number (default: 3) - Stories per section

### Usage Examples

#### Loading states in lists

```tsx
const MyStories = () => {
  const { data: stories, isPending } = useStories();

  if (isPending) {
    return <StoriesSkeleton count={8} />;
  }

  return (
    <ScrollView>
      {stories.map((story) => (
        <Story key={story.id} {...story} />
      ))}
    </ScrollView>
  );
};
```

#### Loading states in grouped lists

```tsx
const MyWork = () => {
  const { data: groupedStories, isPending } = useMyStoriesGrouped();

  if (isPending) {
    return <StoriesListSkeleton />;
  }

  return <GroupedStoriesList sections={groupedStories} />;
};
```
