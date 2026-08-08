/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { Navigation } from "./navigation";

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/calendar",
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
    children,
    href,
  }: {
    active?: boolean;
    children: ReactNode;
    href: string;
  }) => (
    <a aria-current={active ? "page" : undefined} href={href}>
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
  });

  it("groups shared destinations under an expanded Workspace section", () => {
    render(<Navigation />);

    const workspaceTrigger = screen.getByRole("button", {
      name: "Workspace",
    });

    expect(workspaceTrigger).toHaveAttribute("aria-expanded", "true");

    expect(
      screen.getAllByRole("link").map((link) => link.textContent.trim()),
    ).toEqual([
      "My work",
      "Calendar",
      "Active Sprint",
      "AI Assistant",
      "Summary",
      "Roadmap",
      "Strategy Map",
      "Documents",
    ]);

    const calendarLink = screen.getByRole("link", { name: "Calendar" });
    expect(calendarLink).toHaveAttribute("href", "/acme/calendar");
    expect(calendarLink).toHaveAttribute("aria-current", "page");

    expect(screen.getByRole("link", { name: "Active Sprint" })).toHaveAttribute(
      "href",
      "/acme/teams/team-1/sprints/sprint-1/stories",
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
      "Calendar",
      "Active Sprint",
      "AI Assistant",
      "Summary",
    ]);

    fireEvent.click(workspaceTrigger);

    expect(workspaceTrigger).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("link", { name: "Roadmap" })).toBeVisible();
  });

  it("respects a collapsed Workspace preference", () => {
    localStorage.setItem("sidebar:acme:workspace-expanded", "false");

    render(<Navigation />);

    expect(screen.getByRole("button", { name: "Workspace" })).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    expect(screen.getByRole("link", { name: "Active Sprint" })).toBeVisible();
    expect(screen.queryByRole("link", { name: "Roadmap" })).toBeNull();
    expect(screen.getByRole("link", { name: "Calendar" })).toBeVisible();
  });
});
