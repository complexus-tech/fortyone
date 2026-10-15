import type { ReactNode } from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { getTeamCustomFields } from "./api";
import { customFieldKeys, useTeamCustomFields } from "./hooks";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "account" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/shell/desktop/bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./api", () => ({ getTeamCustomFields: jest.fn() }));
beforeEach(() => {
  jest.clearAllMocks();
  window.__FORTYONE_OFFLINE_PAGE__ = true;
  jest.mocked(getTeamCustomFields).mockResolvedValue([]);
});
afterEach(() => {
  delete window.__FORTYONE_OFFLINE_PAGE__;
});
const renderFields = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  client.setQueryData(
    customFieldKeys.team("acme", "team"),
    [{ id: "field", type: "text" }],
    { updatedAt: 1 },
  );
  return renderHook(() => useTeamCustomFields("team"), {
    wrapper: ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    ),
  });
};
it("uses downloaded definitions without a failed offline refetch blocking creation", () => {
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({} as NonNullable<ReturnType<typeof getDesktopBridge>>);
  const view = renderFields();
  expect(view.result.current.data).toEqual([{ id: "field", type: "text" }]);
  expect(view.result.current.isError).toBe(false);
  expect(view.result.current.isPending).toBe(false);
  expect(getTeamCustomFields).not.toHaveBeenCalled();
});
it("preserves the ordinary web app's existing query behavior", async () => {
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  renderFields();
  await waitFor(() => {
    expect(getTeamCustomFields).toHaveBeenCalledTimes(1);
  });
});
