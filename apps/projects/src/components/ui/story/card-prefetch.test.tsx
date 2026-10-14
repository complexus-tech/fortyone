/* global beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { Story } from "@/modules/stories/types";
import { prefetchStory } from "@/modules/story/queries/prefetch-story";
import { StoryCard } from "./card";

const mockQueryClient = {};
const mockSession = { user: { id: "viewer" } };
let mockSignedIn = true;
let mockDragActive = false;

jest.mock("ui", () => ({
  Avatar: () => null,
  Badge: () => null,
  Box: "div",
  Button: () => null,
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
}));
jest.mock("next/link", () => ({ __esModule: true, default: "a" }));
jest.mock("@dnd-kit/core", () => ({
  useDraggable: () => ({
    active: mockDragActive ? { id: "dragging-story" } : null,
    isDragging: mockDragActive,
  }),
}));
jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => mockQueryClient,
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: mockSignedIn ? mockSession : null }),
}));
jest.mock("@/modules/story/queries/prefetch-story", () => ({
  prefetchStory: jest.fn(),
}));
jest.mock("@/hooks", () => ({
  useMediaQuery: () => true,
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    workspaceSlug: "workspace",
    withWorkspace: (path: string) => `/workspace${path}`,
  }),
}));
jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: () => ({ data: { openStoryInDialog: true } }),
}));
jest.mock("@/lib/hooks/figma", () => ({
  useFigmaHandoffStatuses: () => ({ data: null }),
}));
jest.mock("@/modules/settings/workspace/integrations/figma/icon", () => ({
  FigmaIcon: () => null,
}));
jest.mock("@/shared/story/board-property-slots", () => ({
  useBoardPropertySlots: () => ({ Badges: () => null }),
}));
jest.mock("../board-context", () => ({
  useBoardView: () => ({ isColumnVisible: () => false }),
}));
jest.mock("../member-tooltip", () => ({ MemberTooltip: () => null }));
jest.mock("../priority-icon", () => ({ PriorityIcon: () => null }));
jest.mock("./people-avatars", () => ({ PeopleAvatars: () => null }));
jest.mock("./assignees-menu", () => ({ AssigneesMenu: () => null }));
jest.mock("./context-menu", () => ({
  StoryContextMenu: ({ children }: { children: ReactNode }) => children,
}));
jest.mock("./properties", () => ({ StoryProperties: () => null }));

const story = {
  id: "story",
  title: "Open this ticket",
  sequenceId: 41,
  teamId: "team",
  team: { code: "PRO" },
  subStories: [],
} as unknown as Story;

beforeEach(() => {
  jest.mocked(prefetchStory).mockClear();
  mockSignedIn = true;
  mockDragActive = false;
});

it.each(["hover", "focus"])(
  "waits for %s intent before preloading and preserves dialog opening",
  (intent) => {
    const openStory = jest.fn();
    render(<StoryCard handleStoryClick={openStory} story={story} />);
    expect(prefetchStory).not.toHaveBeenCalled();
    const link = screen.getByRole("link", { name: story.title });

    if (intent === "hover") fireEvent.mouseEnter(link);
    else fireEvent.focus(link);

    expect(prefetchStory).toHaveBeenCalledWith(mockQueryClient, story.id, {
      session: mockSession,
      workspaceSlug: "workspace",
    });
    expect(link).toHaveAttribute("href", "/workspace/work/PRO-41");
    fireEvent.click(link);
    expect(openStory).toHaveBeenCalledWith(story.id);
  },
);

it.each(["signed out", "dragging"])(
  "does not preload on hover or focus while %s",
  (state) => {
    mockSignedIn = state !== "signed out";
    mockDragActive = state === "dragging";
    render(<StoryCard handleStoryClick={jest.fn()} story={story} />);
    const link = screen.getByRole("link", { name: story.title });

    fireEvent.mouseEnter(link);
    fireEvent.focus(link);
    expect(prefetchStory).not.toHaveBeenCalled();
  },
);
