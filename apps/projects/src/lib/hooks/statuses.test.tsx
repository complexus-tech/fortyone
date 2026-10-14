import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { statusKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStatuses } from "@/lib/queries/states/get-states";
import type { State } from "@/types/states";
import { useStatuses } from "./statuses";

let mockWorkspaceSlug = "acme";
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspaceSlug }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { token: "session" } }),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: jest.fn(),
}));

const cachedStatuses = [{ id: "todo", name: "To do" }] as State[];
const refreshedStatuses = [{ id: "todo", name: "Ready" }] as State[];
let queryClient: QueryClient;
const Wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

beforeEach(() => {
  mockWorkspaceSlug = "acme";
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  jest.mocked(getStatuses).mockReset().mockResolvedValue(refreshedStatuses);
});

afterEach(() => {
  queryClient.clear();
});

it("reuses fresh hydrated statuses across repeated observers", () => {
  queryClient.setQueryData(statusKeys.lists("acme"), cachedStatuses);
  const first = renderHook(() => useStatuses(), { wrapper: Wrapper });
  const second = renderHook(() => useStatuses(), { wrapper: Wrapper });

  expect(first.result.current.data).toEqual(cachedStatuses);
  expect(second.result.current.data).toEqual(cachedStatuses);
  expect(getStatuses).not.toHaveBeenCalled();
});

it("refreshes expired statuses without hiding the cached data", async () => {
  queryClient.setQueryData(statusKeys.lists("acme"), cachedStatuses, {
    updatedAt: Date.now() - DURATION_FROM_MILLISECONDS.MINUTE - 1,
  });
  const { result } = renderHook(() => useStatuses(), { wrapper: Wrapper });

  expect(result.current.data).toEqual(cachedStatuses);
  await waitFor(() => {
    expect(result.current.data).toEqual(refreshedStatuses);
  });
  expect(getStatuses).toHaveBeenCalledTimes(1);
});

it("refreshes immediately when a mutation invalidates a fresh cache", async () => {
  queryClient.setQueryData(statusKeys.lists("acme"), cachedStatuses);
  const { result } = renderHook(() => useStatuses(), { wrapper: Wrapper });
  expect(getStatuses).not.toHaveBeenCalled();

  await act(async () => {
    await queryClient.invalidateQueries({ queryKey: statusKeys.lists("acme") });
  });
  await waitFor(() => {
    expect(result.current.data).toEqual(refreshedStatuses);
  });
  expect(getStatuses).toHaveBeenCalledTimes(1);
});

it("does not reuse another workspace's statuses", async () => {
  queryClient.setQueryData(statusKeys.lists("acme"), cachedStatuses);
  const { result, rerender } = renderHook(() => useStatuses(), {
    wrapper: Wrapper,
  });
  mockWorkspaceSlug = "other-workspace";
  rerender();

  expect(result.current.data).toBeUndefined();
  await waitFor(() => {
    expect(result.current.data).toEqual(refreshedStatuses);
  });
  expect(getStatuses).toHaveBeenCalledWith({
    session: { token: "session" },
    workspaceSlug: "other-workspace",
  });
  expect(queryClient.getQueryData(statusKeys.lists("acme"))).toEqual(
    cachedStatuses,
  );
});
