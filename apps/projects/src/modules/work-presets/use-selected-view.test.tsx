import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { presetKey } from "./hooks";
import { resolveSavedViews } from "./resolve-views";
import type { SavedView } from "./resolve-views";
import { useSelectedView } from "./use-selected-view";

let mockUser = "owner";
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUser } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("./resolve-views", () => ({ resolveSavedViews: jest.fn() }));
const mockResolve = jest.mocked(resolveSavedViews);
const VIEW: SavedView = {
  id: "view",
  teamId: "team",
  ownerId: "owner",
  kind: "view",
  visibility: "personal",
  name: "Private pipeline",
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
};
const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retryDelay: 0 } },
  });
  function QueryWrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  }
  return { client, wrapper: QueryWrapper };
};

describe("fresh selected-view queries", () => {
  beforeEach(() => {
    mockUser = "owner";
    mockResolve.mockReset();
  });
  it("distinguishes cached data from a successful fresh lookup, including rejected access", async () => {
    const { client, wrapper } = setup();
    const key = [
      ...presetKey("acme", "owner", "team", "view"),
      "selected",
      "view",
      0,
    ];
    client.setQueryData(key, VIEW);
    let reject!: (error: Error) => void;
    mockResolve.mockImplementation(
      () =>
        new Promise((_resolve, rejectLookup) => {
          reject = rejectLookup;
        }),
    );
    const { result } = renderHook(() => useSelectedView("team", "view"), {
      wrapper,
    });
    expect(result.current.data).toEqual(VIEW);
    expect(result.current.isFetchedAfterMount).toBe(false);
    expect(result.current.isFetching).toBe(true);
    mockResolve.mockRejectedValue(new Error("Access denied"));
    reject(new Error("Access denied"));
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(result.current.isFetchedAfterMount).toBe(true);
    expect(result.current.data).toEqual(VIEW);
  });

  it("does not expose another account's cached definition during a scope change", async () => {
    const { wrapper } = setup();
    mockResolve.mockResolvedValueOnce([VIEW]);
    const { result, rerender } = renderHook(
      () => useSelectedView("team", "view"),
      { wrapper },
    );
    await waitFor(() => {
      expect(result.current.data).toEqual(VIEW);
    });
    mockUser = "other";
    mockResolve.mockImplementation(() => new Promise(() => {}));
    rerender();
    expect(result.current.data).toBeUndefined();
    expect(result.current.isFetchedAfterMount).toBe(false);
    expect(mockResolve).toHaveBeenLastCalledWith(
      "team",
      ["view"],
      expect.objectContaining({ session: { user: { id: "other" } } }),
      expect.any(AbortSignal),
    );
  });
});
