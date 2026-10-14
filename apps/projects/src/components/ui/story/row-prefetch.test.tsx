/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { Story } from "@/modules/stories/types";
import { prefetchStory } from "@/modules/story/queries/prefetch-story";
import { StoryRow } from "./row";

const mockRouterPrefetch = jest.fn();
const mockQueryClient = {};
const mockSession = { user: { id: "viewer" } };
let mockSignedIn = true;
let mockDesktop = true;
let mockOpenInDialog = true;

jest.mock("ui", () => ({
  Avatar: () => null,
  Box: "div",
  Checkbox: () => null,
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  Tooltip: ({ children }: { children: ReactNode }) => children,
}));
jest.mock("icons", () => ({
  ArrowRight2Icon: () => null,
  StoryIcon: () => null,
  SubStoryIcon: () => null,
}));
jest.mock("next/link", () => ({
  __esModule: true,
  default: "a",
}));
jest.mock("next/navigation", () => ({
  useRouter: () => ({ prefetch: mockRouterPrefetch }),
}));
jest.mock("@dnd-kit/core", () => ({ useDraggable: () => ({}) }));
jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => mockQueryClient,
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: mockSignedIn ? mockSession : null }),
}));
jest.mock("@/modules/story/queries/prefetch-story", () => ({
  prefetchStory: jest.fn(),
}));
jest.mock("@/modules/story/hooks/update-mutation", () => ({
  useUpdateStoryMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("@/hooks", () => ({
  useMediaQuery: () => mockDesktop,
  useTerminology: () => ({ getTermDisplay: () => "story" }),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    workspaceSlug: "workspace",
    withWorkspace: (path: string) => `/workspace${path}`,
  }),
}));
jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: () => ({
    data: { openStoryInDialog: mockOpenInDialog },
  }),
}));
jest.mock("@/shared/story/board-property-slots", () => ({
  useBoardPropertySlots: () => ({ Badges: () => null }),
}));
jest.mock("../board-context", () => ({
  useBoard: () => ({
    embedded: true,
    selectedStories: [],
    isColumnVisible: () => false,
  }),
}));
jest.mock("../member-tooltip", () => ({ MemberTooltip: () => null }));
jest.mock("../row-wrapper", () => ({ RowWrapper: "div" }));
jest.mock("./people-avatars", () => ({ PeopleAvatars: () => null }));
jest.mock("./assignees-menu", () => ({ AssigneesMenu: () => null }));
jest.mock("./context-menu", () => ({
  StoryContextMenu: ({ children }: { children: ReactNode }) => children,
}));
jest.mock("./drag-handle", () => ({ DragHandle: () => null }));
jest.mock("./properties", () => ({
  StoryPriorityProperty: () => null,
  StoryProperties: () => null,
  StoryStatusProperty: () => null,
}));

const story = {
  id: "story",
  title: "Open this ticket",
  sequenceId: 41,
  teamId: "team",
  team: { code: "PRO" },
  subStories: [],
} as unknown as Story;

beforeEach(() => {
  mockRouterPrefetch.mockClear();
  jest.mocked(prefetchStory).mockClear();
  mockSignedIn = true;
  mockDesktop = true;
  mockOpenInDialog = true;
});

describe("story row intent prefetch", () => {
  it("warms the dialog resources without requesting an unused page", () => {
    const handleStoryClick = jest.fn();
    render(<StoryRow handleStoryClick={handleStoryClick} story={story} />);
    const link = screen.getByRole("link", { name: story.title });

    fireEvent.mouseEnter(link);
    expect(prefetchStory).toHaveBeenCalledWith(mockQueryClient, story.id, {
      session: mockSession,
      workspaceSlug: "workspace",
    });
    expect(mockRouterPrefetch).not.toHaveBeenCalled();
    expect(link).toHaveAttribute("href", "/workspace/work/PRO-41");
    fireEvent.click(link);
    expect(handleStoryClick).toHaveBeenCalledWith(story.id);
  });

  it.each(["page preference", "mobile"])(
    "keeps route prefetch for %s navigation",
    (mode) => {
      mockOpenInDialog = mode !== "page preference";
      mockDesktop = mode !== "mobile";
      render(<StoryRow handleStoryClick={jest.fn()} story={story} />);

      fireEvent.mouseEnter(screen.getByRole("link", { name: story.title }));
      expect(mockRouterPrefetch).toHaveBeenCalledWith("/workspace/work/PRO-41");
    },
  );

  it("does not prefetch authenticated resources without a session", () => {
    mockSignedIn = false;
    render(<StoryRow handleStoryClick={jest.fn()} story={story} />);

    fireEvent.mouseEnter(screen.getByRole("link", { name: story.title }));
    expect(prefetchStory).not.toHaveBeenCalled();
  });
});
