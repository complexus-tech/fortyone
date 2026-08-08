/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { Navigation } from "./navigation";

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/calendar",
}));

jest.mock("icons", () => ({
  ActiveSprintIcon: () => null,
  AiIcon: () => null,
  CalendarIcon: () => null,
  DashboardIcon: () => null,
  DocsIcon: () => null,
  RoadmapIcon: () => null,
  StrategyIcon: () => null,
  UserIcon: () => null,
}));

jest.mock("ui", () => ({
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

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

jest.mock("@/hooks", () => ({
  useFeatures: () => ({ objectiveEnabled: true }),
  useTerminology: () => ({
    getTermDisplay: () => "sprint",
  }),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/modules/sprints/hooks/running-sprints", () => ({
  useRunningSprints: () => ({
    data: [{ id: "sprint-1", teamId: "team-1" }],
  }),
}));

describe("Navigation", () => {
  it("shows Calendar and the active sprint as top-level destinations", () => {
    render(<Navigation />);

    expect(
      screen.getAllByRole("link").map((link) => link.textContent.trim()),
    ).toEqual([
      "My work",
      "Calendar",
      "AI Assistant",
      "Summary",
      "Active Sprint",
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
  });
});
