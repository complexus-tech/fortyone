/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { CommandBar } from "./command-bar";

const mockRouterPush = jest.fn();
const mockSetIsOpen = jest.fn();
const mockUseSearch = jest.fn();
let mockStoryTerm = "task";
let mockObjectiveTerm = "objective";
let mockUserRole = "admin";

Object.defineProperty(globalThis, "ResizeObserver", {
  configurable: true,
  value: class {
    observe = jest.fn();
    unobserve = jest.fn();
    disconnect = jest.fn();
  },
});
Element.prototype.scrollIntoView = jest.fn();

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/my-work",
  useRouter: () => ({ push: mockRouterPush }),
}));

jest.mock("next-themes", () => ({
  useTheme: () => ({ resolvedTheme: "dark", setTheme: jest.fn() }),
}));

jest.mock("@/hooks/tracking", () => ({
  useAnalytics: () => ({ analytics: { logout: jest.fn() } }),
}));

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: (
      term: string,
      options: { variant?: string; capitalize?: boolean } = {},
    ) => {
      let value = term === "objectiveTerm" ? mockObjectiveTerm : mockStoryTerm;
      if (options.variant === "plural") value += "s";
      return options.capitalize
        ? value[0].toUpperCase() + value.slice(1)
        : value;
    },
  }),
}));

jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockUserRole }),
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/modules/search/hooks/use-search", () => ({
  useSearch: (params: unknown) => mockUseSearch(params),
}));

jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({
    data: [{ color: "#3B82F6", id: "status-1", name: "In Progress" }],
  }),
}));

jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({
    data: [{ color: "#22C55E", id: "objective-status-1", name: "Active" }],
  }),
}));

jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({ data: [{ code: "WEB", id: "team-1" }] }),
}));

jest.mock("@/components/shared/sidebar/actions", () => ({
  logOut: jest.fn(),
}));

jest.mock("@/components/shared/keyboard-shortcuts", () => ({
  KeyboardShortcuts: () => null,
}));

jest.mock("@/components/ui/invite-members", () => ({
  InviteMembersDialog: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div aria-label="Invite members" role="dialog" /> : null,
}));

jest.mock("@/components/ui/new-objective", () => ({
  NewObjectiveDialog: () => null,
}));

jest.mock("@/components/ui/new-story-dialog", () => ({
  NewStoryDialog: () => null,
}));

jest.mock("@/components/ui/new-sprint-dialog", () => ({
  NewSprintDialog: () => null,
}));

jest.mock("@/components/shared/sidebar/utils", () => ({
  clearAllStorage: jest.fn(),
}));

jest.mock("icons", () => {
  const Icon = () => <span aria-hidden="true" />;

  return {
    DashboardIcon: Icon,
    EnterIcon: Icon,
    HelpIcon: Icon,
    LoadingIcon: Icon,
    LogoutIcon: Icon,
    MoonIcon: Icon,
    Notification02Icon: Icon,
    ObjectiveIcon: Icon,
    PlusIcon: Icon,
    RoadmapIcon: Icon,
    SearchIcon: Icon,
    SettingsIcon: Icon,
    StoryIcon: Icon,
    SunIcon: Icon,
    UserIcon: Icon,
    UsersAddIcon: Icon,
  };
});

jest.mock("ui", () => {
  const { Command, commandFilter } = jest.requireActual(
    "../../../../../packages/ui/src/command",
  );
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );
  const Text = ({ children }: { children?: ReactNode }) => (
    <span>{children}</span>
  );
  const Dialog = Object.assign(
    ({ children, open }: { children?: ReactNode; open?: boolean }) =>
      open ? <div>{children}</div> : null,
    {
      Body: Container,
      Content: Container,
      Header: Container,
      Title: Text,
    },
  );

  return {
    Box: Container,
    Command,
    commandFilter,
    Dialog,
    Divider: () => <hr />,
    Flex: Container,
    Kbd: ({ children }: { children?: ReactNode }) => <kbd>{children}</kbd>,
    Skeleton: Container,
    Text,
  };
});

const searchResponse = {
  objectives: [
    {
      health: "On Track",
      id: "objective-1",
      name: "Increase workspace activation",
      statusId: "objective-status-1",
      teamId: "team-1",
    },
  ],
  page: 1,
  pageSize: 5,
  stories: [
    {
      id: "story-1",
      priority: "High",
      sequenceId: 29,
      statusId: "status-1",
      teamId: "team-1",
      title: "Review activation metrics",
    },
  ],
  totalObjectives: 1,
  totalPages: 1,
  totalStories: 1,
};

describe("CommandBar workspace search", () => {
  beforeEach(() => {
    mockStoryTerm = "task";
    mockObjectiveTerm = "objective";
    mockUserRole = "admin";
    jest.useFakeTimers();
    mockRouterPush.mockReset();
    mockSetIsOpen.mockReset();
    mockUseSearch.mockImplementation(() => ({
      data: searchResponse,
      isError: false,
      isFetching: false,
    }));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  const searchForActivation = () => {
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);

    fireEvent.change(
      screen.getByPlaceholderText("Search tasks, objectives, or commands…"),
      { target: { value: "activation" } },
    );
    act(() => {
      jest.advanceTimersByTime(250);
    });
  };

  const changeQuery = (query: string) => {
    const input = screen.getByRole("combobox");
    fireEvent.change(input, { target: { value: query } });
    return input;
  };

  const settleSearch = () => {
    act(() => {
      jest.advanceTimersByTime(250);
    });
  };

  it("keeps matching quick actions when the query reaches the workspace search threshold", () => {
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);

    for (const query of ["i", "in", "invite"]) {
      changeQuery(query);
      expect(
        screen.getByRole("option", { name: /Invite Members/ }),
      ).toBeInTheDocument();
      expect(screen.queryByRole("option", { name: /Roadmap/ })).toBeNull();
      settleSearch();
      if (query.length >= 2) {
        expect(
          screen.getByRole("option", { name: /Invite Members/ }),
        ).toHaveAttribute("aria-selected", "true");
      }
    }

    const options = screen.getAllByRole("option");
    expect(options).toHaveLength(4);
    expect(options[0]).toHaveTextContent("Invite Members");
    expect(options[1]).toHaveTextContent("Review activation metrics");
    expect(options[2]).toHaveTextContent("Increase workspace activation");
    expect(options[3]).toHaveTextContent("View all results for “invite”");
    expect(screen.queryByRole("option", { name: /Inbox/ })).toBeNull();
    expect(screen.queryByRole("option", { name: /Settings/ })).toBeNull();
  });

  it("opens the matching Invite Members action with Enter before search settles", () => {
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);
    const input = changeQuery("invite");

    fireEvent.keyDown(input, { key: "Enter" });

    expect(screen.getByRole("dialog", { name: "Invite members" })).toBeTruthy();
    expect(mockSetIsOpen).toHaveBeenCalledWith(false);
    expect(mockRouterPush).not.toHaveBeenCalled();
  });

  it("preserves a selected matching action when workspace results arrive", () => {
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);
    const input = changeQuery("in");

    fireEvent.keyDown(input, { key: "ArrowDown" });
    expect(screen.getByRole("option", { name: /Inbox/ })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    settleSearch();
    expect(screen.getByRole("option", { name: /Inbox/ })).toHaveAttribute(
      "aria-selected",
      "true",
    );

    fireEvent.keyDown(input, { key: "Enter" });
    expect(mockRouterPush).toHaveBeenCalledWith("/acme/notifications");
  });

  it("moves from a matching action to workspace results with the keyboard", () => {
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);
    const input = changeQuery("invite");
    settleSearch();

    fireEvent.keyDown(input, { key: "ArrowDown" });
    expect(
      screen.getByRole("option", { name: /Review activation metrics/ }),
    ).toHaveAttribute("aria-selected", "true");
    fireEvent.keyDown(input, { key: "Enter" });

    expect(mockRouterPush).toHaveBeenCalledWith("/acme/work/story-1");
  });

  it("defaults to the first story when no quick actions match the query", () => {
    searchForActivation();

    const options = screen.getAllByRole("option");
    expect(options).toHaveLength(3);
    expect(options[0]).toHaveTextContent("Review activation metrics");
    expect(options[0]).toHaveAttribute("aria-selected", "true");
    expect(screen.queryByText("Quick Actions")).toBeNull();
    expect(screen.queryByText("Go To")).toBeNull();
    expect(screen.queryByText("Settings & Display")).toBeNull();

    fireEvent.keyDown(screen.getByRole("combobox"), { key: "Enter" });
    expect(mockRouterPush).toHaveBeenCalledWith("/acme/work/story-1");
  });

  it("preserves guest permissions while searching quick actions", () => {
    mockUserRole = "guest";
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);
    const input = changeQuery("invite");
    settleSearch();

    expect(screen.queryByRole("option", { name: /Invite Members/ })).toBeNull();

    changeQuery("new");
    settleSearch();
    expect(screen.getByRole("option", { name: /New task/ })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    expect(
      screen.getByRole("option", { name: /New objective/ }),
    ).toHaveAttribute("aria-disabled", "true");
    expect(
      screen.getByRole("option", { name: /Review activation metrics/ }),
    ).toHaveAttribute("aria-selected", "true");

    fireEvent.keyDown(input, { key: "ArrowUp" });
    expect(
      screen.getByRole("option", { name: /View all results for “new”/ }),
    ).toHaveAttribute("aria-selected", "true");
  });

  it("debounces workspace search and opens a task result directly", () => {
    searchForActivation();

    expect(mockUseSearch).toHaveBeenLastCalledWith({
      pageSize: 5,
      query: "activation",
      type: "all",
    });
    expect(screen.getByText("WEB-29").parentElement?.textContent).toBe(
      "WEB-29 · In Progress · High",
    );
    expect(screen.getByText("In Progress").className).toContain("font-medium");
    expect(screen.getByText("High").className).toContain("font-medium");

    fireEvent.click(
      screen.getByRole("option", { name: /Review activation metrics/ }),
    );

    expect(mockSetIsOpen).toHaveBeenCalledWith(false);
    expect(mockRouterPush).toHaveBeenCalledWith("/acme/work/story-1");
  });

  it("opens an objective result directly", () => {
    searchForActivation();

    fireEvent.click(
      screen.getByRole("option", { name: /Increase workspace activation/ }),
    );

    expect(mockRouterPush).toHaveBeenCalledWith(
      "/acme/teams/team-1/objectives/objective-1",
    );
  });

  it("keeps the full search page as an optional fallback", () => {
    searchForActivation();

    fireEvent.click(
      screen.getByRole("option", {
        name: /View all results for “activation”/,
      }),
    );

    expect(mockRouterPush).toHaveBeenCalledWith(
      "/acme/search?query=activation&type=all",
    );
  });
  it("uses workspace terminology in the search input and result headings", () => {
    mockStoryTerm = "issue";
    mockObjectiveTerm = "goal";
    render(<CommandBar isOpen setIsOpen={mockSetIsOpen} />);
    fireEvent.change(
      screen.getByPlaceholderText("Search issues, goals, or commands…"),
      {
        target: { value: "activation" },
      },
    );
    act(() => {
      jest.advanceTimersByTime(300);
    });
    expect(screen.getByText("Issues")).toBeTruthy();
    expect(screen.getByText("Goals")).toBeTruthy();
  });
});
