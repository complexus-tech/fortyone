import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ApiError } from "api-client";
import type { DetailedStory } from "@/shared/story/types";
import type { Team } from "@/modules/teams/public/types";
import type { SavedView } from "@/modules/work-presets/public/views";
import type { FavoriteRef } from "@/shared/favorites";
import { getStory } from "@/modules/story/public/queries";
import { getTeamDetails } from "@/modules/teams/public/queries";
import { resolveSavedViews } from "@/modules/work-presets/public/views";
import { getStatuses } from "@/lib/queries/states/get-states";
import { teamKeys } from "@/constants/keys";
import {
  updateFavorites,
  readFavoriteSnapshot,
} from "@/shared/favorites/store";
import { FavoritesSidebar } from "./favorites-sidebar";
import { favoriteRequests } from "./favorites-resolution";

let mockUserId = "one";
let mockWorkspace = "acme";
let mockPath = "/acme/my-work";
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: mockWorkspace,
    withWorkspace: (path: string) => `/${mockWorkspace}${path}`,
  }),
}));
jest.mock("next/navigation", () => ({
  usePathname: () => mockPath,
  useSearchParams: () => new URLSearchParams(),
}));
jest.mock("@/modules/story/public/queries", () => ({ getStory: jest.fn() }));
jest.mock("@/modules/teams/public/queries", () => ({
  getTeamDetails: jest.fn(),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: jest.fn(),
}));
jest.mock("@/modules/work-presets/public/views", () => ({
  resolveSavedViews: jest.fn(),
  selectCurrentView: jest.fn(),
  presetKey: (workspace: string, user: string, team: string, kind: string) => [
    "work-presets",
    workspace,
    user,
    team,
    kind,
  ],
  savedViewPath: (team: string, id: string) =>
    `/teams/${team}/stories?view=${id}`,
}));

const scope = { workspaceSlug: "acme", userId: "one" };
const first = "00000000-0000-4000-8000-000000000001";
const second = "00000000-0000-4000-8000-000000000002";
const third = "00000000-0000-4000-8000-000000000003";
const team = (id: string, name: string) =>
  ({ id, name, color: "#123456" }) as Team;
const view = (id: string, name: string) =>
  ({
    id,
    teamId: third,
    kind: "view",
    name,
    configuration: { layout: "kanban" },
  }) as SavedView;
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
const sidebarUI = (client: QueryClient, isCollapsed = false) => (
  <QueryClientProvider client={client}>
    <div data-sidebar-content>
      <button data-workspace-switcher type="button">
        Workspace
      </button>
      <FavoritesSidebar isCollapsed={isCollapsed} />
    </div>
    <button type="button">Dialog action</button>
  </QueryClientProvider>
);
const setup = (
  favorites: FavoriteRef[],
  client = new QueryClient({ defaultOptions: { queries: { retryDelay: 0 } } }),
) => {
  updateFavorites(scope, () => favorites);
  const rendered = render(sidebarUI(client));
  return { ...rendered, client };
};

describe("mixed Favorites sidebar", () => {
  beforeEach(() => {
    localStorage.clear();
    jest.clearAllMocks();
    mockUserId = "one";
    mockWorkspace = "acme";
    mockPath = "/acme/my-work";
    jest.mocked(getStatuses).mockResolvedValue([]);
  });

  it("resolves current names and canonical links for all entity kinds and recognizes UUID task routes", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(second, "Current team"));
    jest.mocked(getStory).mockResolvedValue({
      id: first,
      title: "Current task",
      teamCode: "NEW",
      sequenceId: 26,
      statusId: "status",
    } as DetailedStory);
    jest
      .mocked(resolveSavedViews)
      .mockResolvedValue([view(first, "Current board")]);
    mockPath = `/acme/work/${first}`;
    setup([
      { kind: "story", id: first },
      { kind: "team", id: second },
      { kind: "view", id: first, teamId: third },
    ]);
    expect(
      await screen.findByRole("link", { name: "Current task" }),
    ).toHaveAttribute("href", "/acme/work/NEW-26");
    expect(screen.getByRole("link", { name: "Current task" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(
      await screen.findByRole("link", { name: "Current team" }),
    ).toHaveAttribute("href", `/acme/teams/${second}/stories`);
    expect(
      await screen.findByRole("link", { name: "Current board" }),
    ).toHaveAttribute("href", `/acme/teams/${third}/stories?view=${first}`);
    expect(screen.getAllByRole("heading", { name: "Favorites" })).toHaveLength(
      1,
    );
  });

  it("moves removal focus to the next/previous favorite and a stable local target for the final row", async () => {
    const names = new Map([
      [first, "First"],
      [second, "Second"],
      [third, "Third"],
    ]);
    jest
      .mocked(getTeamDetails)
      .mockImplementation(async (id) => team(id, names.get(id) ?? "Team"));
    setup([
      { kind: "team", id: first },
      { kind: "team", id: second },
      { kind: "team", id: third },
    ]);
    await screen.findByRole("link", { name: "Second" });
    const removeSecond = screen.getByRole("button", {
      name: "Remove Second from favorites",
    });
    act(() => {
      removeSecond.focus();
    });
    fireEvent.click(removeSecond);
    expect(screen.getByRole("link", { name: "Third" })).toHaveFocus();
    const removeThird = screen.getByRole("button", {
      name: "Remove Third from favorites",
    });
    act(() => {
      removeThird.focus();
    });
    fireEvent.click(removeThird);
    expect(screen.getByRole("link", { name: "First" })).toHaveFocus();
    const removeFirst = screen.getByRole("button", {
      name: "Remove First from favorites",
    });
    act(() => {
      removeFirst.focus();
    });
    fireEvent.click(removeFirst);
    expect(screen.getByRole("button", { name: "Workspace" })).toHaveFocus();
    expect(
      screen.queryByRole("heading", { name: "Favorites" }),
    ).not.toBeInTheDocument();
  });

  it("uses the local fallback when removing a view remounts its neighbor's grouped query", async () => {
    const remaining = deferred<SavedView[]>();
    jest
      .mocked(resolveSavedViews)
      .mockResolvedValueOnce([view(first, "First"), view(second, "Second")])
      .mockImplementationOnce(() => remaining.promise);
    setup([
      { kind: "view", id: first, teamId: third },
      { kind: "view", id: second, teamId: third },
    ]);
    const button = await screen.findByRole("button", {
      name: "Remove First from favorites",
    });
    act(() => {
      button.focus();
    });
    fireEvent.click(button);
    expect(screen.getByRole("button", { name: "Workspace" })).toHaveFocus();
    expect(
      screen.queryByRole("link", { name: "Second" }),
    ).not.toBeInTheDocument();
    await act(async () => {
      remaining.resolve([view(second, "Second")]);
    });
    expect(
      await screen.findByRole("link", { name: "Second" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Workspace" })).toHaveFocus();
  });

  it("keeps final-row focus inside the active mobile sidebar rather than a hidden desktop instance", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    updateFavorites(scope, () => [{ kind: "team", id: first }]);
    const client = new QueryClient();
    render(
      <QueryClientProvider client={client}>
        <div data-sidebar-content hidden>
          <button data-workspace-switcher type="button">
            Desktop workspace
          </button>
        </div>
        <div data-sidebar-content>
          <button data-workspace-switcher type="button">
            Mobile workspace
          </button>
          <FavoritesSidebar isCollapsed={false} />
        </div>
      </QueryClientProvider>,
    );
    const button = await screen.findByRole("button", {
      name: "Remove Team from favorites",
    });
    act(() => {
      button.focus();
    });
    fireEvent.click(button);
    expect(
      screen.getByRole("button", { name: "Mobile workspace" }),
    ).toHaveFocus();
  });

  it("preserves failed references while cleaning confirmed inaccessible ones and restores cleanup focus", async () => {
    jest.mocked(getTeamDetails).mockImplementation(async (id) => {
      throw new ApiError("Failure", id === first ? 403 : 500, null);
    });
    setup([
      { kind: "team", id: first },
      { kind: "team", id: second },
    ]);
    const cleanup = await screen.findByRole("button", {
      name: "Remove unavailable favorites",
    });
    await screen.findByRole("alert");
    act(() => {
      cleanup.focus();
    });
    fireEvent.click(cleanup);
    expect(readFavoriteSnapshot(scope).favorites).toEqual([
      { kind: "team", id: second },
    ]);
    expect(screen.getByRole("button", { name: "Workspace" })).toHaveFocus();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });

  it("recovers retry focus to the freshly authorized link without clearing a transiently failed favorite", async () => {
    jest
      .mocked(getTeamDetails)
      .mockRejectedValue(new ApiError("Offline", 500, null));
    setup([{ kind: "team", id: first }]);
    const retry = await screen.findByRole("button", { name: "Try again" });
    await waitFor(() => {
      expect(retry).toBeEnabled();
    });
    expect(
      screen.queryByRole("button", { name: "Remove unavailable favorites" }),
    ).not.toBeInTheDocument();
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Recovered"));
    act(() => {
      retry.focus();
    });
    fireEvent.click(retry);
    const link = await screen.findByRole("link", { name: "Recovered" });
    await waitFor(() => {
      expect(link).toHaveFocus();
    });
    expect(readFavoriteSnapshot(scope).favorites).toEqual([
      { kind: "team", id: first },
    ]);
  });

  it("does not steal focus from a dialog when favorites change elsewhere", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    setup([{ kind: "team", id: first }]);
    await screen.findByRole("link", { name: "Team" });
    const action = screen.getByRole("button", { name: "Dialog action" });
    act(() => {
      action.focus();
    });
    act(() => {
      updateFavorites(scope, () => []);
    });
    expect(action).toHaveFocus();
  });

  it("leaves focus with the user when they move away while a retry is pending", async () => {
    const recovery = deferred<Team>();
    jest
      .mocked(getTeamDetails)
      .mockRejectedValue(new ApiError("Offline", 500, null));
    setup([{ kind: "team", id: first }]);
    const retry = await screen.findByRole("button", { name: "Try again" });
    await waitFor(() => {
      expect(retry).toBeEnabled();
    });
    jest.mocked(getTeamDetails).mockImplementation(() => recovery.promise);
    act(() => {
      retry.focus();
    });
    fireEvent.click(retry);
    const action = screen.getByRole("button", { name: "Dialog action" });
    act(() => {
      action.focus();
    });
    await act(async () => {
      recovery.resolve(team(first, "Recovered"));
    });
    await screen.findByRole("link", { name: "Recovered" });
    expect(action).toHaveFocus();
  });

  it("keeps the row and removal focus when browser persistence fails", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    setup([{ kind: "team", id: first }]);
    const button = await screen.findByRole("button", {
      name: "Remove Team from favorites",
    });
    const write = jest
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new Error("Storage blocked");
      });
    act(() => {
      button.focus();
    });
    fireEvent.click(button);
    expect(screen.getByRole("link", { name: "Team" })).toBeInTheDocument();
    expect(button).toHaveFocus();
    write.mockRestore();
  });

  it("refreshes team favorites after ordinary team-list invalidation, including lost access", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Before rename"));
    const { client } = setup([{ kind: "team", id: first }]);
    await screen.findByRole("link", { name: "Before rename" });
    jest
      .mocked(getTeamDetails)
      .mockResolvedValue({ ...team(first, "After rename"), color: "#abcdef" });
    await act(async () => {
      await client.invalidateQueries({ queryKey: teamKeys.lists("acme") });
    });
    expect(
      await screen.findByRole("link", { name: "After rename" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Before rename" }),
    ).not.toBeInTheDocument();
    jest
      .mocked(getTeamDetails)
      .mockRejectedValue(new ApiError("Denied", 403, null));
    await act(async () => {
      await client.invalidateQueries({ queryKey: teamKeys.lists("acme") });
    });
    await waitFor(() => {
      expect(
        screen.queryByRole("link", { name: "After rename" }),
      ).not.toBeInTheDocument();
    });
    expect(
      screen.getByRole("button", { name: "Remove unavailable favorites" }),
    ).toBeInTheDocument();
  });

  it("withholds another account's cached name until its own fresh authorization succeeds", async () => {
    const fresh = deferred<Team>();
    jest
      .mocked(getTeamDetails)
      .mockResolvedValueOnce(team(first, "First account"))
      .mockImplementationOnce(() => fresh.promise);
    const { client, rerender } = setup([{ kind: "team", id: first }]);
    await screen.findByRole("link", { name: "First account" });
    updateFavorites({ ...scope, userId: "two" }, () => [
      { kind: "team", id: first },
    ]);
    mockUserId = "two";
    const key = favoriteRequests([{ kind: "team", id: first }], "two", {
      workspaceSlug: "acme",
      session: null,
    })[0].key;
    client.setQueryData(key, [team(first, "Cached sensitive name")]);
    rerender(
      <QueryClientProvider client={client}>
        <div data-sidebar-content>
          <FavoritesSidebar isCollapsed={false} />
        </div>
      </QueryClientProvider>,
    );
    expect(
      screen.queryByRole("link", { name: "First account" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Cached sensitive name" }),
    ).not.toBeInTheDocument();
    await act(async () => {
      fresh.resolve(team(first, "Authorized current name"));
    });
    expect(
      await screen.findByRole("link", { name: "Authorized current name" }),
    ).toBeInTheDocument();
  });

  it("allows an active favorite to be collapsed and reopened with an accessible heading trigger", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Active team"));
    mockPath = `/acme/teams/${first}/stories`;
    setup([{ kind: "team", id: first }]);
    const link = await screen.findByRole("link", { name: "Active team" });
    expect(link).toHaveAttribute("aria-current", "page");
    const trigger = screen.getByRole("button", { name: "Favorites" });
    expect(screen.getByRole("heading", { name: "Favorites" })).toContainElement(
      trigger,
    );
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    const contentId = trigger.getAttribute("aria-controls");
    expect(contentId).toBeTruthy();
    expect(document.getElementById(contentId!)).toContainElement(link);
    act(() => {
      trigger.focus();
    });
    fireEvent.click(trigger);
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(trigger).toHaveFocus();
    expect(
      screen.queryByRole("link", { name: "Active team" }),
    ).not.toBeInTheDocument();
    fireEvent.click(trigger);
    expect(
      await screen.findByRole("link", { name: "Active team" }),
    ).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
  });

  it("restores the saved expansion state after remounting", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    const initial = setup([{ kind: "team", id: first }]);
    await screen.findByRole("link", { name: "Team" });
    fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
    initial.unmount();
    setup([{ kind: "team", id: first }]);
    expect(screen.getByRole("button", { name: "Favorites" })).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    expect(
      screen.queryByRole("link", { name: "Team" }),
    ).not.toBeInTheDocument();
  });

  it("isolates expansion preferences by account and workspace without overwriting another scope", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    const favorite = { kind: "team" as const, id: first };
    const { client, rerender } = setup([favorite]);
    await screen.findByRole("link", { name: "Team" });
    fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
    updateFavorites({ ...scope, userId: "two" }, () => [favorite]);
    mockUserId = "two";
    const renderSidebar = () => sidebarUI(client);
    rerender(renderSidebar());
    expect(screen.getByRole("button", { name: "Favorites" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
    fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
    updateFavorites({ userId: "two", workspaceSlug: "other" }, () => [
      favorite,
    ]);
    mockWorkspace = "other";
    rerender(renderSidebar());
    expect(screen.getByRole("button", { name: "Favorites" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
    mockWorkspace = "acme";
    mockUserId = "one";
    rerender(renderSidebar());
    expect(screen.getByRole("button", { name: "Favorites" })).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    expect(localStorage.getItem("sidebar:acme:two:favorites-expanded")).toBe(
      "false",
    );
  });

  it("shows links in the narrow sidebar while retaining the saved closed state for wide mode", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    const { client, rerender } = setup([{ kind: "team", id: first }]);
    await screen.findByRole("link", { name: "Team" });
    fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
    const renderSidebar = (isCollapsed: boolean) =>
      sidebarUI(client, isCollapsed);
    rerender(renderSidebar(true));
    expect(screen.getByRole("link", { name: "Team" })).toBeVisible();
    expect(
      screen.queryByRole("button", { name: "Favorites" }),
    ).not.toBeInTheDocument();
    rerender(renderSidebar(false));
    expect(screen.getByRole("button", { name: "Favorites" })).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    expect(
      screen.queryByRole("link", { name: "Team" }),
    ).not.toBeInTheDocument();
  });

  it.each(["{invalid", '"closed"', "{}"])(
    "defaults safely to expanded for malformed or non-boolean preferences (%s)",
    async (stored) => {
      localStorage.setItem("sidebar:acme:one:favorites-expanded", stored);
      jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
      setup([{ kind: "team", id: first }]);
      expect(
        await screen.findByRole("link", { name: "Team" }),
      ).toBeInTheDocument();
      fireEvent.click(screen.getByRole("button", { name: "Favorites" }));
      expect(screen.getByRole("button", { name: "Favorites" })).toHaveAttribute(
        "aria-expanded",
        "false",
      );
    },
  );

  it("allows collapse and reopening in memory when preference reads and writes are blocked", async () => {
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Team"));
    const get = window.localStorage.getItem.bind(window.localStorage);
    const read = jest
      .spyOn(Storage.prototype, "getItem")
      .mockImplementation((key) => {
        if (key.endsWith(":favorites-expanded")) throw new Error("Blocked");
        return get(key);
      });
    setup([{ kind: "team", id: first }]);
    await screen.findByRole("link", { name: "Team" });
    const write = jest
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new Error("Blocked");
      });
    const trigger = screen.getByRole("button", { name: "Favorites" });
    fireEvent.click(trigger);
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    fireEvent.click(trigger);
    expect(screen.getByRole("link", { name: "Team" })).toBeVisible();
    read.mockRestore();
    write.mockRestore();
  });

  it("keeps failure and retry feedback usable while collapsed without reopening automatically", async () => {
    jest
      .mocked(getTeamDetails)
      .mockRejectedValue(new ApiError("Offline", 500, null));
    setup([{ kind: "team", id: first }]);
    const trigger = screen.getByRole("button", { name: "Favorites" });
    fireEvent.click(trigger);
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Some favorites could not be loaded.",
    );
    const retry = screen.getByRole("button", { name: "Try again" });
    await waitFor(() => {
      expect(retry).toBeEnabled();
    });
    jest.mocked(getTeamDetails).mockResolvedValue(team(first, "Recovered"));
    act(() => {
      retry.focus();
    });
    fireEvent.click(retry);
    await waitFor(() => {
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(
      screen.queryByRole("link", { name: "Recovered" }),
    ).not.toBeInTheDocument();
    fireEvent.click(trigger);
    expect(
      await screen.findByRole("link", { name: "Recovered" }),
    ).toBeVisible();
  });

  it("uses the resolved task status tint for both active and inactive favorite links", async () => {
    jest.mocked(getStory).mockResolvedValue({
      id: first,
      title: "Task",
      statusId: "status",
      teamCode: "NEW",
      sequenceId: 26,
    } as DetailedStory);
    jest
      .mocked(getStatuses)
      .mockResolvedValue([{ id: "status", color: "#3478f6" }] as Awaited<
        ReturnType<typeof getStatuses>
      >);
    mockPath = "/acme/work/NEW-26";
    const { client, rerender } = setup([{ kind: "story", id: first }]);
    const link = await screen.findByRole("link", { name: "Task" });
    await waitFor(() => {
      expect(link.querySelector("svg")).toHaveStyle({ color: "#3478f6" });
    });
    expect(link).toHaveAttribute("aria-current", "page");
    mockPath = "/acme/my-work";
    rerender(sidebarUI(client));
    expect(screen.getByRole("link", { name: "Task" })).not.toHaveAttribute(
      "aria-current",
    );
    expect(
      screen.getByRole("link", { name: "Task" }).querySelector("svg"),
    ).toHaveStyle({ color: "#3478f6" });
  });
});
