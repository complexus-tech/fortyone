/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { Team } from "./team";

let mockTeamsWithViews = new Set<string>();

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/my-work",
  useSearchParams: () => new URLSearchParams(),
}));

jest.mock("@dnd-kit/sortable", () => ({
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: jest.fn(),
    transform: null,
    transition: undefined,
    isDragging: false,
  }),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user-1" } } }),
}));

jest.mock("@/hooks", () => ({
  useFeatures: () => ({ objectiveEnabled: true }),
  useSprintsEnabled: () => true,
  useTerminology: () => ({
    getTermDisplay: (term: "storyTerm" | "objectiveTerm" | "sprintTerm") =>
      ({
        storyTerm: "Tasks",
        objectiveTerm: "Objectives",
        sprintTerm: "Sprints",
      })[term],
  }),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/shared/views/presence-context", () => ({
  useViewsPresence: () => ({
    hasTeamViews: (teamId: string) => mockTeamsWithViews.has(teamId),
  }),
}));

jest.mock("@/modules/teams/hooks/remove-member-mutation", () => ({
  useRemoveMemberMutation: () => ({ isPending: false, mutate: jest.fn() }),
}));

jest.mock("@/modules/integration-requests/hooks/use-team-requests", () => ({
  useTeamIntegrationRequests: () => ({ data: undefined }),
}));

jest.mock("@/shared/favorites", () => ({
  FavoriteContextMenuItem: () => null,
  FavoriteMenuItem: () => null,
}));

jest.mock("@/components/ui", () => ({
  ConfirmDialog: () => null,
  TeamColor: () => null,
  NavLink: ({ children, href }: { children: ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

const teamNavigation = (isCollapsed = false) => (
  <Team
    color="#6366f1"
    id="team-1"
    isCollapsed={isCollapsed}
    isOpen
    isPrivate={false}
    name="Product"
    onOpenChange={jest.fn()}
    sortingDisabled
    totalTeams={1}
  />
);

describe("Team Views navigation", () => {
  beforeEach(() => {
    mockTeamsWithViews = new Set();
  });

  it("only shows Views for this team's saved views and updates when availability changes", () => {
    mockTeamsWithViews.add("another-team");
    const { rerender } = render(teamNavigation());
    expect(screen.queryByRole("link", { name: "Views" })).toBeNull();
    expect(screen.getByRole("link", { name: "Tasks" })).toBeVisible();

    mockTeamsWithViews.add("team-1");
    rerender(teamNavigation());
    expect(screen.getByRole("link", { name: "Views" })).toHaveAttribute(
      "href",
      "/acme/views?team=team-1",
    );

    mockTeamsWithViews.delete("team-1");
    rerender(teamNavigation());
    expect(screen.queryByRole("link", { name: "Views" })).toBeNull();
  });

  it.each([false, true])(
    "uses the same availability in collapsed team navigation (views=%s)",
    (hasViews) => {
      if (hasViews) mockTeamsWithViews.add("team-1");
      render(teamNavigation(true));
      fireEvent.click(
        screen.getByRole("button", { name: "Open Product team navigation" }),
      );

      expect(Boolean(screen.queryByRole("link", { name: "Views" }))).toBe(
        hasViews,
      );
      expect(screen.getByRole("link", { name: "Tasks" })).toBeVisible();
    },
  );
});
