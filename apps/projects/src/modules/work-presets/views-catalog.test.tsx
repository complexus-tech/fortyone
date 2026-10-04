import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { listPresets } from "./api";
import type { SavedView } from "./resolve-views";
import { ViewsCatalog } from "./views-catalog";
import { presetKey } from "./hooks";

jest.mock("./api", () => ({ listPresets: jest.fn() }));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "owner" } } }),
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
jest.mock("./view-row", () => ({
  ViewRow: ({ view, detail }: { view: SavedView; detail: string }) => (
    <div data-testid={`view-${view.id}`}>
      {view.name} · {detail}
    </div>
  ),
}));

const mockList = jest.mocked(listPresets);
const view = (number: number, extra = {}): SavedView => ({
  id: String(number),
  teamId: "team",
  ownerId: "owner",
  name: `Pipeline ${number}`,
  kind: "view",
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
      showSubStories: false,
      showEmptyGroups: true,
    },
  },
  ...extra,
});
const defaults = {
  teams: [{ id: "team", name: "Product" }],
  owners: { teammate: "Sam" },
  teamsPending: false,
  teamsError: false,
  retryTeams: jest.fn(),
  hasMoreTeams: false,
  loadingMoreTeams: false,
  loadMoreTeams: jest.fn(),
};
const wrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function QueryWrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };
};

describe("workspace Views catalog", () => {
  beforeEach(() => mockList.mockReset());
  it("loads subsequent view pages deliberately and groups personal/team results", async () => {
    const firstPage = Array.from({ length: 50 }, (_, index) => view(index + 1));
    firstPage[0] = view(1, { visibility: "personal" });
    firstPage[1] = view(2, { visibility: "personal", ownerId: "other" });
    mockList.mockImplementation(async (_team, _kind, cursor) =>
      cursor
        ? {
            items: [view(51, { ownerId: "teammate", canEdit: false })],
            nextCursor: "",
          }
        : { items: firstPage, nextCursor: "second" },
    );
    render(<ViewsCatalog {...defaults} />, { wrapper: wrapper() });
    expect(await screen.findByTestId("view-1")).toHaveTextContent(
      "Pipeline 1 · Product",
    );
    expect(screen.queryByTestId("view-2")).not.toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Personal views" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Team views" }),
    ).toBeInTheDocument();
    expect(mockList).toHaveBeenCalledTimes(1);
    fireEvent.click(
      screen.getByRole("button", { name: "Load more from Product" }),
    );
    expect(await screen.findByTestId("view-51")).toHaveTextContent(
      "Pipeline 51 · Product · Sam",
    );
    expect(mockList).toHaveBeenNthCalledWith(
      3,
      "team",
      "view",
      "second",
      expect.objectContaining({ workspaceSlug: "acme" }),
      expect.any(AbortSignal),
    );
    expect(
      screen.queryByRole("button", { name: "Load more from Product" }),
    ).not.toBeInTheDocument();
    fireEvent.change(screen.getByRole("textbox", { name: "Search views" }), {
      target: { value: "Pipeline 51" },
    });
    expect(screen.queryByTestId("view-1")).not.toBeInTheDocument();
    expect(screen.getByTestId("view-51")).toBeInTheDocument();
  });

  it("reports failed discovery without claiming there are no saved views, then retries", async () => {
    mockList
      .mockRejectedValueOnce(new Error("Network unavailable"))
      .mockResolvedValueOnce({ items: [view(1)], nextCursor: "" });
    render(<ViewsCatalog {...defaults} />, { wrapper: wrapper() });
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Views from Product could not be loaded.",
    );
    expect(
      screen.queryByText("No personal views yet."),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("No shared views yet.")).not.toBeInTheDocument();
    expect(
      screen.queryByText(/Save filters and display options/),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Try Product again" }));
    expect(await screen.findByTestId("view-1")).toBeInTheDocument();
  });

  it("does not announce definitive empty results while team or view pages remain", async () => {
    mockList.mockResolvedValue({ items: [], nextCursor: "more" });
    render(<ViewsCatalog {...defaults} hasMoreTeams />, { wrapper: wrapper() });
    await screen.findByRole("button", { name: "Load more from Product" });
    expect(
      screen.queryByText("No personal views yet."),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("No shared views yet.")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Load more teams" }));
    expect(defaults.loadMoreTeams).toHaveBeenCalledTimes(1);
  });

  it("stops cyclic page cursors and offers retry without another Load more", async () => {
    mockList.mockResolvedValue({ items: [], nextCursor: "same" });
    render(<ViewsCatalog {...defaults} />, { wrapper: wrapper() });
    fireEvent.click(
      await screen.findByRole("button", { name: "Load more from Product" }),
    );
    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(
        "Views from Product could not be loaded.",
      );
    });
    expect(mockList).toHaveBeenCalledTimes(3);
    expect(
      screen.queryByRole("button", { name: "Load more from Product" }),
    ).not.toBeInTheDocument();
  });
  it("rebuilds a loaded cursor chain after insertion without omitting the old page boundary", async () => {
    let refreshed = false;
    mockList.mockImplementation(async (_team, _kind, cursor) => {
      if (!cursor)
        return {
          items: Array.from({ length: 50 }, (_, index) =>
            view(index + (refreshed ? 0 : 1)),
          ),
          nextCursor: refreshed ? "after49" : "after50",
        };
      return {
        items: refreshed ? [view(50), view(51)] : [view(51)],
        nextCursor: "",
      };
    });
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    render(
      <QueryClientProvider client={client}>
        <ViewsCatalog {...defaults} />
      </QueryClientProvider>,
    );
    fireEvent.click(
      await screen.findByRole("button", { name: "Load more from Product" }),
    );
    await screen.findByTestId("view-51");
    refreshed = true;
    await act(async () => {
      await client.invalidateQueries({
        queryKey: presetKey("acme", "owner", "team", "view"),
      });
    });
    expect(await screen.findByTestId("view-0")).toBeInTheDocument();
    expect(screen.getByTestId("view-50")).toBeInTheDocument();
    expect(screen.getByTestId("view-51")).toBeInTheDocument();
    expect(mockList).toHaveBeenLastCalledWith(
      "team",
      "view",
      "after49",
      expect.any(Object),
      expect.any(AbortSignal),
    );
  });

  it("surfaces a first-page refresh failure after Load more and hides cached rows", async () => {
    let fail = false;
    mockList.mockImplementation(async (_team, _kind, cursor) => {
      if (fail && !cursor) throw new Error("Access denied");
      return cursor
        ? { items: [view(51)], nextCursor: "" }
        : { items: [view(1)], nextCursor: "second" };
    });
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    render(
      <QueryClientProvider client={client}>
        <ViewsCatalog {...defaults} />
      </QueryClientProvider>,
    );
    fireEvent.click(
      await screen.findByRole("button", { name: "Load more from Product" }),
    );
    await screen.findByTestId("view-51");
    fail = true;
    await act(async () => {
      await client.invalidateQueries({
        queryKey: presetKey("acme", "owner", "team", "view"),
      });
    });
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Views from Product could not be loaded.",
    );
    expect(screen.queryByTestId("view-1")).not.toBeInTheDocument();
    expect(screen.queryByTestId("view-51")).not.toBeInTheDocument();
    expect(screen.queryByText("No shared views yet.")).not.toBeInTheDocument();
    fail = false;
    fireEvent.click(screen.getByRole("button", { name: "Try Product again" }));
    expect(await screen.findByTestId("view-51")).toBeInTheDocument();
  });

  it("recovers from an expired cursor by fetching a fresh first-page cursor on retry", async () => {
    let refreshed = false;
    mockList.mockImplementation(async (_team, _kind, cursor) => {
      if (!cursor)
        return {
          items: [view(1)],
          nextCursor: refreshed ? "fresh" : "expired",
        };
      if (cursor === "expired") throw new Error("Cursor expired");
      return { items: [view(51)], nextCursor: "" };
    });
    render(<ViewsCatalog {...defaults} />, { wrapper: wrapper() });
    fireEvent.click(
      await screen.findByRole("button", { name: "Load more from Product" }),
    );
    await screen.findByRole("alert");
    refreshed = true;
    fireEvent.click(screen.getByRole("button", { name: "Try Product again" }));
    expect(await screen.findByTestId("view-51")).toBeInTheDocument();
    expect(mockList).toHaveBeenLastCalledWith(
      "team",
      "view",
      "fresh",
      expect.any(Object),
      expect.any(AbortSignal),
    );
  });
});
