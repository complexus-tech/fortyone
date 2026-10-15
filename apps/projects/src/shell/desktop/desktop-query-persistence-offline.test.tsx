import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, waitFor } from "@testing-library/react";
import { getDesktopBridge } from "./bridge";
import { createDesktopCache } from "./query-cache";
import { clearDesktopBindings } from "./session-binding";
import { DesktopQueryPersistence } from "./desktop-query-persistence";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./connectivity", () => ({ getDesktopOnline: () => false }));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "alice" } } }),
}));

it("hydrates downloaded tasks and teams without waiting for an unresponsive server", async () => {
  clearDesktopBindings();
  const scope = {
    accountId: "alice",
    workspaceId: "one",
    workspaceSlug: "first",
  };
  const source = new QueryClient();
  const storyKey = ["stories", "first", "my-work"];
  const teamKey = ["teams", "first", "list"];
  source.setQueryData(storyKey, [{ id: "story", title: "Downloaded work" }]);
  source.setQueryData(teamKey, [{ id: "team", name: "Engineering" }]);
  const bindSession = jest.fn(
    () =>
      new Promise<void>(() => {
        /* Simulate a paused server that never responds. */
      }),
  );
  const cacheRead = jest
    .fn()
    .mockResolvedValue(createDesktopCache(source, scope));
  const restoreSession = jest
    .fn()
    .mockResolvedValue({ ...scope, expiresAt: Date.now() + 60_000 });
  jest.mocked(getDesktopBridge).mockReturnValue({
    bindSession,
    restoreSession,
    cacheRead,
    cacheWrite: jest.fn().mockResolvedValue(undefined),
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
  const client = new QueryClient();
  const view = render(
    <QueryClientProvider client={client}>
      <DesktopQueryPersistence workspaceId="one" workspaceSlug="first" />
    </QueryClientProvider>,
  );
  await waitFor(() => {
    expect(client.getQueryData(storyKey)).toEqual(
      source.getQueryData(storyKey),
    );
    expect(client.getQueryData(teamKey)).toEqual(source.getQueryData(teamKey));
  });
  expect(bindSession).not.toHaveBeenCalled();
  expect(restoreSession).toHaveBeenCalledWith({ workspaceSlug: "first" });
  expect(cacheRead).toHaveBeenCalledWith({
    accountId: "alice",
    workspaceId: "one",
    version: 1,
  });
  view.unmount();
  clearDesktopBindings();
});
