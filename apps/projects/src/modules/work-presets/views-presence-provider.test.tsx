import type { ReactNode } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import { getTeamsPage } from "@/modules/teams/public/queries";
import type { TeamsPage } from "@/modules/teams/types";
import { useViewsPresence } from "@/shared/views/presence-context";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { archivePreset, createPreset, listPresets } from "./api";
import { presetKey, usePresetMutations } from "./hooks";
import type { SavedView } from "./resolve-views";
import { SavedViewsPresenceProvider } from "./views-presence-provider";

let mockUserId = "owner";
let mockWorkspace = "acme";
const FIRST_TEAM = "00000000-0000-4000-8000-000000000001";
const SECOND_TEAM = "00000000-0000-4000-8000-000000000002";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspace }),
}));
jest.mock("@/modules/teams/public/queries", () => ({
  getTeamsPage: jest.fn(),
}));
jest.mock("./api", () => ({
  listPresets: jest.fn(),
  createPreset: jest.fn(),
  archivePreset: jest.fn(),
}));

const teamsPage = (
  ids = [FIRST_TEAM],
  page = 1,
  hasMore = false,
): TeamsPage => ({
  teams: ids.map((id) => ({
    id,
    name: "Product",
    code: "PRO",
    color: "#123456",
    isPrivate: false,
    workspaceId: "workspace",
    createdAt: "2026-10-01",
    updatedAt: "2026-10-01",
    memberCount: 1,
    sprintsEnabled: true,
  })),
  pagination: { page, pageSize: 15, hasMore, nextPage: hasMore ? page + 1 : 0 },
});
const view = (teamId = FIRST_TEAM): SavedView => ({
  id: "00000000-0000-4000-8000-000000000003",
  teamId,
  name: "Pipeline",
  kind: "view",
  ownerId: "owner",
  visibility: "team",
  canEdit: true,
  createdAt: "2026-10-01",
  updatedAt: "2026-10-01",
  configuration: {
    version: 1,
    layout: "list",
    filters: DEFAULT_STORIES_FILTER,
    viewOptions: {
      groupBy: "status",
      orderBy: "created",
      orderDirection: "desc",
      displayColumns: [],
      showEmptyGroups: false,
      showSubStories: false,
    },
  },
});
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
const Presence = () => {
  const { hasViews, hasTeamViews } = useViewsPresence();
  const { create, archive } = usePresetMutations(FIRST_TEAM, "view");
  return (
    <>
      <output aria-label="Workspace views">{String(hasViews)}</output>
      <output aria-label="Product views">
        {String(hasTeamViews(FIRST_TEAM))}
      </output>
      <output aria-label="Engineering views">
        {String(hasTeamViews(SECOND_TEAM))}
      </output>
      <button
        onClick={() => {
          create.mutate(view());
        }}
        type="button"
      >
        Create
      </button>
      <button
        onClick={() => {
          archive.mutate(view().id);
        }}
        type="button"
      >
        Archive
      </button>
    </>
  );
};
const setup = (
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } }),
) => {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return {
    ...render(
      <SavedViewsPresenceProvider>
        <Presence />
      </SavedViewsPresenceProvider>,
      { wrapper },
    ),
    client,
  };
};
const expectPresence = (
  workspace: boolean,
  product: boolean,
  engineering = false,
) => {
  expect(screen.getByLabelText("Workspace views")).toHaveTextContent(
    String(workspace),
  );
  expect(screen.getByLabelText("Product views")).toHaveTextContent(
    String(product),
  );
  expect(screen.getByLabelText("Engineering views")).toHaveTextContent(
    String(engineering),
  );
};

describe("authorized sidebar Views presence", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUserId = "owner";
    mockWorkspace = "acme";
    jest.mocked(getTeamsPage).mockResolvedValue(teamsPage());
    jest.mocked(listPresets).mockResolvedValue({ items: [], nextCursor: "" });
  });

  it("discovers a view beyond the first team page using one-item existence requests", async () => {
    jest
      .mocked(getTeamsPage)
      .mockResolvedValueOnce(teamsPage([FIRST_TEAM], 1, true))
      .mockResolvedValueOnce(teamsPage([SECOND_TEAM], 2));
    jest.mocked(listPresets).mockImplementation(async (teamId) => ({
      items: teamId === SECOND_TEAM ? [view(SECOND_TEAM)] : [],
      nextCursor: "",
    }));
    setup();
    expectPresence(false, false);
    await waitFor(() => {
      expectPresence(true, false, true);
    });
    expect(getTeamsPage).toHaveBeenCalledTimes(2);
    expect(listPresets).toHaveBeenCalledWith(
      SECOND_TEAM,
      "view",
      "",
      expect.objectContaining({ workspaceSlug: "acme" }),
      expect.any(AbortSignal),
      1,
    );
  });

  it("withholds cached presence until membership and the current presence request resolve", async () => {
    const membership = deferred<TeamsPage>();
    const presence = deferred<Awaited<ReturnType<typeof listPresets>>>();
    jest.mocked(getTeamsPage).mockReturnValue(membership.promise);
    jest.mocked(listPresets).mockReturnValue(presence.promise);
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    client.setQueryData([...teamKeys.lists("acme"), "views", "owner"], {
      pages: [teamsPage()],
      pageParams: [1],
    });
    client.setQueryData(
      [...presetKey("acme", "owner", FIRST_TEAM, "view"), "presence"],
      true,
    );
    setup(client);
    expectPresence(false, false);
    expect(listPresets).not.toHaveBeenCalled();
    await act(async () => {
      membership.resolve(teamsPage());
    });
    await waitFor(() => {
      expect(listPresets).toHaveBeenCalled();
    });
    expectPresence(false, false);
    await act(async () => {
      presence.resolve({ items: [], nextCursor: "" });
    });
    await waitFor(() => {
      expectPresence(false, false);
    });
  });

  it.each(["account", "workspace"])(
    "hides prior presence immediately when the %s changes",
    async (change) => {
      jest
        .mocked(listPresets)
        .mockResolvedValue({ items: [view()], nextCursor: "" });
      const { rerender } = setup();
      await waitFor(() => {
        expectPresence(true, true);
      });
      const fresh = deferred<TeamsPage>();
      jest.mocked(getTeamsPage).mockReturnValue(fresh.promise);
      if (change === "account") mockUserId = "other";
      else mockWorkspace = "other-workspace";
      rerender(
        <SavedViewsPresenceProvider>
          <Presence />
        </SavedViewsPresenceProvider>,
      );
      expectPresence(false, false);
      await act(async () => {
        fresh.resolve(teamsPage([]));
      });
      await waitFor(() => {
        expectPresence(false, false);
      });
    },
  );

  it("ignores a late team-directory response from the previous account", async () => {
    const previous = deferred<TeamsPage>();
    jest
      .mocked(getTeamsPage)
      .mockReturnValueOnce(previous.promise)
      .mockResolvedValueOnce(teamsPage([SECOND_TEAM]));
    jest
      .mocked(listPresets)
      .mockResolvedValue({ items: [view(SECOND_TEAM)], nextCursor: "" });
    const { rerender } = setup();
    mockUserId = "other";
    rerender(
      <SavedViewsPresenceProvider>
        <Presence />
      </SavedViewsPresenceProvider>,
    );
    await waitFor(() => {
      expectPresence(true, false, true);
    });
    await act(async () => {
      previous.resolve(teamsPage());
    });
    expectPresence(true, false, true);
    expect(listPresets).not.toHaveBeenCalledWith(
      FIRST_TEAM,
      expect.anything(),
      expect.anything(),
      expect.anything(),
      expect.anything(),
      expect.anything(),
    );
  });

  it("refreshes presence after existing create and archive mutation invalidation", async () => {
    let items: SavedView[] = [];
    jest
      .mocked(listPresets)
      .mockImplementation(async () => ({ items, nextCursor: "" }));
    jest.mocked(createPreset).mockImplementation(async () => {
      items = [view()];
      return view();
    });
    jest.mocked(archivePreset).mockImplementation(async () => {
      items = [];
    });
    setup();
    await waitFor(() => {
      expect(listPresets).toHaveBeenCalledTimes(1);
    });
    expectPresence(false, false);
    fireEvent.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() => {
      expectPresence(true, true);
    });
    fireEvent.click(screen.getByRole("button", { name: "Archive" }));
    await waitFor(() => {
      expectPresence(false, false);
    });
    expect(createPreset).toHaveBeenCalledTimes(1);
    expect(archivePreset).toHaveBeenCalledTimes(1);
  });

  it("removes retained positive presence when a refresh fails authorization", async () => {
    jest
      .mocked(listPresets)
      .mockResolvedValue({ items: [view()], nextCursor: "" });
    const { client } = setup();
    await waitFor(() => {
      expectPresence(true, true);
    });
    jest.mocked(listPresets).mockRejectedValue(new Error("Access denied"));
    await act(async () => {
      await client.invalidateQueries({
        queryKey: presetKey("acme", "owner", FIRST_TEAM, "view"),
      });
    });
    await waitFor(() => {
      expectPresence(false, false);
    });
  });

  it("removes presence when the current membership directory no longer contains the team", async () => {
    jest
      .mocked(listPresets)
      .mockResolvedValue({ items: [view()], nextCursor: "" });
    const { client } = setup();
    await waitFor(() => {
      expectPresence(true, true);
    });
    jest.mocked(getTeamsPage).mockResolvedValue(teamsPage([]));
    await act(async () => {
      await client.invalidateQueries({ queryKey: teamKeys.lists("acme") });
    });
    await waitFor(() => {
      expectPresence(false, false);
    });
  });
});
