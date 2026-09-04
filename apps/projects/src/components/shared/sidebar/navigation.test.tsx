/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { Navigation } from "./navigation";

let mockPathname = "/acme/calendar";

jest.mock("next/navigation", () => ({
  usePathname: () => mockPathname,
}));

jest.mock("icons", () => ({
  ActiveSprintIcon: () => null,
  AiIcon: () => null,
  CalendarIcon: () => null,
  ChevronRightIcon: () => null,
  DashboardIcon: () => null,
  DocsIcon: () => null,
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
  }: {
    active?: boolean;
    className?: string;
    children: ReactNode;
    href: string;
    "data-walkthrough-target"?: string;
  }) => (
    <a
      aria-current={active ? "page" : undefined}
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
      withWorkspace: (path: string) => `/acme${path}`,
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
      "AI Agent",
      "Calendar",
      "Active Sprint",
      "Summary",
      "Roadmap",
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
    ).toEqual(["My work", "AI Agent", "Calendar", "Active Sprint", "Summary"]);

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
});
