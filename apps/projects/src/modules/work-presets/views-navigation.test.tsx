import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getTeamDetails, getTeamsPage } from "@/modules/teams/public/queries";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { listPresets } from "./api";
import type { SavedView } from "./resolve-views";
import { presetKey } from "./hooks";
import { ViewsPage } from "./views-page";
import { ViewsSwitcher } from "./views-switcher";

let mockUserId = "owner";
let mockTeamId: string | null = null;
const mockReplace = jest.fn();
const firstTeam = "00000000-0000-4000-8000-000000000001";
const secondTeam = "00000000-0000-4000-8000-000000000002";
jest.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
  usePathname: () => "/acme/views",
  useSearchParams: () =>
    new URLSearchParams(mockTeamId ? { team: mockTeamId } : {}),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: "acme",
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "member" }),
}));
jest.mock("@/components/shared/header-container", () => ({
  HeaderContainer: ({ children }: { children: React.ReactNode }) => (
    <header>{children}</header>
  ),
}));
jest.mock("@/components/shared/mobile-menu", () => ({
  MobileMenuButton: () => null,
}));
jest.mock("@/modules/teams/public/queries", () => ({
  getTeamsPage: jest.fn(),
  getTeamDetails: jest.fn(),
}));
jest.mock("./api", () => ({ listPresets: jest.fn() }));
const saved = (
  teamId = firstTeam,
  name = "Product board",
  id = "00000000-0000-4000-8000-000000000003",
): SavedView => ({
  id,
  teamId,
  name,
  kind: "view",
  ownerId: "owner",
  visibility: "team",
  canEdit: true,
  createdAt: "2026-10-04",
  updatedAt: "2026-10-04",
  configuration: {
    version: 1,
    layout: "kanban",
    filters: DEFAULT_STORIES_FILTER,
    viewOptions: {
      groupBy: "status",
      orderBy: "created",
      orderDirection: "desc",
      displayColumns: [],
      showEmptyGroups: true,
      showSubStories: false,
    },
  },
});
const teamsPage = (ids = [firstTeam], hasMore = false, page = 1) => ({
  teams: ids.map((id) => ({
    id,
    name: id === firstTeam ? "Product" : "Engineering",
  })),
  pagination: { page, pageSize: 15, hasMore, nextPage: hasMore ? page + 1 : 0 },
});
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
const setup = (
  child = <ViewsPage />,
  client = new QueryClient({
    defaultOptions: { queries: { retry: false, retryDelay: 0 } },
  }),
) => ({
  ...render(<QueryClientProvider client={client}>{child}</QueryClientProvider>),
  client,
});

describe("saved Views navigation", () => {
  beforeAll(() => {
    global.ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
    Element.prototype.scrollIntoView = jest.fn();
  });
  beforeEach(() => {
    jest.clearAllMocks();
    mockUserId = "owner";
    mockTeamId = null;
    jest
      .mocked(getTeamsPage)
      .mockResolvedValue(
        teamsPage() as Awaited<ReturnType<typeof getTeamsPage>>,
      );
    jest
      .mocked(listPresets)
      .mockResolvedValue({ items: [saved()], nextCursor: "" });
  });

  it("waits for fresh authorization before opening a cached first view", async () => {
    const fresh = deferred<Awaited<ReturnType<typeof listPresets>>>();
    jest.mocked(listPresets).mockReturnValue(fresh.promise);
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    client.setQueryData(
      [...presetKey("acme", "owner", firstTeam, "view"), "catalog", 1],
      [{ items: [saved(firstTeam, "Cached private name")], nextCursor: "" }],
    );
    const { container } = setup(<ViewsPage />, client);
    await waitFor(() => {
      expect(listPresets).toHaveBeenCalled();
    });
    expect(mockReplace).not.toHaveBeenCalled();
    expect(screen.queryByText(/Cached private name/)).not.toBeInTheDocument();
    expect(screen.getAllByRole("status")).toHaveLength(1);
    expect(container.querySelector("[aria-hidden][inert]")).toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    await act(async () => {
      fresh.resolve({ items: [saved()], nextCursor: "" });
    });
    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith(
        `/acme/teams/${firstTeam}/stories?view=${saved().id}`,
      );
    });
    expect(mockReplace).toHaveBeenCalledTimes(1);
    expect(container.querySelector("[aria-hidden][inert]")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(
      "Opening Product board...",
    );
  });

  it("discovers the first available view beyond an empty team page", async () => {
    jest
      .mocked(getTeamsPage)
      .mockResolvedValueOnce(
        teamsPage([firstTeam], true) as Awaited<
          ReturnType<typeof getTeamsPage>
        >,
      )
      .mockResolvedValueOnce(
        teamsPage([secondTeam], false, 2) as Awaited<
          ReturnType<typeof getTeamsPage>
        >,
      );
    jest.mocked(listPresets).mockImplementation(async (teamId) => ({
      items:
        teamId === firstTeam ? [] : [saved(secondTeam, "Engineering queue")],
      nextCursor: "",
    }));
    setup();
    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith(
        `/acme/teams/${secondTeam}/stories?view=${saved().id}`,
      );
    });
    expect(getTeamsPage).toHaveBeenLastCalledWith(
      expect.objectContaining({ workspaceSlug: "acme" }),
      "",
      2,
      15,
    );
  });

  it("does not navigate from an older account's delayed discovery after the account changes", async () => {
    const old = deferred<Awaited<ReturnType<typeof listPresets>>>();
    const current = deferred<Awaited<ReturnType<typeof listPresets>>>();
    jest
      .mocked(listPresets)
      .mockImplementation(() =>
        mockUserId === "owner" ? old.promise : current.promise,
      );
    const { client, rerender } = setup();
    await waitFor(() => {
      expect(listPresets).toHaveBeenCalledTimes(1);
    });
    mockUserId = "another";
    rerender(
      <QueryClientProvider client={client}>
        <ViewsPage />
      </QueryClientProvider>,
    );
    await waitFor(() => {
      expect(listPresets).toHaveBeenCalledTimes(2);
    });
    await act(async () => {
      old.resolve({
        items: [saved(firstTeam, "Previous account")],
        nextCursor: "",
      });
    });
    expect(mockReplace).not.toHaveBeenCalled();
    await act(async () => {
      current.resolve({
        items: [saved(firstTeam, "Current account", secondTeam)],
        nextCursor: "",
      });
    });
    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith(
        `/acme/teams/${firstTeam}/stories?view=${secondTeam}`,
      );
    });
    expect(mockReplace).toHaveBeenCalledTimes(1);
  });

  it("uses a team-scoped entry without discovering other teams", async () => {
    mockTeamId = firstTeam;
    jest
      .mocked(getTeamDetails)
      .mockResolvedValue(
        teamsPage().teams[0] as Awaited<ReturnType<typeof getTeamDetails>>,
      );
    setup();
    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledTimes(1);
    });
    expect(getTeamsPage).not.toHaveBeenCalled();
    expect(getTeamDetails).toHaveBeenCalledWith(
      firstTeam,
      expect.any(Object),
      expect.any(AbortSignal),
    );
  });

  it("does not redirect or claim an empty workspace when discovery fails, and recovers on retry", async () => {
    jest.mocked(listPresets).mockRejectedValue(new Error("Offline"));
    setup();
    const retry = await screen.findByRole("button", {
      name: "Try Product again",
    });
    expect(mockReplace).not.toHaveBeenCalled();
    expect(screen.queryByText("No saved views yet")).not.toBeInTheDocument();
    jest
      .mocked(listPresets)
      .mockResolvedValue({ items: [saved()], nextCursor: "" });
    fireEvent.click(retry);
    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledTimes(1);
    });
  });

  it("shows a definitive empty state pointing to filter-toolbar Save as without redirect loops", async () => {
    jest.mocked(listPresets).mockResolvedValue({ items: [], nextCursor: "" });
    setup();
    expect(
      await screen.findByRole("heading", { name: "No saved views yet" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Create view" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Create view" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText(/Apply filters in My Work/)).toBeInTheDocument();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it("searches across accessible teams and returns keyboard focus after selection", async () => {
    jest
      .mocked(getTeamsPage)
      .mockResolvedValue(
        teamsPage([firstTeam, secondTeam]) as Awaited<
          ReturnType<typeof getTeamsPage>
        >,
      );
    jest.mocked(listPresets).mockImplementation(async (teamId) => ({
      items: [
        saved(
          teamId,
          teamId === firstTeam ? "Product board" : "Release queue",
          teamId,
        ),
      ],
      nextCursor: "",
    }));
    const select = jest.fn();
    setup(<ViewsSwitcher onSelect={select} selected={saved()} />);
    const trigger = screen.getByRole("button", {
      name: "Switch view: Product board",
    });
    const heading = screen.getByRole("heading", {
      name: "Product board",
      level: 1,
    });
    expect(heading).toContainElement(trigger);
    expect(trigger.closest("a")).toBeNull();
    fireEvent.click(trigger);
    const search = screen.getByRole("combobox", { name: "Search views" });
    await screen.findByRole("option", { name: /Release queue/ });
    fireEvent.change(search, { target: { value: "Engineering" } });
    expect(
      screen.queryByRole("option", { name: /Product board/ }),
    ).not.toBeInTheDocument();
    fireEvent.keyDown(search, { key: "ArrowDown" });
    fireEvent.keyDown(search, { key: "Enter" });
    expect(select).toHaveBeenCalledWith(
      expect.objectContaining({ teamId: secondTeam, name: "Release queue" }),
    );
    await waitFor(() => {
      expect(trigger).toHaveFocus();
    });
  });
});
