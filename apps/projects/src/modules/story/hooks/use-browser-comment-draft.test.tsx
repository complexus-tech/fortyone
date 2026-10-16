import { act, renderHook } from "@testing-library/react";
import { toast } from "sonner";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import { useBrowserCommentDraft } from "./use-browser-comment-draft";

jest.mock("@/shell/desktop/workspace-scope", () => ({
  useDesktopWorkspaceScope: jest.fn(),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const scope = {
  accountId: "account-a",
  workspaceId: "workspace-a",
  workspaceSlug: "workspace-a",
};
const draft = { version: 1 as const, html: "<p>Unfinished comment</p>" };

beforeEach(() => {
  jest.restoreAllMocks();
  jest.clearAllMocks();
  localStorage.clear();
  jest.mocked(useDesktopWorkspaceScope).mockReturnValue(scope);
});

it("recovers a draft after the composer unmounts without requiring a browser notice", async () => {
  const first = renderHook(() => useBrowserCommentDraft("comment:story"));
  await act(() => first.result.current.save(draft));
  first.unmount();

  const second = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(second.result.current.ready).toBe(true);
  expect(second.result.current.recovered).toEqual(draft);

  act(() => {
    second.result.current.acceptRecovered();
  });
  expect(second.result.current.recovered).toBeNull();
  second.unmount();

  const third = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(third.result.current.recovered).toEqual(draft);
});

it("isolates drafts by account, workspace, story, and reply or edit target", async () => {
  const { result, rerender } = renderHook(
    ({ key }) => useBrowserCommentDraft(key),
    { initialProps: { key: "comment:story" } },
  );
  await act(() => result.current.save(draft));

  for (const key of [
    "comment:other-story",
    "comment:story:reply:parent",
    "comment:story:edit:comment",
  ]) {
    rerender({ key });
    expect(result.current.recovered).toBeNull();
  }

  jest.mocked(useDesktopWorkspaceScope).mockReturnValue({
    ...scope,
    accountId: "account-b",
  });
  rerender({ key: "comment:story" });
  expect(result.current.recovered).toBeNull();

  jest.mocked(useDesktopWorkspaceScope).mockReturnValue({
    ...scope,
    workspaceId: "workspace-b",
  });
  rerender({ key: "comment:story" });
  expect(result.current.recovered).toBeNull();

  jest.mocked(useDesktopWorkspaceScope).mockReturnValue(scope);
  rerender({ key: "comment:story" });
  expect(result.current.recovered).toEqual(draft);
});

it("does not discard a newer draft when a submitted composer finishes after unmount", async () => {
  const first = renderHook(() => useBrowserCommentDraft("comment:story"));
  await act(() => first.result.current.save(draft));
  const discardSubmitted = first.result.current.discard;
  first.unmount();

  const newer = { version: 1 as const, html: "<p>Next comment</p>" };
  const second = renderHook(() => useBrowserCommentDraft("comment:story"));
  await act(() => second.result.current.save(newer));
  await act(() => discardSubmitted(draft.html));
  second.unmount();

  const third = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(third.result.current.recovered).toEqual(newer);
  await act(() => third.result.current.discard(newer.html));
  third.unmount();

  const fourth = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(fourth.result.current.recovered).toBeNull();
});

it("does not resurrect a recovered draft after a new editor update", async () => {
  const first = renderHook(() => useBrowserCommentDraft("comment:story"));
  await act(() => first.result.current.save(draft));
  first.unmount();

  const second = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(second.result.current.recovered).toEqual(draft);
  await act(() =>
    second.result.current.save({ version: 1, html: "<p>New text</p>" }),
  );
  expect(second.result.current.recovered).toBeNull();
});

it.each([
  "not valid JSON",
  JSON.stringify({ version: 2, html: "<p>Unknown version</p>" }),
  JSON.stringify({ version: 1, html: 123 }),
])("ignores malformed or unsupported saved drafts: %s", async (stored) => {
  const first = renderHook(() => useBrowserCommentDraft("comment:story"));
  await act(() => first.result.current.save(draft));
  first.unmount();
  localStorage.setItem(localStorage.key(0)!, stored);

  const second = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(second.result.current.ready).toBe(true);
  expect(second.result.current.recovered).toBeNull();
});

it("allows composing when browser storage cannot be read", () => {
  jest.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
    throw new Error("Storage blocked");
  });

  const { result } = renderHook(() => useBrowserCommentDraft("comment:story"));
  expect(result.current.ready).toBe(true);
  expect(result.current.recovered).toBeNull();
  expect(result.current.error).toContain("could not be recovered");
});

it("reports a failed write and clears the error after storage recovers", async () => {
  const { result } = renderHook(() => useBrowserCommentDraft("comment:story"));
  const write = jest
    .spyOn(Storage.prototype, "setItem")
    .mockImplementation(() => {
      throw new Error("Storage full");
    });
  await act(async () => {
    await expect(result.current.save(draft)).rejects.toThrow("Storage full");
  });
  expect(result.current.ready).toBe(true);
  expect(result.current.error).toContain("could not be saved");
  await act(async () => {
    await expect(result.current.save(draft)).rejects.toThrow("Storage full");
  });
  expect(toast.error).toHaveBeenCalledTimes(1);

  write.mockRestore();
  await act(() => result.current.save(draft));
  expect(result.current.error).toBeNull();
});

it("does not read or write browser storage when disabled or outside an authenticated workspace", async () => {
  const read = jest.spyOn(Storage.prototype, "getItem");
  const write = jest.spyOn(Storage.prototype, "setItem");
  const { result, rerender } = renderHook(
    ({ enabled }) => useBrowserCommentDraft("comment:story", enabled),
    { initialProps: { enabled: false } },
  );
  expect(result.current.ready).toBe(true);
  await act(() => result.current.save(draft));
  await act(() => result.current.discard());

  jest.mocked(useDesktopWorkspaceScope).mockReturnValue(null);
  rerender({ enabled: true });
  expect(result.current.ready).toBe(true);
  await act(() => result.current.save(draft));
  await act(() => result.current.discard());
  expect(read).not.toHaveBeenCalled();
  expect(write).not.toHaveBeenCalled();
});
