import type * as NodeCrypto from "node:crypto";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useDocumentCommentDraft } from "./use-document-comment-draft";

jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: jest.fn(),
}));
jest.mock("@/shell/desktop/workspace-scope", () => ({
  useDesktopWorkspaceScope: () => null,
}));

beforeEach(() => {
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: jest.requireActual<typeof NodeCrypto>("node:crypto").randomUUID,
  });
  localStorage.clear();
  jest
    .mocked(useSession)
    .mockReturnValue({ data: { user: { id: "alice" } } } as ReturnType<
      typeof useSession
    >);
  jest
    .mocked(useWorkspacePath)
    .mockReturnValue({ workspaceSlug: "first", withWorkspace: (path) => path });
});
afterEach(() => jest.restoreAllMocks());

test("restores a scoped draft after unmount and keeps document, thread, workspace and account drafts separate", async () => {
  const first = renderHook(() =>
    useDocumentCommentDraft("doc", "reply:thread"),
  );
  await waitFor(() => {
    expect(first.result.current.ready).toBe(true);
  });
  act(() => {
    first.result.current.save("A saved reply");
  });
  first.unmount();
  const second = renderHook(
    ({ doc, target }) => useDocumentCommentDraft(doc, target),
    { initialProps: { doc: "doc", target: "reply:thread" } },
  );
  await waitFor(() => {
    expect(second.result.current.body).toBe("A saved reply");
  });
  second.rerender({ doc: "other-doc", target: "reply:thread" });
  expect(second.result.current.body).toBe("");
  second.rerender({ doc: "doc", target: "reply:other-thread" });
  expect(second.result.current.body).toBe("");
  jest.mocked(useWorkspacePath).mockReturnValue({
    workspaceSlug: "second",
    withWorkspace: (path) => path,
  });
  second.rerender({ doc: "doc", target: "reply:thread" });
  expect(second.result.current.body).toBe("");
  jest
    .mocked(useWorkspacePath)
    .mockReturnValue({ workspaceSlug: "first", withWorkspace: (path) => path });
  jest
    .mocked(useSession)
    .mockReturnValue({ data: { user: { id: "bob" } } } as ReturnType<
      typeof useSession
    >);
  second.rerender({ doc: "doc", target: "reply:thread" });
  expect(second.result.current.body).toBe("");
});

test("a completed request cannot remove a newer draft saved by a remounted composer", async () => {
  const first = renderHook(() =>
    useDocumentCommentDraft("doc", "selection:stable"),
  );
  await waitFor(() => {
    expect(first.result.current.ready).toBe(true);
  });
  act(() => {
    first.result.current.save("Submitted");
  });
  const clearSubmitted = first.result.current.clear;
  first.unmount();
  const next = renderHook(() =>
    useDocumentCommentDraft("doc", "selection:stable"),
  );
  await waitFor(() => {
    expect(next.result.current.body).toBe("Submitted");
  });
  act(() => {
    next.result.current.save("Newer draft");
  });
  act(() => {
    clearSubmitted("Submitted");
  });
  next.unmount();
  const restored = renderHook(() =>
    useDocumentCommentDraft("doc", "selection:stable"),
  );
  await waitFor(() => {
    expect(restored.result.current.body).toBe("Newer draft");
  });
  act(() => {
    restored.result.current.clear("Newer draft");
  });
  expect(localStorage.length).toBe(0);
});

test("waits for a stable identity and reports storage errors while retaining typed content", async () => {
  jest
    .mocked(useSession)
    .mockReturnValue({ data: null } as ReturnType<typeof useSession>);
  const draft = renderHook(() => useDocumentCommentDraft("doc", "selection"));
  expect(draft.result.current.ready).toBe(false);
  act(() => {
    draft.result.current.save("Unscoped");
  });
  expect(localStorage.length).toBe(0);
  jest
    .mocked(useSession)
    .mockReturnValue({ data: { user: { id: "alice" } } } as ReturnType<
      typeof useSession
    >);
  draft.rerender();
  await waitFor(() => {
    expect(draft.result.current.ready).toBe(true);
  });
  jest.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
    throw new Error("Quota exceeded");
  });
  act(() => {
    draft.result.current.save("Do not lose me");
  });
  expect(draft.result.current.body).toBe("Do not lose me");
  expect(draft.result.current.error).toContain("could not be saved");
});

test("restores the original reply target and creation ID across composer remounts", async () => {
  const first = renderHook(() =>
    useDocumentCommentDraft("reply-doc", "reply:thread"),
  );
  await waitFor(() => {
    expect(first.result.current.ready).toBe(true);
  });
  let creationId: string | undefined;
  act(() => {
    creationId = first.result.current.save(
      "Yes, make that change",
      "maya-question",
    );
  });
  first.unmount();
  const restored = renderHook(() =>
    useDocumentCommentDraft("reply-doc", "reply:thread"),
  );
  await waitFor(() => {
    expect(restored.result.current.replyToCommentId).toBe("maya-question");
  });
  expect(restored.result.current.creationId).toBe(creationId);
  act(() => {
    expect(
      restored.result.current.save(
        "Yes, make that change",
        restored.result.current.replyToCommentId,
      ),
    ).toBe(creationId);
  });
});

test("a completed request cannot clear the same text addressed to a different comment", async () => {
  const draft = renderHook(() =>
    useDocumentCommentDraft("reply-target-doc", "reply:thread"),
  );
  await waitFor(() => {
    expect(draft.result.current.ready).toBe(true);
  });
  let sentId: string | undefined;
  act(() => {
    sentId = draft.result.current.save("Agreed", "original-question");
  });
  act(() => {
    draft.result.current.save("Agreed", "different-question");
  });
  expect(draft.result.current.creationId).not.toBe(sentId);
  act(() => {
    draft.result.current.clear("Agreed", sentId);
  });
  expect(draft.result.current.body).toBe("Agreed");
  expect(draft.result.current.replyToCommentId).toBe("different-question");
});
