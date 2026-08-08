/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { Navigation } from "./navigation";

jest.mock("next/navigation", () => ({
  usePathname: () => "/acme/calendar",
}));

jest.mock("icons", () => ({
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
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

describe("Navigation", () => {
  it("shows Calendar as a top-level destination without Active sprint", () => {
    render(<Navigation />);

    expect(
      screen.getAllByRole("link").map((link) => link.textContent.trim()),
    ).toEqual([
      "My work",
      "Calendar",
      "AI Assistant",
      "Summary",
      "Roadmap",
      "Strategy Map",
      "Documents",
    ]);

    const calendarLink = screen.getByRole("link", { name: "Calendar" });
    expect(calendarLink).toHaveAttribute("href", "/acme/calendar");
    expect(calendarLink).toHaveAttribute("aria-current", "page");
    expect(screen.queryByText(/active sprint/i)).toBeNull();
  });
});
