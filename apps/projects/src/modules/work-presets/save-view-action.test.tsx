import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import type { SavedViewConfiguration } from "./types";
import { SaveViewAction } from "./save-view-action";

let mockUser = "user";
let mockWorkspace = "acme";
const mockCreate = jest.fn();
const mockPush = jest.fn();
const mockRoute = jest.fn();
const mockTeams = jest.fn();
jest.mock("next/navigation", () => ({ useRouter: () => ({ push: mockPush }) }));
jest.mock("nuqs", () => ({
  parseAsString: {},
  useQueryStates: () => [{}, mockRoute],
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUser } } }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "member" }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: mockWorkspace,
    withWorkspace: (path: string) => `/${mockWorkspace}${path}`,
  }),
}));
jest.mock("@/modules/teams/public/queries", () => ({
  getJoinedTeams: (...args: unknown[]) => mockTeams(...args),
}));
jest.mock("./hooks", () => ({
  usePresetMutations: () => ({ create: { mutateAsync: mockCreate } }),
}));

const configuration: SavedViewConfiguration = {
  version: 1,
  layout: "kanban",
  scope: {
    kind: "my-work",
    tab: "upcoming",
    category: "paused",
    createdAfter: "2026-10-01",
    createdBefore: null,
    overdue: false,
  },
  filters: {
    ...DEFAULT_STORIES_FILTER,
    teamIds: ["one", "two"],
    operators: { teamIds: "isNotAnyOf" },
  },
  viewOptions: {
    groupBy: "priority",
    orderBy: "deadline",
    orderDirection: "asc",
    displayColumns: ["Status", "Created", "Epic"],
    selectedCustomFieldIds: ["field"],
    hiddenKanbanGroups: { priority: ["Low"] },
    showSubStories: true,
    showEmptyGroups: false,
  },
};
const harness = (teamId?: string) => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <SaveViewAction configuration={configuration} teamId={teamId} />
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  jest.clearAllMocks();
  mockUser = "user";
  mockWorkspace = "acme";
  mockTeams.mockResolvedValue([
    { id: "one", name: "Product" },
    { id: "two", name: "Operations" },
  ]);
  mockCreate.mockResolvedValue({ id: "saved" });
});

it("requires explicit owner team and keeps cross-team filters/scope/display unchanged", async () => {
  harness();
  fireEvent.click(screen.getByRole("button", { name: "Save as" }));
  fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
    target: { value: "Upcoming response" },
  });
  expect(screen.getByRole("button", { name: "Create view" })).toBeDisabled();
  const team = screen.getByRole("combobox", { name: "Save to team" });
  await waitFor(() => {
    expect(team).toBeEnabled();
  });
  fireEvent.keyDown(team, { key: "ArrowDown" });
  fireEvent.click(screen.getByRole("option", { name: "Product" }));
  fireEvent.click(screen.getByRole("button", { name: "Create view" }));
  await waitFor(() => {
    expect(mockCreate).toHaveBeenCalledWith({
      teamId: "one",
      kind: "view",
      visibility: "personal",
      name: "Upcoming response",
      configuration,
    });
  });
  expect(mockPush).toHaveBeenCalledWith("/acme/teams/one/stories?view=saved");
  expect(mockRoute).not.toHaveBeenCalled();
});

it("shows a preselected ownership team and returns keyboard focus on cancel", async () => {
  harness("two");
  fireEvent.click(screen.getByRole("button", { name: "Save as" }));
  await waitFor(() => {
    expect(
      screen.getByRole("combobox", { name: "Save to team" }),
    ).toHaveTextContent("Operations");
  });
  fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
  await waitFor(() => {
    expect(screen.getByRole("button", { name: "Save as" })).toHaveFocus();
  });
  expect(mockCreate).not.toHaveBeenCalled();
});

it("does not navigate or restore old-scope focus after the account changes during saving", async () => {
  let finish!: (value: { id: string }) => void;
  mockCreate.mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  const result = harness("one");
  fireEvent.click(screen.getByRole("button", { name: "Save as" }));
  fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
    target: { value: "My view" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Create view" }));
  await waitFor(() => {
    expect(mockCreate).toHaveBeenCalled();
  });
  mockUser = "other";
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  result.rerender(
    <QueryClientProvider client={client}>
      <SaveViewAction configuration={configuration} teamId="one" />
    </QueryClientProvider>,
  );
  finish({ id: "saved" });
  await waitFor(() => {
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
  expect(mockPush).not.toHaveBeenCalled();
  expect(screen.getByRole("button", { name: "Save as" })).not.toHaveFocus();
});

it("does not redirect a completed save after the My Work source tab changes", async () => {
  let finish!: (value: { id: string }) => void;
  mockCreate.mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  const result = harness("one");
  fireEvent.click(screen.getByRole("button", { name: "Save as" }));
  fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
    target: { value: "My upcoming" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Create view" }));
  await waitFor(() => {
    expect(mockCreate).toHaveBeenCalled();
  });
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  result.rerender(
    <QueryClientProvider client={client}>
      <SaveViewAction
        configuration={{
          ...configuration,
          scope: { kind: "my-work", tab: "assigned" },
        }}
        teamId="one"
      />
    </QueryClientProvider>,
  );
  finish({ id: "saved" });
  await waitFor(() => {
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
  expect(mockPush).not.toHaveBeenCalled();
});
