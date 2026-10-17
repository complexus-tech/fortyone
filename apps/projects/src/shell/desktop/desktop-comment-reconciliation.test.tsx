import { act, render } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { DesktopOutboxCoordinator } from "./desktop-outbox-coordinator";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";
import { useDesktopWorkspaceScope } from "./workspace-scope";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./session-binding", () => ({ bindDesktopSession: jest.fn() }));
jest.mock("./workspace-scope", () => ({ useDesktopWorkspaceScope: jest.fn() }));
jest.mock("./desktop-provider", () => ({
  useDesktop: () => ({ isDesktop: true }),
}));
jest.mock("./connectivity", () => ({
  getDesktopOnline: () => true,
  probeDesktopConnectivity: jest.fn(),
  DESKTOP_CONNECTIVITY_EVENT: "fortyone:desktop-connectivity",
}));

const scope = {
  accountId: "alice",
  workspaceId: "workspace",
  workspaceSlug: "acme",
};
const root = {
  id: "root",
  storyId: "story",
  parentId: null,
  userId: "alice",
  comment: "@Maya hi",
  createdAt: "2026-10-17T10:00:00.000Z",
  updatedAt: "2026-10-17T10:00:00.000Z",
  subComments: [],
};
const reply = {
  ...root,
  id: "maya-reply",
  parentId: root.id,
  userId: "maya",
  comment: "Hi Alice! How can I help?",
};
const snapshot = {
  operations: [],
  completed: [
    {
      operationId: "create-root",
      kind: "comment.create",
      provisionalId: root.id,
      resourceId: root.id,
      resource: root,
    },
  ],
};
const show = (client: QueryClient) =>
  render(
    <QueryClientProvider client={client}>
      <DesktopOutboxCoordinator />
    </QueryClientProvider>,
  );

beforeEach(() => {
  jest.useFakeTimers();
  jest.mocked(useDesktopWorkspaceScope).mockReturnValue(scope);
  jest.mocked(bindDesktopSession).mockResolvedValue(true);
  jest.mocked(getDesktopBridge).mockReturnValue({
    readOutbox: jest.fn().mockResolvedValue(snapshot),
    syncOutbox: jest.fn(),
    resolveOperation: jest.fn(),
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
});
afterEach(() => {
  jest.clearAllMocks();
  jest.useRealTimers();
});

it("keeps a real reply after query notifications replay the saved desktop receipt, including after restart", async () => {
  const mainKey = storyKeys.commentsInfinite(scope.workspaceSlug, root.storyId);
  const threadKey = storyKeys.commentThread(
    scope.workspaceSlug,
    root.storyId,
    reply.id,
  );
  const client = new QueryClient();
  client.setQueryData(mainKey, {
    pages: [{ comments: [root] }],
    pageParams: [1],
  });
  const view = show(client);
  await act(async () => {
    await Promise.resolve();
  });

  const current = { ...root, subComments: [reply] };
  act(() => {
    // This is the successful response arriving after publication. The actual
    // coordinator reprojects its durable acknowledgement on the next task.
    client.setQueryData(mainKey, {
      pages: [{ comments: [current] }],
      pageParams: [1],
    });
    client.setQueryData(threadKey, current);
  });
  await act(async () => {
    jest.advanceTimersByTime(0);
  });
  expect(client.getQueryData(mainKey)).toMatchObject({
    pages: [{ comments: [current] }],
  });
  expect(client.getQueryData(threadKey)).toMatchObject(current);
  view.unmount();

  // Native receipts survive app restarts. Replaying one over a hydrated cache
  // must not erase replies a second time, even without another API response.
  const restarted = new QueryClient();
  restarted.setQueryData(mainKey, client.getQueryData(mainKey));
  const reopened = show(restarted);
  await act(async () => {
    await Promise.resolve();
  });
  expect(restarted.getQueryData(mainKey)).toMatchObject({
    pages: [{ comments: [current] }],
  });

  act(() => {
    restarted.setQueryData(mainKey, {
      pages: [{ comments: [root] }],
      pageParams: [1],
    });
  });
  await act(async () => {
    jest.advanceTimersByTime(0);
  });
  expect(restarted.getQueryData(mainKey)).toMatchObject({
    pages: [{ comments: [root] }],
  });
  reopened.unmount();
  client.clear();
  restarted.clear();
});
