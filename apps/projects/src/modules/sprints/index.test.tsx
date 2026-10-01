/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { useUserRole } from "@/hooks";
import { SprintsList } from "./index";

jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team-1" }),
}));
jest.mock("@/hooks", () => ({
  useUserRole: jest.fn(),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/first${path}`,
  }),
  useTerminology: () => ({
    getTermDisplay: (_term: string, options?: { variant?: string }) =>
      options?.variant === "plural" ? "sprints" : "sprint",
  }),
}));
jest.mock("@/components/shared", () => ({
  BodyContainer: ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  ),
}));
jest.mock("@/components/ui", () => ({
  NewSprintButton: ({ children }: { children: ReactNode }) => (
    <button type="button">{children}</button>
  ),
}));
jest.mock("@/components/ui/illustrations/empty-state-illustrations", () => ({
  SprintsEmptyIllustration: () => null,
}));
jest.mock("./components/header", () => ({ SprintsHeader: () => null }));
jest.mock("./components/row", () => ({ SprintRow: () => null }));
jest.mock("./components/sprints-skeleton", () => ({
  SprintsSkeleton: () => null,
}));
jest.mock("./hooks/team-sprints", () => ({
  useTeamSprints: () => ({ data: [], isPending: false }),
}));

describe("empty sprint settings entry point", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("takes an admin directly to sprint scheduling inside Planning", () => {
    jest
      .mocked(useUserRole)
      .mockReturnValue({ userRole: "admin" } as ReturnType<typeof useUserRole>);
    render(<SprintsList />);

    expect(
      screen.getByRole("link", { name: "Set up automations" }),
    ).toHaveAttribute(
      "href",
      "/first/settings/workspace/teams/team-1?tab=planning&section=sprints",
    );
    expect(
      screen.queryByRole("button", { name: "Create new sprint" }),
    ).not.toBeInTheDocument();
  });

  it("keeps members on their existing create workflow without an admin settings link", () => {
    jest
      .mocked(useUserRole)
      .mockReturnValue({ userRole: "member" } as ReturnType<
        typeof useUserRole
      >);
    render(<SprintsList />);

    expect(
      screen.getByRole("button", { name: "Create new sprint" }),
    ).toBeVisible();
    expect(
      screen.queryByRole("link", { name: "Set up automations" }),
    ).not.toBeInTheDocument();
  });

  it("keeps guests read-only and explains who can configure scheduling", () => {
    jest
      .mocked(useUserRole)
      .mockReturnValue({ userRole: "guest" } as ReturnType<typeof useUserRole>);
    render(<SprintsList />);

    expect(
      screen.queryByRole("button", { name: "Create new sprint" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Set up automations" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText(/Ask an admin to set up sprint automations/),
    ).toBeVisible();
  });
});
