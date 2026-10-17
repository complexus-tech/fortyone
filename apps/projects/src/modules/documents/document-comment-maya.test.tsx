import type { ReactNode } from "react";
import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { documentKeys } from "@/shared/documents/keys";
import { useDocumentMayaRefresh } from "./document-comment-maya";
import type { DocumentMayaRun } from "./document-comment-maya";
import type { DocumentCommentThread } from "./types";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "first" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "me" } } }),
}));
jest.mock("@/lib/http", () => ({ get: jest.fn(), post: jest.fn() }));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const KEY = documentKeys.comments("first", "document");
const completed: DocumentMayaRun = {
  id: "run",
  sourceId: "source",
  requesterId: "me",
  status: "succeeded",
  replyCommentId: "reply",
  summary: "Done",
  updatedAt: "2026-10-16T12:00:00Z",
  version: 3,
};
const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  client.setQueryData(KEY, []);
  const invalidate = jest
    .spyOn(client, "invalidateQueries")
    .mockResolvedValue();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  const commentRefreshes = () =>
    invalidate.mock.calls.filter(
      ([filters]) => JSON.stringify(filters?.queryKey) === JSON.stringify(KEY),
    );
  return { client, invalidate, wrapper, commentRefreshes };
};
beforeEach(() => {
  jest.useFakeTimers();
});
afterEach(() => {
  jest.useRealTimers();
  jest.restoreAllMocks();
});
const tick = async (duration = 2_000) => {
  await act(async () => {
    await jest.advanceTimersByTimeAsync(duration);
  });
};

test("retries a published reply handoff and stops immediately once it is cached", async () => {
  const { client, wrapper, commentRefreshes } = setup();
  renderHook(
    () => {
      useDocumentMayaRefresh("document", [completed]);
    },
    {
      wrapper,
    },
  );
  expect(commentRefreshes()).toHaveLength(1);
  await tick();
  expect(commentRefreshes()).toHaveLength(2);
  client.setQueryData<DocumentCommentThread[]>(KEY, [
    { id: "thread", comments: [{ id: "reply" }] } as DocumentCommentThread,
  ]);
  await tick();
  await tick();
  expect(commentRefreshes()).toHaveLength(2);
});

test("caps retries for an inaccessible reply and does not reset the budget on repeated polls", async () => {
  const { wrapper, commentRefreshes } = setup();
  const hook = renderHook(
    ({ runs }) => {
      useDocumentMayaRefresh("document", runs);
    },
    { wrapper, initialProps: { runs: [completed] } },
  );
  await tick(16_000);
  expect(commentRefreshes()).toHaveLength(6);
  hook.rerender({ runs: [{ ...completed }] });
  await tick(6_000);
  expect(commentRefreshes()).toHaveLength(6);
});

test("does not refresh content for an active run until a reply or terminal state appears", async () => {
  const { wrapper, invalidate, commentRefreshes } = setup();
  const running: DocumentMayaRun = {
    ...completed,
    status: "running" as const,
    replyCommentId: undefined,
    version: 2,
  };
  const hook = renderHook(
    ({ run }: { run: DocumentMayaRun }) => {
      useDocumentMayaRefresh("document", [run]);
    },
    { wrapper, initialProps: { run: running } },
  );
  await tick();
  expect(invalidate).not.toHaveBeenCalled();
  hook.rerender({ run: completed });
  expect(commentRefreshes()).toHaveLength(1);
  hook.unmount();
  await tick();
  expect(commentRefreshes()).toHaveLength(1);
});
