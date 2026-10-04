import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { Objective } from "../types";
import { ObjectiveCard } from "./card";

let mockCanUpdate = true;
const mockMutate = jest.fn();
jest.mock("next/navigation", () => ({ usePathname: () => "/acme/roadmap" }));
jest.mock("../hooks/use-can-update-objective", () => ({
  useCanUpdateObjective: () => mockCanUpdate,
}));
jest.mock("../hooks/update-mutation", () => ({
  useUpdateObjectiveMutation: () => ({ mutate: mockMutate }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
  useTerminology: () => ({ getTermDisplay: () => "key results" }),
}));
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({ data: [{ id: "team-1", code: "ENG" }] }),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({ data: [] }),
}));
jest.mock("@/components/ui", () => {
  const Pass = ({ children }: { children?: ReactNode }) => children;
  const Compound = Object.assign(Pass, { Trigger: Pass, Items: () => null });
  return {
    AssigneesMenu: Compound,
    PrioritiesMenu: Compound,
    PriorityIcon: () => null,
    ObjectiveHealthIcon: () => null,
  };
});
jest.mock("@/components/ui/objective-statuses-menu", () => {
  const Pass = ({ children }: { children?: ReactNode }) => children;
  return {
    ObjectiveStatusesMenu: Object.assign(Pass, {
      Trigger: Pass,
      Items: ({ setStatusId }: { setStatusId: (id: string) => void }) => (
        <button
          onClick={() => {
            setStatusId("started");
          }}
          type="button"
        >
          Set status
        </button>
      ),
    }),
  };
});
jest.mock("@/components/ui/objective-status-icon", () => ({
  ObjectiveStatusIcon: () => null,
}));
jest.mock("./objective-health-editor", () => ({
  ObjectiveHealthEditor: ({ children }: { children?: ReactNode }) => children,
}));
jest.mock("./objective-forecast-risk", () => ({
  ObjectiveForecastRiskBadge: () => null,
}));
const OBJECTIVE: Objective = {
  id: "objective-1",
  sequenceId: 1,
  name: "Improve reliability",
  description: "",
  shortSummary: null,
  leadUser: "owner",
  teamId: "team-1",
  workspaceId: "workspace-1",
  startDate: "2026-01-01",
  endDate: "2026-03-31",
  isPrivate: false,
  createdAt: "2026-01-01",
  updatedAt: "2026-01-01",
  createdBy: "owner",
  statusId: "backlog",
  keyResultCount: 1,
  health: "On Track",
  color: "#4A90E2",
  forecastStartDate: null,
  forecastEndDate: null,
  scheduleStatus: "no_schedule",
  forecastDaysDelta: 0,
  forecastCauseStory: null,
};
describe("objective list title navigation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockCanUpdate = true;
  });
  it("links to the full objective while preserving checkbox, expansion and property actions", () => {
    const select = jest.fn();
    const expand = jest.fn();
    render(
      <ObjectiveCard
        {...OBJECTIVE}
        childCount={1}
        onSelectionChange={select}
        onToggleExpanded={expand}
      />,
    );
    const link = screen.getByRole("link", { name: "Improve reliability" });
    expect(link).toHaveAttribute(
      "href",
      "/acme/teams/team-1/objectives/objective-1",
    );
    expect(
      screen.queryByRole("button", { name: "Improve reliability" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("checkbox"));
    expect(select).toHaveBeenCalledWith(true);
    fireEvent.click(screen.getByRole("button", { name: "Expand key results" }));
    expect(expand).toHaveBeenCalledTimes(1);
    fireEvent.click(screen.getByRole("button", { name: "Set status" }));
    expect(mockMutate).toHaveBeenCalledWith({
      objectiveId: "objective-1",
      data: { statusId: "started" },
    });
    expect(link.querySelector("button")).toBeNull();
  });
  it("keeps the objective link available when the viewer cannot update it", () => {
    mockCanUpdate = false;
    render(<ObjectiveCard {...OBJECTIVE} onSelectionChange={jest.fn()} />);
    expect(
      screen.getByRole("link", { name: "Improve reliability" }),
    ).toHaveAttribute("href", "/acme/teams/team-1/objectives/objective-1");
    expect(screen.getByRole("checkbox")).toBeDisabled();
  });
});
