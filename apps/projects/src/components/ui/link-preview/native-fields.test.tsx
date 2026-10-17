/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import type { ReactNode } from "react";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { DetailedStory } from "@/shared/story/types";
import { useUserRole } from "@/hooks/role";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { getJoinedTeams } from "@/modules/teams/queries/get-teams";
import { NativeStoryFields, canEditNativeWork } from "./native-fields";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/hooks/role", () => ({ useUserRole: jest.fn() }));
jest.mock("@/modules/story/hooks/update-mutation", () => ({
  useUpdateStoryMutation: jest.fn(),
}));
jest.mock("@/modules/objectives/hooks/update-mutation", () => ({
  useUpdateObjectiveMutation: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-teams", () => ({
  getJoinedTeams: jest.fn(),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: async () => [
    { id: "started", name: "In progress", color: "red", teamId: "team" },
  ],
}));
jest.mock("@/lib/queries/members/get-members", () => ({
  getTeamMembers: async () => [],
}));
jest.mock("@/modules/objectives/queries/statuses", () => ({
  getObjectiveStatuses: async () => [],
}));
jest.mock("../story/statuses-menu", () => ({
  StatusesMenu: Object.assign(
    ({ children }: { children: ReactNode }) => <>{children}</>,
    {
      Trigger: ({ children }: { children: ReactNode }) => <>{children}</>,
      Items: ({ setStatusId }: { setStatusId: (id: string) => void }) => (
        <button
          onClick={() => {
            setStatusId("done");
          }}
          type="button"
        >
          Select done
        </button>
      ),
    },
  ),
}));
jest.mock("../story/priorities-menu", () => ({
  PrioritiesMenu: Object.assign(
    ({ children }: { children: ReactNode }) => <>{children}</>,
    {
      Trigger: ({ children }: { children: ReactNode }) => <>{children}</>,
      Items: ({ setPriority }: { setPriority: (priority: string) => void }) => (
        <button
          onClick={() => {
            setPriority("High");
          }}
          type="button"
        >
          Select high
        </button>
      ),
    },
  ),
}));
jest.mock("../story/assignees-menu", () => ({
  AssigneesMenu: Object.assign(
    ({ children }: { children: ReactNode }) => <>{children}</>,
    {
      Trigger: ({ children }: { children: ReactNode }) => <>{children}</>,
      Items: ({
        onAssigneeSelected,
      }: {
        onAssigneeSelected: (id: string | null) => void;
      }) => (
        <button
          onClick={() => {
            onAssigneeSelected(null);
          }}
          type="button"
        >
          Select unassigned
        </button>
      ),
    },
  ),
}));
jest.mock("../objective-statuses-menu", () => ({
  ObjectiveStatusesMenu: () => null,
}));

const story = {
  id: "story-id",
  title: "Plan the launch",
  teamId: "team",
  statusId: "started",
  priority: "Low",
  assigneeId: null,
  archivedAt: null,
  deletedAt: null,
} as DetailedStory;
let client: QueryClient;
const mutate = jest.fn();
beforeEach(() => {
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  jest.mocked(useUserRole).mockReturnValue({ userRole: "member" });
  jest
    .mocked(getJoinedTeams)
    .mockResolvedValue([{ id: "team" }] as Awaited<
      ReturnType<typeof getJoinedTeams>
    >);
  jest.mocked(useUpdateStoryMutation).mockReturnValue({
    mutate,
    isPending: false,
    isError: false,
  } as unknown as ReturnType<typeof useUpdateStoryMutation>);
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.clearAllMocks();
});
const renderFields = (value = story) =>
  render(
    <QueryClientProvider client={client}>
      <NativeStoryFields story={value} />
    </QueryClientProvider>,
  );

describe("native inline editing", () => {
  it.each([
    { role: undefined, sameWorkspace: true, joinedTeam: true },
    { role: "guest", sameWorkspace: true, joinedTeam: true },
    { role: "admin", sameWorkspace: false, joinedTeam: true },
    { role: "member", sameWorkspace: true, joinedTeam: false },
    { role: "member", sameWorkspace: true, joinedTeam: true, archived: true },
  ])("keeps unauthorized context read only: %j", (context) => {
    expect(canEditNativeWork(context)).toBe(false);
  });

  it("edits through existing mutations with only the changed field in the payload", async () => {
    renderFields();
    expect(
      screen.queryByRole("button", { name: "Edit title" }),
    ).not.toBeInTheDocument();
    await screen.findByText("In progress");
    fireEvent.click(screen.getByRole("button", { name: "Select done" }));
    expect(mutate).toHaveBeenCalledWith(
      { storyId: "story-id", payload: { statusId: "done" } },
      { onSuccess: undefined },
    );
    fireEvent.click(screen.getByRole("button", { name: "Select high" }));
    expect(mutate).toHaveBeenCalledWith(
      { storyId: "story-id", payload: { priority: "High" } },
      { onSuccess: undefined },
    );
    fireEvent.click(screen.getByRole("button", { name: "Select unassigned" }));
    expect(mutate).toHaveBeenCalledWith(
      { storyId: "story-id", payload: { assigneeId: null } },
      { onSuccess: undefined },
    );
  });

  it("never exposes edit controls to guests even when they belong to the team", async () => {
    jest.mocked(useUserRole).mockReturnValue({ userRole: "guest" });
    renderFields();
    await waitFor(() => {
      expect(screen.getByText("In progress")).toBeVisible();
    });
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(mutate).not.toHaveBeenCalled();
  });

  it("renders the story title without an inline title editor", async () => {
    renderFields();
    await screen.findByText("In progress");
    expect(screen.getByText(story.title)).toBeVisible();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Edit title" }),
    ).not.toBeInTheDocument();
    expect(mutate).not.toHaveBeenCalled();
  });
});
