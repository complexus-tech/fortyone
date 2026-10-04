import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { SavedViewLoadState } from "@/shared/story/view-configuration";
import { ListStories } from "./list-stories";

let mockUserId = "owner";
let mockWorkspace = "acme";
let mockViewId: string | null = "view";
jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team" }),
  usePathname: () => "/acme/teams/team/stories",
  useSearchParams: () =>
    new URLSearchParams(mockViewId ? { view: mockViewId } : {}),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: mockWorkspace,
    withWorkspace: (path: string) => `/${mockWorkspace}${path}`,
  }),
}));
jest.mock("@/hooks", () => ({ useLocalStorage: () => ["list", jest.fn()] }));
jest.mock("./provider", () => ({
  TeamOptionsProvider: ({ children }: { children: ReactNode }) => children,
  useTeamOptions: () => ({
    filters: {},
    resetFilters: jest.fn(),
    setFilters: jest.fn(),
  }),
}));
jest.mock("@/components/ui/stories-filter-bar", () => ({
  StoriesFilterBar: () => <div>Filter bar</div>,
}));
jest.mock("./all-stories", () => ({
  AllStories: () => <div>Story board</div>,
}));
jest.mock("./header", () => ({
  Header: ({
    onViewLoadStateChange,
  }: {
    onViewLoadStateChange: (state: SavedViewLoadState) => void;
  }) => (
    <>
      {["ready", "error", "unavailable"].map((status) => (
        <button
          key={status}
          onClick={() => {
            onViewLoadStateChange({
              workspaceSlug: mockWorkspace,
              userId: mockUserId,
              teamId: "team",
              viewId: mockViewId,
              status: status as SavedViewLoadState["status"],
            });
          }}
          type="button"
        >
          {status}
        </button>
      ))}
    </>
  ),
}));

describe("saved-view story-board readiness", () => {
  beforeEach(() => {
    mockUserId = "owner";
    mockWorkspace = "acme";
    mockViewId = "view";
  });
  it("hides prior board/filter state until the selected saved view has been applied", () => {
    const { container } = render(<ListStories renderSavedViews={() => null} />);
    expect(screen.getByRole("status")).toHaveTextContent(
      "Loading saved view...",
    );
    expect(screen.queryByText("Story board")).not.toBeInTheDocument();
    expect(screen.queryByText("Filter bar")).not.toBeInTheDocument();
    expect(screen.getAllByRole("status")).toHaveLength(1);
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(container.querySelector("[aria-hidden][inert]")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "ready" }));
    expect(screen.getByText("Story board")).toBeInTheDocument();
    expect(screen.getByText("Filter bar")).toBeInTheDocument();
    expect(
      container.querySelector("[aria-hidden][inert]"),
    ).not.toBeInTheDocument();
  });
  it.each(["account", "workspace", "view"])(
    "blocks prior ready state immediately when the %s changes",
    (scope) => {
      const { rerender } = render(
        <ListStories renderSavedViews={() => null} />,
      );
      fireEvent.click(screen.getByRole("button", { name: "ready" }));
      if (scope === "account") mockUserId = "other";
      else if (scope === "workspace") mockWorkspace = "another";
      else mockViewId = "another-view";
      rerender(<ListStories renderSavedViews={() => null} />);
      expect(screen.queryByText("Story board")).not.toBeInTheDocument();
      expect(screen.getByRole("status")).toHaveTextContent(
        "Loading saved view...",
      );
    },
  );
  it.each(["error", "unavailable"])(
    "shows an honest %s state and an exit to team tasks",
    (status) => {
      render(<ListStories renderSavedViews={() => null} />);
      fireEvent.click(screen.getByRole("button", { name: status }));
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.queryByText("Story board")).not.toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: "Back to team tasks" }),
      ).toHaveAttribute("href", "/acme/teams/team/stories");
    },
  );
  it("keeps the normal board available when there is no saved-view URL", () => {
    mockViewId = null;
    render(<ListStories renderSavedViews={() => null} />);
    expect(screen.getByText("Story board")).toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});
