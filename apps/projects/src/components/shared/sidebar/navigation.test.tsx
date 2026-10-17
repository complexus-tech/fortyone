/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import type * as UI from "ui";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { Navigation } from "./navigation";

let mockPathname = "/acme/calendar";
let mockTeamScope: string | null = null;
let mockView: string | null = null;
let mockHasViews = true;
let mockChatEnabled = true;
let mockSubdomain = false;
let mockConversation: string | null = null;
let mockChatUnread = 0;

jest.mock("@/modules/chat/use-chat-unread", () => ({
  useChatUnread: () => ({ data: { totalUnread: mockChatUnread } }),
}));

jest.mock("@/lib/hooks/workspace/chat-capabilities", () => ({
  useChatCapabilities: () => ({ enabled: mockChatEnabled }),
}));

jest.mock("@/shared/views/presence-context", () => ({
  useViewsPresence: () => ({ hasViews: mockHasViews }),
}));

jest.mock("next/navigation", () => ({
  usePathname: () => mockPathname,
  useSearchParams: () =>
    new URLSearchParams({
      ...(mockTeamScope ? { team: mockTeamScope } : {}),
      ...(mockView ? { view: mockView } : {}),
      ...(mockConversation ? { conversation: mockConversation } : {}),
    }),
}));

jest.mock("icons", () => ({
  ActiveSprintIcon: () => null,
  AiIcon: () => null,
  CalendarIcon: () => null,
  ChatIcon: () => null,
  ChevronRightIcon: () => null,
  DashboardIcon: () => null,
  DocsIcon: () => null,
  ViewsIcon: () => null,
  RoadmapIcon: () => null,
  StrategyIcon: () => null,
  UserIcon: () => null,
}));

jest.mock("ui", () => {
  const React = jest.requireActual("react");
  const CollapsibleContext = React.createContext({
    onOpenChange: (_open: boolean) => undefined,
    open: true,
  });
  const Collapsible = Object.assign(
    ({
      children,
      onOpenChange,
      open,
    }: {
      children: ReactNode;
      onOpenChange: (open: boolean) => void;
      open: boolean;
    }) => (
      <CollapsibleContext.Provider value={{ onOpenChange, open }}>
        <div>{children}</div>
      </CollapsibleContext.Provider>
    ),
    {
      Content: ({ children }: { children: ReactNode }) => {
        const { open } = React.useContext(CollapsibleContext);
        return open ? <div>{children}</div> : null;
      },
      Trigger: ({ children }: { children: ReactNode }) => {
        const { onOpenChange, open } = React.useContext(CollapsibleContext);

        if (!React.isValidElement(children)) return children;

        return React.cloneElement(children, {
          "aria-expanded": open,
          onClick: () => {
            onOpenChange(!open);
          },
        });
      },
    },
  );

  return {
    Badge: jest.requireActual<typeof UI>("ui").Badge,
    Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    Collapsible,
    Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  };
});

jest.mock("@/components/ui", () => ({
  NavLink: ({
    active,
    className,
    children,
    href,
    "data-walkthrough-target": walkthroughTarget,
    "aria-label": ariaLabel,
  }: {
    active?: boolean;
    className?: string;
    children: ReactNode;
    href: string;
    "data-walkthrough-target"?: string;
    "aria-label"?: string;
  }) => (
    <a
      aria-current={active ? "page" : undefined}
      aria-label={ariaLabel}
      className={className}
      data-walkthrough-target={walkthroughTarget}
      href={href}
    >
      {children}
    </a>
  ),
}));

jest.mock("@/hooks", () => {
  const { useLocalStorage } = jest.requireActual("@/hooks/local-storage");

  return {
    useFeatures: () => ({ objectiveEnabled: true }),
    useLocalStorage,
    useTerminology: () => ({
      getTermDisplay: () => "sprint",
    }),
    useUserRole: () => ({ userRole: "member" }),
    useWorkspacePath: () => ({
      withWorkspace: (path: string) => (mockSubdomain ? path : `/acme${path}`),
      workspaceSlug: "acme",
    }),
  };
});

jest.mock("@/modules/sprints/hooks/running-sprints", () => ({
  useRunningSprints: () => ({
    data: [{ id: "sprint-1", teamId: "team-1" }],
  }),
}));

describe("Navigation", () => {
  beforeEach(() => {
    localStorage.clear();
    mockPathname = "/acme/calendar";
    mockTeamScope = null;
    mockView = null;
    mockHasViews = true;
    mockChatEnabled = true;
    mockSubdomain = false;
    mockConversation = null;
    mockChatUnread = 0;
  });

  it.each([false, true])(
    "shows all unread chat messages when collapsed=%s",
    (isCollapsed) => {
      mockChatUnread = 500;
      render(<Navigation isCollapsed={isCollapsed} />);
      expect(
        screen.getByRole("link", { name: "Chat, 500 unread messages" }),
      ).toHaveTextContent("500");
    },
  );

  it("caps the visual count while retaining the exact accessible count", () => {
    mockChatUnread = 1250;
    render(<Navigation />);
    expect(
      screen.getByRole("link", { name: "Chat, 1,250 unread messages" }),
    ).toHaveTextContent("999+");
  });

  it.each([
    { path: "/acme/chat", subdomain: false, conversation: "team" },
    { path: "/acme/chat", subdomain: false, conversation: "direct-message" },
    { path: "/chat", subdomain: true, conversation: "team" },
    { path: "/chat", subdomain: true, conversation: "direct-message" },
    { path: "/acme/chat", subdomain: true, conversation: "team" },
    { path: "/acme/chat", subdomain: true, conversation: "direct-message" },
  ])(
    "keeps Chat highlighted on $path for $conversation (subdomain=$subdomain)",
    ({ path, subdomain, conversation }) => {
      mockPathname = path;
      mockSubdomain = subdomain;
      mockConversation = conversation;
      render(<Navigation />);
      expect(screen.getByRole("link", { name: "Chat" })).toHaveAttribute(
        "aria-current",
        "page",
      );
      expect(screen.getByRole("link", { name: "My work" })).not.toHaveAttribute(
        "aria-current",
      );
    },
  );

  it("does not highlight Chat for an unrelated route with a similar prefix", () => {
    mockPathname = "/acme/chat-settings";
    render(<Navigation />);
    expect(screen.getByRole("link", { name: "Chat" })).not.toHaveAttribute(
      "aria-current",
    );
  });

  it("hides Chat until the server enables it for this workspace", () => {
    mockChatEnabled = false;
    const view = render(<Navigation />);
    expect(
      screen.queryByRole("link", { name: "Chat" }),
    ).not.toBeInTheDocument();
    mockChatEnabled = true;
    view.rerender(<Navigation />);
    expect(screen.getByRole("link", { name: "Chat" })).toHaveAttribute(
      "href",
      "/acme/chat",
    );
  });

  it("groups shared destinations under an expanded Workspace section", () => {
    render(<Navigation />);

    const workspaceTrigger = screen.getByRole("button", {
      name: "Workspace navigation",
    });

    expect(workspaceTrigger).toHaveAttribute("aria-expanded", "true");

    expect(
      screen.getAllByRole("link").map((link) => link.textContent.trim()),
    ).toEqual([
      "My work",
      "Chat",
      "AI Agent",
      "Calendar",
      "Active Sprint",
      "Summary",
      "Roadmap",
      "Views",
      "Strategy Map",
      "Documents",
    ]);

    const calendarLink = screen.getByRole("link", { name: "Calendar" });
    expect(calendarLink).toHaveAttribute("href", "/acme/calendar");
    expect(calendarLink).toHaveAttribute("aria-current", "page");
    expect(calendarLink).toHaveClass(
      "before:w-1",
      "before:h-[30px]",
      "before:-left-4",
    );

    const activeSprintLink = screen.getByRole("link", {
      name: "Active Sprint",
    });
    expect(activeSprintLink).toHaveAttribute(
      "href",
      "/acme/teams/team-1/sprints/sprint-1/stories",
    );
    expect(activeSprintLink).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.sprintsNavigation,
    );
    expect(screen.getByRole("link", { name: "AI Agent" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.mayaNavigation,
    );
    expect(screen.getByRole("link", { name: "My work" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.myWork,
    );
    expect(calendarLink).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.calendar,
    );
    expect(screen.getByRole("link", { name: "Roadmap" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.roadmap,
    );
    expect(screen.getByRole("link", { name: "Summary" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.summary,
    );
    expect(screen.getByRole("link", { name: "Strategy Map" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.strategy,
    );
    expect(screen.getByRole("link", { name: "Documents" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.documents,
    );

    fireEvent.click(workspaceTrigger);

    expect(workspaceTrigger).toHaveAttribute("aria-expanded", "false");
    expect(localStorage.getItem("sidebar:acme:workspace-expanded")).toBe(
      "false",
    );
    expect(
      screen.getAllByRole("link").map((link) => link.textContent.trim()),
    ).toEqual([
      "My work",
      "Chat",
      "AI Agent",
      "Calendar",
      "Active Sprint",
      "Summary",
    ]);

    fireEvent.click(workspaceTrigger);

    expect(workspaceTrigger).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("link", { name: "Roadmap" })).toBeVisible();
  });

  it("respects a collapsed Workspace preference", () => {
    localStorage.setItem("sidebar:acme:workspace-expanded", "false");

    render(<Navigation />);

    expect(
      screen.getByRole("button", { name: "Workspace navigation" }),
    ).toHaveAttribute("aria-expanded", "false");
    expect(screen.getByRole("link", { name: "Active Sprint" })).toBeVisible();
    expect(screen.queryByRole("link", { name: "Roadmap" })).toBeNull();
    expect(screen.getByRole("link", { name: "Calendar" })).toBeVisible();
  });

  it.each([false, true])(
    "only shows Views when saved views exist (collapsed=%s)",
    (isCollapsed) => {
      mockHasViews = false;
      const { rerender } = render(<Navigation isCollapsed={isCollapsed} />);
      expect(screen.queryByRole("link", { name: "Views" })).toBeNull();
      expect(screen.getByRole("link", { name: "Documents" })).toBeVisible();

      mockHasViews = true;
      rerender(<Navigation isCollapsed={isCollapsed} />);
      expect(screen.getByRole("link", { name: "Views" })).toBeVisible();

      mockHasViews = false;
      rerender(<Navigation isCollapsed={isCollapsed} />);
      expect(screen.queryByRole("link", { name: "Views" })).toBeNull();
    },
  );

  it("opens Workspace when the current route belongs there", () => {
    localStorage.setItem("sidebar:acme:workspace-expanded", "false");
    mockPathname = "/acme/roadmap";

    render(<Navigation />);

    expect(
      screen.getByRole("button", { name: "Workspace navigation" }),
    ).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("link", { name: "Roadmap" })).toBeVisible();
  });

  it("keeps destination labels visible in the collapsed sidebar", () => {
    render(<Navigation isCollapsed />);

    const calendarLink = screen.getByRole("link", { name: "Calendar" });
    const calendarLabel = screen.getByText("Calendar");
    const myWorkLink = screen.getByRole("link", { name: "My work" });

    expect(calendarLink).toHaveClass("flex-col", "text-center");
    expect(calendarLink).toHaveClass(
      "bg-primary/5",
      "text-primary",
      "before:w-1",
      "before:h-10",
      "before:-left-3",
    );
    expect(myWorkLink).toHaveClass(
      "hover:bg-primary/5",
      "hover:text-primary",
      "hover:[&_svg]:text-primary",
    );
    expect(calendarLabel).toBeVisible();
    expect(calendarLabel).not.toHaveClass("hidden");
  });

  it("marks Views active and opens Workspace for a selected story view", () => {
    localStorage.setItem("sidebar:acme:workspace-expanded", "false");
    mockPathname = "/acme/teams/team-1/stories";
    mockView = "23a20d25-331e-44d6-929c-0309eae1735c";
    render(<Navigation />);
    expect(screen.getByRole("link", { name: "Views" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(
      screen.getByRole("button", { name: "Workspace navigation" }),
    ).toHaveAttribute("aria-expanded", "true");
  });

  it.each([
    "/acme/teams/team-1/objectives",
    "/acme/teams/team-1/sprints/sprint-1/stories",
  ])(
    "does not treat a layout parameter on %s as a saved story view",
    (path) => {
      mockPathname = path;
      mockView = "gantt";
      render(<Navigation />);
      expect(screen.getByRole("link", { name: "Views" })).not.toHaveAttribute(
        "aria-current",
      );
    },
  );
});
