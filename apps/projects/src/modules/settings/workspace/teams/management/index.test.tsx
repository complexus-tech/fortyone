/* global afterAll, beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { useState as UseState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { Team } from "@/modules/teams/types";
import { useTeam } from "@/modules/teams/hooks/use-team";
import type { TeamSettingsQueryTab } from "./navigation";
import { TeamManagement } from "./index";

const mockSetQuery = jest.fn();
const mockScrollIntoView = jest.fn();
let mockInitialQuery: { tab: TeamSettingsQueryTab; section: string | null };
let mockWorkflowsEnabled = true;

jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team-1" }),
}));

jest.mock("nuqs", () => ({
  parseAsString: {},
  parseAsStringLiteral: () => ({ withDefault: () => ({}) }),
  useQueryStates: () => {
    const { useState } = jest.requireActual<{ useState: typeof UseState }>(
      "react",
    );
    const [query, setQuery] = useState(mockInitialQuery);
    return [
      query,
      (patch: Partial<typeof mockInitialQuery>) => {
        mockSetQuery(patch);
        setQuery((current) => ({ ...current, ...patch }));
      },
    ];
  },
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: (
      term: string,
      options?: { variant?: string; capitalize?: boolean },
    ) => {
      const singular = term === "sprintTerm" ? "sprint" : "task";
      const value = options?.variant === "plural" ? `${singular}s` : singular;
      return options?.capitalize
        ? value.charAt(0).toUpperCase() + value.slice(1)
        : value;
    },
  }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/first${path}`,
  }),
}));

jest.mock("@/modules/teams/hooks/use-team", () => ({ useTeam: jest.fn() }));
jest.mock("@/components/ui", () => ({ TeamColor: () => null }));
jest.mock("@/modules/settings/components", () => ({
  SettingsBackButton: ({ href, label }: { href: string; label: string }) => (
    <a href={href}>{label}</a>
  ),
}));
jest.mock("./components/general", () => ({
  GeneralSettings: () => <div>General editor</div>,
}));
jest.mock("./components/members", () => ({
  MembersSettings: () => <div>Membership editor</div>,
}));
jest.mock("./components/delete", () => ({
  DeleteTeam: () => <button type="button">Delete team</button>,
}));
jest.mock("./components/workflows", () => ({
  WorkflowSettings: () => (
    <div>
      {mockWorkflowsEnabled ? "Workflow editor" : "Workflow upgrade required"}
    </div>
  ),
}));
jest.mock("./components/estimation", () => ({
  EstimationSettings: () => <div>Complexity editor</div>,
}));
jest.mock("@/modules/custom-fields/public/settings", () => ({
  TeamCustomFieldSettings: () => <div>Field editor</div>,
}));
jest.mock("@/modules/team-automations/public/settings", () => ({
  TeamAutomations: () => <div>Rule and recurrence editor</div>,
}));
jest.mock("./components/sprints", () => ({
  SprintSettings: () => <div>Sprint scheduling editor</div>,
}));
jest.mock("./components/github-automations", () => ({
  GitHubAutomations: () => <div>GitHub automation editor</div>,
}));
jest.mock("@/modules/teams/hooks/use-team-settings", () => ({
  useTeamSettings: () => ({
    data: {
      storyAutomationSettings: {
        autoCloseInactiveEnabled: false,
        autoArchiveEnabled: false,
      },
    },
  }),
}));
jest.mock(
  "@/modules/teams/hooks/update-story-automation-settings-mutation",
  () => ({
    useUpdateStoryAutomationSettingsMutation: () => ({ mutate: jest.fn() }),
  }),
);

const team: Team = {
  id: "team-1",
  name: "Product",
  code: "PRO",
  color: "blue",
  isPrivate: false,
  memberCount: 1,
  workspaceId: "workspace-1",
  sprintsEnabled: true,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

const loadedTeam = { data: team } as ReturnType<typeof useTeam>;
const scrollIntoViewDescriptor = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollIntoView",
);

describe("team settings organization", () => {
  beforeAll(() => {
    Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
      configurable: true,
      value: mockScrollIntoView,
    });
  });

  afterAll(() => {
    if (scrollIntoViewDescriptor) {
      Object.defineProperty(
        HTMLElement.prototype,
        "scrollIntoView",
        scrollIntoViewDescriptor,
      );
    } else {
      Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
    }
  });

  beforeEach(() => {
    jest.clearAllMocks();
    mockInitialQuery = { tab: "general", section: null };
    mockWorkflowsEnabled = true;
    jest.mocked(useTeam).mockReturnValue(loadedTeam);
  });

  it("has six native tabs and keeps General focused on team details and danger controls", () => {
    render(<TeamManagement />);

    expect(screen.getAllByRole("tab").map((tab) => tab.textContent)).toEqual([
      "General",
      "Members",
      "Workflow",
      "Planning",
      "Custom fields",
      "Automations",
    ]);
    expect(screen.getByRole("tab", { name: "General" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(
      screen.getByRole("region", { name: "Team details" }),
    ).toHaveTextContent("General editor");
    expect(
      screen.getByRole("region", { name: "Danger zone" }),
    ).toHaveTextContent("Delete team");
    expect(screen.queryByText("Complexity editor")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Back to teams" })).toHaveAttribute(
      "href",
      "/first/settings/workspace/teams",
    );
  });

  it("keeps Planning and custom fields available when the workflow requires an upgrade", () => {
    mockInitialQuery = { tab: "workflows", section: null };
    mockWorkflowsEnabled = false;
    render(<TeamManagement />);

    expect(screen.getByText("Workflow upgrade required")).toBeVisible();
    fireEvent.mouseDown(screen.getByRole("tab", { name: "Planning" }), {
      button: 0,
      ctrlKey: false,
    });
    expect(screen.getByText("Complexity editor")).toBeVisible();
    expect(screen.getByText("Sprint scheduling editor")).toBeVisible();
    fireEvent.mouseDown(screen.getByRole("tab", { name: "Custom fields" }), {
      button: 0,
      ctrlKey: false,
    });
    expect(screen.getByText("Field editor")).toBeVisible();
  });

  it.each([
    [{ tab: "sprints", section: null }, "Planning", "Sprints scheduling"],
    [{ tab: "delete", section: null }, "General", "Danger zone"],
    [{ tab: "workflows", section: "fields" }, "Custom fields", "Custom fields"],
    [{ tab: "workflows", section: "complexity" }, "Planning", "Complexity"],
    [
      { tab: "automations", section: "sprints" },
      "Planning",
      "Sprints scheduling",
    ],
  ] as const)(
    "opens the correct tab and focuses the section for legacy location %j",
    (query, tabLabel, sectionLabel) => {
      mockInitialQuery = query;
      render(<TeamManagement />);

      expect(screen.getByRole("tab", { name: tabLabel })).toHaveAttribute(
        "aria-selected",
        "true",
      );
      expect(screen.getByRole("region", { name: sectionLabel })).toHaveFocus();
      expect(mockScrollIntoView).toHaveBeenCalledWith({ block: "start" });
    },
  );

  it("focuses the target after team data arrives on a cold legacy link", () => {
    mockInitialQuery = { tab: "workflows", section: "fields" };
    jest
      .mocked(useTeam)
      .mockReturnValue({ data: undefined } as ReturnType<typeof useTeam>);
    const view = render(<TeamManagement />);
    expect(mockScrollIntoView).not.toHaveBeenCalled();

    jest.mocked(useTeam).mockReturnValue(loadedTeam);
    view.rerender(<TeamManagement />);

    expect(screen.getByRole("region", { name: "Custom fields" })).toHaveFocus();
    expect(mockScrollIntoView).toHaveBeenCalledWith({ block: "start" });
  });

  it("clears a stale section when changing tabs", () => {
    mockInitialQuery = { tab: "workflows", section: "fields" };
    render(<TeamManagement />);

    fireEvent.mouseDown(screen.getByRole("tab", { name: "Automations" }), {
      button: 0,
      ctrlKey: false,
    });

    expect(mockSetQuery).toHaveBeenCalledWith({
      tab: "automations",
      section: null,
    });
    expect(screen.getByRole("tab", { name: "Automations" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.queryByText("Field editor")).not.toBeInTheDocument();
    expect(screen.getByText("Rule and recurrence editor")).toBeVisible();
  });

  it("groups related automation controls inside one tab", () => {
    mockInitialQuery = { tab: "automations", section: null };
    render(<TeamManagement />);

    expect(
      screen.getByRole("region", { name: "Rules and recurring work" }),
    ).toHaveTextContent("Rule and recurrence editor");
    expect(screen.getByRole("region", { name: "Cleanup" })).toHaveTextContent(
      "Auto-close tasks",
    );
    expect(
      screen.getByRole("region", { name: "GitHub automations" }),
    ).toHaveTextContent("GitHub automation editor");
    expect(
      screen.queryByText("Sprint scheduling editor"),
    ).not.toBeInTheDocument();
  });

  it("groups Complexity and sprint scheduling in Planning", () => {
    mockInitialQuery = { tab: "planning", section: null };
    render(<TeamManagement />);

    expect(screen.getByRole("tab", { name: "Planning" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(
      screen.getByRole("region", { name: "Complexity" }),
    ).toHaveTextContent("Complexity editor");
    expect(
      screen.getByRole("region", { name: "Sprints scheduling" }),
    ).toHaveTextContent("Sprint scheduling editor");
    expect(screen.queryByText("Field editor")).not.toBeInTheDocument();
  });

  it("retains the direct Custom fields tab next to Automations", () => {
    mockInitialQuery = { tab: "fields", section: null };
    render(<TeamManagement />);

    expect(screen.getByRole("tab", { name: "Custom fields" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByText("Field editor")).toBeVisible();
    const tabs = screen.getAllByRole("tab");
    expect(tabs.at(-2)).toHaveTextContent("Custom fields");
    expect(tabs.at(-1)).toHaveTextContent("Automations");
  });
});
