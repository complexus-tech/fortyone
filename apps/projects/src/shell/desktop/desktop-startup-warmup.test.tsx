/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, render } from "@testing-library/react";
import { DesktopStartupWarmup } from "./desktop-startup-warmup";

const mockPrefetch = jest.fn();
const mockPrefetchQuery = jest.fn().mockResolvedValue(undefined);
const mockSettingsWarmup = jest.fn().mockResolvedValue([]);
const mockRouter = { prefetch: mockPrefetch };
const mockQueryClient = { prefetchQuery: mockPrefetchQuery };
let mockDesktop = true;
let mockSession: { user: { id: string } } | null = { user: { id: "user" } };
let mockWorkspace = "acme";
jest.mock("next/navigation", () => ({ useRouter: () => mockRouter }));
jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => mockQueryClient,
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: mockSession }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: mockWorkspace,
    withWorkspace: (path: string) => `/${mockWorkspace}${path}`,
  }),
}));
jest.mock("./bridge", () => ({
  getDesktopBridge: () => (mockDesktop ? {} : null),
}));
jest.mock("./startup-queries", () => ({
  warmDesktopSettings: (...args: unknown[]) => mockSettingsWarmup(...args),
  warmDesktopMyWork: (_client: unknown, _ctx: unknown, params: unknown) =>
    mockPrefetchQuery({ queryKey: [params] }),
}));
jest.mock("@/modules/stories/queries/get-grouped-stories", () => ({
  getGroupedStories: jest.fn(),
}));

beforeEach(() => {
  jest.useFakeTimers();
  jest.clearAllMocks();
  localStorage.clear();
  delete window.__FORTYONE_OFFLINE_PAGE__;
  mockDesktop = true;
  mockSession = { user: { id: "user" } };
  mockWorkspace = "acme";
});
afterEach(() => {
  jest.useRealTimers();
});
const warm = async () => {
  await act(async () => {
    jest.runOnlyPendingTimers();
  });
};

it("does no route or data warmup in the ordinary web app", async () => {
  mockDesktop = false;
  render(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  expect(mockPrefetch).not.toHaveBeenCalled();
  expect(mockPrefetchQuery).not.toHaveBeenCalled();
});
it("waits for authentication and a workspace", async () => {
  mockSession = null;
  const view = render(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  expect(mockPrefetchQuery).not.toHaveBeenCalled();
  mockSession = { user: { id: "user" } };
  mockWorkspace = "";
  view.rerender(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  expect(mockPrefetchQuery).not.toHaveBeenCalled();
});
it("warms My Work and Settings once and respects the saved My Work view", async () => {
  localStorage.setItem("my-stories:stories:layout", JSON.stringify("list"));
  localStorage.setItem(
    "my-work:view-options:list",
    JSON.stringify({
      groupBy: "priority",
      orderBy: "updated",
      orderDirection: "asc",
      showSubStories: false,
    }),
  );
  const view = render(<DesktopStartupWarmup workspaceId="workspace" />);
  expect(mockPrefetchQuery).not.toHaveBeenCalled();
  await warm();
  expect(mockPrefetch.mock.calls).toEqual([
    ["/acme/my-work"],
    ["/acme/settings"],
  ]);
  expect(mockSettingsWarmup).toHaveBeenCalledTimes(1);
  expect(mockPrefetchQuery).toHaveBeenCalledTimes(1);
  const params = mockPrefetchQuery.mock.calls[0][0].queryKey.at(-1);
  expect(params).toMatchObject({
    groupBy: "priority",
    orderBy: "updated",
    orderDirection: "asc",
    assignedToMe: true,
    collaboratingWithMe: true,
    createdByMe: true,
  });
  expect(params.showSubStories).toBeUndefined();
  view.rerender(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  expect(mockSettingsWarmup).toHaveBeenCalledTimes(1);
  expect(mockPrefetchQuery).toHaveBeenCalledTimes(1);
});
it("cancels work on unmount and warms the newly selected workspace", async () => {
  const view = render(<DesktopStartupWarmup workspaceId="workspace" />);
  view.unmount();
  await warm();
  expect(mockPrefetchQuery).not.toHaveBeenCalled();
  const next = render(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  mockWorkspace = "product";
  next.rerender(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  expect(mockPrefetch).toHaveBeenCalledWith("/product/my-work");
  expect(mockPrefetchQuery).toHaveBeenCalledTimes(2);
});

it("waits for verified server recovery when a cached page boots on a connected network", async () => {
  window.__FORTYONE_OFFLINE_PAGE__ = true;
  render(<DesktopStartupWarmup workspaceId="workspace" />);
  await warm();
  expect(mockSettingsWarmup).not.toHaveBeenCalled();
  delete window.__FORTYONE_OFFLINE_PAGE__;
  await act(async () => {
    window.dispatchEvent(new Event("fortyone:desktop-connectivity"));
  });
  expect(mockSettingsWarmup).toHaveBeenCalledTimes(1);
});
