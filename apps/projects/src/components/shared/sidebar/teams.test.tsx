/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { Teams } from "./teams";

const mockReorderTeams = jest.fn();
const mockReact = jest.requireActual("react");
let mockActiveTeamId: string | undefined;
const mockTeams = Array.from({ length: 6 }, (_, index) => ({
  id: `team-${index + 1}`,
  name: `Team ${index + 1}`,
  color: "#6366f1",
  isPrivate: false,
}));

jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: mockActiveTeamId }),
}));

jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Button: ({
    children,
    leftIcon,
    "data-walkthrough-target": walkthroughTarget,
  }: {
    children: ReactNode;
    leftIcon?: ReactNode;
    "data-walkthrough-target"?: string;
  }) => (
    <button data-walkthrough-target={walkthroughTarget} type="button">
      {leftIcon}
      {children}
    </button>
  ),
  Flex: ({
    children,
    "data-walkthrough-target": walkthroughTarget,
  }: {
    children: ReactNode;
    "data-walkthrough-target"?: string;
  }) => <div data-walkthrough-target={walkthroughTarget}>{children}</div>,
  Divider: () => <hr />,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
}));

jest.mock("@dnd-kit/core", () => ({
  DndContext: ({
    children,
    onDragEnd,
  }: {
    children: ReactNode;
    onDragEnd: (event: {
      active: { id: string };
      over: { id: string };
    }) => void;
  }) => (
    <div>
      {children}
      <button
        onClick={() => {
          onDragEnd({
            active: { id: "team-1" },
            over: { id: "team-3" },
          });
        }}
        type="button"
      >
        Reorder teams
      </button>
    </div>
  ),
}));

jest.mock("@dnd-kit/sortable", () => ({
  SortableContext: ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  ),
  verticalListSortingStrategy: {},
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user-1" } } }),
}));

jest.mock("@/modules/teams/hooks/teams", () => ({
  useJoinedTeams: () => ({ data: mockTeams }),
}));

jest.mock("@/hooks", () => ({
  useLocalStorage: (_key: string, initialValue: unknown) =>
    mockReact.useState(initialValue),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("@/modules/teams/hooks/remove-member-mutation", () => ({
  useRemoveMemberMutation: () => ({ isPending: false, mutate: jest.fn() }),
}));

jest.mock("@/modules/teams/hooks/add-member-mutation", () => ({
  useAddMemberMutation: () => ({ mutate: jest.fn() }),
}));

jest.mock("@/modules/teams/hooks/use-reorder-teams", () => ({
  useReorderTeamsMutation: () => ({ mutate: mockReorderTeams }),
}));

jest.mock("@/modules/team-feedback/hooks/use-team-feedback-summaries", () => ({
  useTeamFeedbackSummaries: () => ({ data: [] }),
}));

jest.mock("@/components/ui", () => ({
  ConfirmDialog: () => null,
}));

jest.mock("@/components/ui/teams-menu", () => {
  const MockTeamsMenu = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const MockTeamsMenuTrigger = ({ children }: { children: ReactNode }) => (
    <>{children}</>
  );
  const MockTeamsMenuItems = ({
    onPinTeam,
    overflowTeams,
  }: {
    onPinTeam: (teamId: string) => void;
    overflowTeams: typeof mockTeams;
  }) => (
    <div data-testid="teams-menu-items">
      {overflowTeams.map((team) => (
        <button
          key={team.id}
          onClick={() => {
            onPinTeam(team.id);
          }}
          type="button"
        >
          Pin {team.name}
        </button>
      ))}
    </div>
  );
  MockTeamsMenu.Trigger = MockTeamsMenuTrigger;
  MockTeamsMenu.Items = MockTeamsMenuItems;

  return { TeamsMenu: MockTeamsMenu };
});

jest.mock("./team", () => ({
  Team: ({
    id,
    isOpen,
    name,
    onOpenChange,
  }: {
    id: string;
    isOpen: boolean;
    name: string;
    onOpenChange: (open: boolean) => void;
  }) => (
    <button
      aria-expanded={isOpen}
      data-team-id={id}
      onClick={() => {
        onOpenChange(!isOpen);
      }}
      type="button"
    >
      {name}
    </button>
  ),
}));

describe("Teams", () => {
  beforeEach(() => {
    mockActiveTeamId = undefined;
    mockReorderTeams.mockReset();
  });

  it("renders four teams and keeps only one expanded", () => {
    render(<Teams />);

    expect(screen.getByRole("button", { name: "Team 1" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
    expect(screen.getByRole("button", { name: "Team 4" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Team 5" })).toBeNull();
    expect(screen.getByTestId("teams-menu-items")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Team 2" }));

    expect(screen.getByRole("button", { name: "Team 1" })).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    expect(screen.getByRole("button", { name: "Team 2" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
  });

  it("preserves every team in the reorder payload", () => {
    render(<Teams />);

    fireEvent.click(screen.getByRole("button", { name: "Reorder teams" }));

    expect(mockReorderTeams).toHaveBeenCalledWith({
      teamIds: ["team-2", "team-3", "team-1", "team-4", "team-5", "team-6"],
    });
  });

  it("pins an overflow team first and preserves every team", () => {
    render(<Teams />);

    fireEvent.click(screen.getByRole("button", { name: "Pin Team 5" }));

    expect(mockReorderTeams).toHaveBeenCalledWith({
      teamIds: ["team-5", "team-1", "team-2", "team-3", "team-4", "team-6"],
    });
  });

  it("rearranges visible teams when the active team came from overflow", () => {
    mockActiveTeamId = "team-6";
    render(<Teams />);

    fireEvent.click(screen.getByRole("button", { name: "Reorder teams" }));

    expect(mockReorderTeams).toHaveBeenCalledWith({
      teamIds: ["team-2", "team-3", "team-1", "team-6", "team-4", "team-5"],
    });
  });

  it("keeps walkthrough targets available in the collapsed sidebar", () => {
    const { container } = render(<Teams isCollapsed />);

    expect(
      container.querySelector(
        `[data-walkthrough-target="${walkthroughTargets.teams}"]`,
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /your teams/i })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.manageTeams,
    );
  });
});
