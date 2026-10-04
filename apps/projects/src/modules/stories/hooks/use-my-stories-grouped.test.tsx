import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { GroupedStoriesResponse } from "../types";
import { getGroupedStories } from "../queries/get-grouped-stories";
import { useMyStoriesGrouped } from "./use-my-stories-grouped";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "owner" } } }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("../queries/get-grouped-stories", () => ({
  getGroupedStories: jest.fn(),
}));

const response = (totalGroups: number): GroupedStoriesResponse => ({
  groups: [],
  meta: {
    totalGroups,
    filters: {},
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
  },
});

it("freshly resolves a previously loaded saved-view filter key instead of leaving its readiness gate blocked", async () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  let resolveFresh!: (data: GroupedStoriesResponse) => void;
  const fresh = new Promise<GroupedStoriesResponse>((resolve) => {
    resolveFresh = resolve;
  });
  jest
    .mocked(getGroupedStories)
    .mockResolvedValueOnce(response(1))
    .mockResolvedValueOnce(response(2))
    .mockReturnValueOnce(fresh);

  const { result, rerender, unmount } = renderHook(
    ({ statusIds }: { statusIds: string[] }) =>
      useMyStoriesGrouped("status", { statusIds }, { accountScoped: true }),
    { initialProps: { statusIds: ["in-progress"] }, wrapper },
  );
  await waitFor(() => {
    expect(result.current.isFetchedAfterMount).toBe(true);
    expect(result.current.data?.meta.totalGroups).toBe(1);
  });
  rerender({ statusIds: ["in-progress", "done"] });
  await waitFor(() => {
    expect(result.current.isFetchedAfterMount).toBe(true);
    expect(result.current.data?.meta.totalGroups).toBe(2);
  });
  rerender({ statusIds: ["in-progress"] });
  await waitFor(() => {
    expect(getGroupedStories).toHaveBeenCalledTimes(3);
  });
  expect(result.current.isFetchedAfterMount).toBe(false);
  expect(result.current.isFetching).toBe(true);
  await act(async () => {
    resolveFresh(response(3));
  });
  await waitFor(() => {
    expect(result.current.isFetchedAfterMount).toBe(true);
    expect(result.current.data?.meta.totalGroups).toBe(3);
  });
  unmount();
  client.clear();
});
