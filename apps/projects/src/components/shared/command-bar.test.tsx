/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { CommandBar } from "./command-bar";

const mockRouterPush = jest.fn();
const mockSetIsOpen = jest.fn();
const mockUseSearch = jest.fn();

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/my-work",
  useRouter: () => ({ push: mockRouterPush }),
}));

jest.mock("next-themes", () => ({
  useTheme: () => ({ resolvedTheme: "dark", setTheme: jest.fn() }),
}));

jest.mock("@/hooks", () => ({
  useAnalytics: () => ({ analytics: { logout: jest.fn() } }),
  useTerminology: () => ({
    getTermDisplay: (term: string) =>
      term === "objectiveTerm" ? "objective" : "task",
  }),
  useUserRole: () => ({ userRole: "admin" }),
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

jest.mock("@/components/ui", () => ({
  InviteMembersDialog: () => null,
  NewObjectiveDialog: () => null,
  NewStoryDialog: () => null,
}));

jest.mock("@/components/ui/new-sprint-dialog", () => ({
  NewSprintDialog: () => null,
}));

jest.mock("./sidebar/utils", () => ({
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
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );
  const Text = ({ children }: { children?: ReactNode }) => (
    <span>{children}</span>
  );
  const Command = Object.assign(Container, {
    Group: ({
      children,
      heading,
    }: {
      children?: ReactNode;
      heading?: ReactNode;
    }) => (
      <section>
        {heading}
        {children}
      </section>
    ),
    Input: ({
      onValueChange,
      placeholder,
      value,
    }: {
      onValueChange?: (value: string) => void;
      placeholder?: string;
      value?: string;
    }) => (
      <input
        onChange={(event) => {
          onValueChange?.(event.target.value);
        }}
        placeholder={placeholder}
        value={value}
      />
    ),
    Item: ({
      children,
      disabled,
      onSelect,
    }: {
      children?: ReactNode;
      disabled?: boolean;
      onSelect?: () => void;
    }) => (
      <button disabled={disabled} onClick={onSelect} type="button">
        {children}
      </button>
    ),
    List: Container,
    Loading: Container,
  });
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
      screen.getByRole("button", { name: /Review activation metrics/ }),
    );

    expect(mockSetIsOpen).toHaveBeenCalledWith(false);
    expect(mockRouterPush).toHaveBeenCalledWith("/acme/work/story-1");
  });

  it("opens an objective result directly", () => {
    searchForActivation();

    fireEvent.click(
      screen.getByRole("button", { name: /Increase workspace activation/ }),
    );

    expect(mockRouterPush).toHaveBeenCalledWith(
      "/acme/teams/team-1/objectives/objective-1",
    );
  });

  it("keeps the full search page as an optional fallback", () => {
    searchForActivation();

    fireEvent.click(
      screen.getByRole("button", {
        name: /View all results for “activation”/,
      }),
    );

    expect(mockRouterPush).toHaveBeenCalledWith(
      "/acme/search?query=activation&type=all",
    );
  });
});
