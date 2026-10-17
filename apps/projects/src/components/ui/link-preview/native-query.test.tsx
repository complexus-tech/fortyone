/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  useMutation,
} from "@tanstack/react-query";
import { storyKeys } from "@/shared/story/cache-keys";
import { documentKeys } from "@/shared/documents/keys";
import { get } from "@/lib/http";
import { useSession } from "@/lib/auth/client";
import { PREVIEW_CACHE_TTL_MS } from "@/lib/queries/links/preview-cache";
import { useNativeLink, fetchNativeLink } from "./native-query";
import type { NativeStorySummary } from "./native-query";
import type { NativeLinkTarget } from "./native-target";

jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("@/lib/http", () => ({ get: jest.fn() }));

const story = {
  id: "8c86611a-a44b-4be1-9bea-b537d52280b4",
  title: "Private launch plan",
  teamId: "team",
  deletedAt: null,
  archivedAt: null,
} as NativeStorySummary;
const target: NativeLinkTarget = {
  href: "https://other.fortyone.app/work/ENG-42",
  workspaceSlug: "other",
  kind: "story",
  id: "ENG-42",
};
const response = (value = story) => ({
  data: {
    kind: "story",
    value,
    version: "v1",
    refreshedAt: "2026-10-17T10:00:00Z",
  },
});
let client: QueryClient;
const session = (id: string) =>
  ({ data: { user: { id } }, status: "authenticated" }) as ReturnType<
    typeof useSession
  >;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);

beforeEach(() => {
  localStorage.clear();
  client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  jest.mocked(useSession).mockReturnValue(session("reader-a"));
  jest.mocked(get).mockResolvedValue(response());
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.resetAllMocks();
});

const loaded = async (result: {
  current: ReturnType<typeof useNativeLink>;
}) => {
  await waitFor(() => {
    expect(result.current.data).not.toBeNull();
  });
};

describe("cached native preview summaries", () => {
  it("recovers from a first transient failure without requiring a tab switch", async () => {
    jest
      .mocked(get)
      .mockRejectedValueOnce(new Error("Temporary connection failure"));
    const { result } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    expect(result.current.loading).toBe(true);
    await loaded(result);
    expect(result.current.unavailable).toBe(false);
    expect(result.current.data).toMatchObject({
      value: { title: story.title },
    });
    expect(get).toHaveBeenCalledTimes(2);
  });

  it("invalidates a persisted preview when its story changes after the card unmounts", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    await act(async () => {
      await client.invalidateQueries({
        queryKey: storyKeys.detail("other", story.id),
      });
    });
    client.clear();
    client = new QueryClient();
    jest
      .mocked(get)
      .mockResolvedValue(response({ ...story, title: "Changed elsewhere" }));
    const next = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(next.result.current.data).toBeNull();
    await loaded(next.result);
    expect(next.result.current.data).toMatchObject({
      value: { title: "Changed elsewhere" },
    });
    expect(get).toHaveBeenCalledTimes(2);
  });

  it("restores a successful summary after a browser reload without regenerating it", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    client.clear();
    client = new QueryClient();
    const restored = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(restored.result.current.data).toMatchObject({
      value: { title: story.title },
    });
    expect(get).toHaveBeenCalledTimes(1);
    expect(
      client.getQueryData(storyKeys.detail("other", story.id)),
    ).toBeUndefined();
  });
  it("keeps a restored preview visible when a manual refresh is temporarily offline", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    client.clear();
    client = new QueryClient();
    jest.mocked(get).mockRejectedValue(new Error("Offline"));
    const restored = renderHook(() => useNativeLink(target, true), { wrapper });
    await act(async () => {
      await expect(restored.result.current.refresh()).rejects.toThrow(
        "Offline",
      );
    });
    expect(restored.result.current.data).toMatchObject({
      value: { title: story.title },
    });
    expect(restored.result.current.unavailable).toBe(false);
  });
  it("evicts a denied preview from browser storage so it cannot reappear on reload", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    jest
      .mocked(get)
      .mockRejectedValue(
        Object.assign(new Error("Forbidden"), { status: 403 }),
      );
    await act(async () => {
      await expect(first.result.current.refresh()).rejects.toThrow(
        "no longer available",
      );
    });
    first.unmount();
    client.clear();
    client = new QueryClient();
    const restored = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(restored.result.current.data).toBeNull();
    await waitFor(() => {
      expect(restored.result.current.unavailable).toBe(true);
    });
    expect(get).toHaveBeenCalledTimes(3);
  });
  it("clears browser persistence when canonical access is revoked", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    jest
      .mocked(get)
      .mockRejectedValue(
        Object.assign(new Error("Forbidden"), { status: 403 }),
      );
    await act(async () => {
      await client
        .fetchQuery({
          queryKey: storyKeys.detail("other", story.id),
          queryFn: () =>
            Promise.reject(
              Object.assign(new Error("Forbidden"), { status: 403 }),
            ),
          retry: false,
        })
        .catch(() => undefined);
    });
    await waitFor(() => {
      expect(first.result.current.data).toBeNull();
    });
    first.unmount();
    client.clear();
    client = new QueryClient();
    const restored = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(restored.result.current.data).toBeNull();
    await waitFor(() => {
      expect(restored.result.current.unavailable).toBe(true);
    });
  });
  it("does not reuse persisted summaries across workspaces or identities", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    client.clear();
    client = new QueryClient();
    jest
      .mocked(get)
      .mockRejectedValue(
        Object.assign(new Error("Forbidden"), { status: 403 }),
      );
    const otherWorkspace = renderHook(
      () => useNativeLink({ ...target, workspaceSlug: "another" }, true),
      { wrapper },
    );
    expect(otherWorkspace.result.current.data).toBeNull();
    await waitFor(() => {
      expect(otherWorkspace.result.current.unavailable).toBe(true);
    });
    jest.mocked(useSession).mockReturnValue(session("reader-b"));
    const otherUser = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    expect(otherUser.result.current.data).toBeNull();
    await waitFor(() => {
      expect(otherUser.result.current.unavailable).toBe(true);
    });
  });

  it("defers until visible, requests a small summary, and never seeds partial detail data", async () => {
    const { result, rerender } = renderHook(
      ({ enabled }) => useNativeLink(target, enabled),
      { initialProps: { enabled: false }, wrapper },
    );
    expect(get).not.toHaveBeenCalled();
    rerender({ enabled: true });
    await loaded(result);
    expect(get).toHaveBeenCalledWith(
      "previews/native/story/ENG-42",
      { session: session("reader-a").data, workspaceSlug: "other" },
      expect.objectContaining({ cache: "no-store", retry: 0 }),
    );
    expect(get).toHaveBeenCalledTimes(1);
    expect(
      client.getQueryData(storyKeys.detail("other", story.id)),
    ).toBeUndefined();
  });
  it("reuses the identity-scoped summary on remount inside the access lease", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    const next = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(next.result.current.data).toMatchObject({
      value: { title: story.title },
    });
    await loaded(next.result);
    expect(get).toHaveBeenCalledTimes(1);
  });
  it("forces the shared request when refresh is clicked on an earlier duplicate card", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    const second = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(second.result);
    await act(async () => {
      await first.result.current.refresh();
    });
    expect(get).toHaveBeenLastCalledWith(
      expect.stringContaining("refresh=true"),
      expect.anything(),
      expect.anything(),
    );
  });
  it("does not reveal an expired cached grant on a new mount before reauthorization", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    const key = [
      "native-link-preview",
      "reader-a",
      "other",
      "story",
      "ENG-42",
      "",
    ];
    client.setQueryData(
      key,
      {
        resource: response().data,
        requestedAt: Date.now() - PREVIEW_CACHE_TTL_MS - 1,
      },
      { updatedAt: Date.now() - PREVIEW_CACHE_TTL_MS - 1 },
    );
    jest.mocked(get).mockRejectedValue(new Error("Offline"));
    const next = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(next.result.current.data).toBeNull();
    await waitFor(() => {
      expect(next.result.current.refreshError?.message).toBe("Offline");
    });
    expect(next.result.current.data).toBeNull();
  });
  it("shows loading instead of unavailable while reauthorizing an expired cached preview", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    client.setQueryData(
      ["native-link-preview", "reader-a", "other", "story", "ENG-42", ""],
      {
        resource: response().data,
        requestedAt: Date.now() - PREVIEW_CACHE_TTL_MS - 1,
      },
      { updatedAt: Date.now() - PREVIEW_CACHE_TTL_MS - 1 },
    );
    let resolveRead: ((value: unknown) => void) | undefined;
    jest.mocked(get).mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveRead = resolve;
        }),
    );
    const next = renderHook(() => useNativeLink(target, true), { wrapper });
    expect(next.result.current.data).toBeNull();
    expect(next.result.current.loading).toBe(true);
    expect(next.result.current.unavailable).toBe(false);
    await act(async () => {
      resolveRead?.(response());
    });
    await loaded(next.result);
    expect(next.result.current.loading).toBe(false);
    expect(next.result.current.unavailable).toBe(false);
  });
  it("does not show an access failure while a cached card awaits its visibility observer", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    first.unmount();
    // Ensure the cached response predates this new card mount.
    client.setQueryData(
      ["native-link-preview", "reader-a", "other", "story", "ENG-42", ""],
      { resource: response().data, requestedAt: Date.now() - 1_000 },
      { updatedAt: Date.now() - 1_000 },
    );
    const { result, rerender } = renderHook(
      ({ active }) => useNativeLink(target, true, active),
      { initialProps: { active: false }, wrapper },
    );
    expect(result.current.data).toMatchObject({
      value: { title: story.title },
    });
    expect(result.current.loading).toBe(false);
    expect(result.current.unavailable).toBe(false);
    rerender({ active: true });
    await loaded(result);
    expect(result.current.loading).toBe(false);
    expect(result.current.unavailable).toBe(false);
    expect(get).toHaveBeenCalledTimes(1);
  });
  it("shares optimistic detail updates and refetches summary on canonical invalidation", async () => {
    const { result } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    await loaded(result);
    act(() => {
      client.setQueryData(
        storyKeys.detail("other", story.id),
        { ...story, title: "Edited in dialog" },
        { updatedAt: Date.now() + 1 },
      );
    });
    await waitFor(() => {
      expect(result.current.data).toMatchObject({
        value: { title: "Edited in dialog" },
      });
    });
    jest
      .mocked(get)
      .mockResolvedValue(response({ ...story, title: "Fresh server value" }));
    await act(async () => {
      await client.invalidateQueries({
        queryKey: storyKeys.detail("other", story.id),
      });
    });
    await waitFor(() => {
      expect(result.current.data).toMatchObject({
        value: { title: "Fresh server value" },
      });
    });
  });
  it("never reveals a prior identity's canonical value before new authorization", async () => {
    const { result, rerender } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    await loaded(result);
    jest.mocked(useSession).mockReturnValue(session("reader-b"));
    jest
      .mocked(get)
      .mockRejectedValue(
        Object.assign(new Error("Forbidden"), { status: 403 }),
      );
    rerender();
    expect(result.current.data).toBeNull();
    await waitFor(() => {
      expect(result.current.unavailable).toBe(true);
    });
    expect(result.current.data).toBeNull();
  });
  it("cancels a pre-edit read once when duplicate cards receive a later invalidation", async () => {
    const first = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(first.result);
    const second = renderHook(() => useNativeLink(target, true), { wrapper });
    await loaded(second.result);
    const pendingReads: {
      resolve: (value: unknown) => void;
      signal: AbortSignal;
    }[] = [];
    jest.mocked(get).mockImplementation((_path, _context, options) => {
      return new Promise((resolve) => {
        pendingReads.push({ resolve, signal: options!.signal! });
      });
    });
    act(() => {
      first.result.current.retry();
    });
    await waitFor(() => {
      expect(pendingReads).toHaveLength(1);
    });
    act(() => {
      void client.invalidateQueries({
        queryKey: storyKeys.detail("other", story.id),
      });
    });
    await waitFor(() => {
      expect(pendingReads).toHaveLength(2);
    });
    expect(pendingReads[0].signal.aborted).toBe(true);
    expect(pendingReads[1].signal.aborted).toBe(false);
    await act(async () => {
      pendingReads[1].resolve(response({ ...story, title: "After edit" }));
      pendingReads[0].resolve(response());
    });
    await waitFor(() => {
      for (const preview of [first.result, second.result]) {
        expect(preview.current.data).toMatchObject({
          value: { title: "After edit" },
        });
      }
    });
    expect(get).toHaveBeenCalledTimes(3);
  });
  it("does not let a late access response overwrite an in-flight optimistic edit", async () => {
    const { result } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    await loaded(result);
    let resolveRead: ((value: unknown) => void) | undefined;
    jest.mocked(get).mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveRead = resolve;
        }),
    );
    act(() => {
      result.current.retry();
    });
    await waitFor(() => {
      expect(resolveRead).toBeDefined();
    });
    act(() => {
      client.setQueryData(
        storyKeys.detail("other", story.id),
        { ...story, title: "New optimistic title" },
        { updatedAt: Date.now() + 10 },
      );
    });
    await act(async () => {
      resolveRead?.(response());
    });
    await waitFor(() => {
      expect(result.current.data).toMatchObject({
        value: { title: "New optimistic title" },
      });
    });
  });
  it("forces refresh, keeps the old card visible while pending, and rejects temporary failure", async () => {
    const { result } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    await loaded(result);
    let rejectRead: ((error: Error) => void) | undefined;
    jest.mocked(get).mockImplementation(
      () =>
        new Promise((_resolve, reject) => {
          rejectRead = reject;
        }),
    );
    let refresh: Promise<void> | undefined;
    act(() => {
      refresh = result.current.refresh();
    });
    await waitFor(() => {
      expect(result.current.refreshing).toBe(true);
    });
    expect(result.current.data).toMatchObject({
      value: { title: story.title },
    });
    expect(get).toHaveBeenLastCalledWith(
      expect.stringContaining("refresh=true"),
      expect.anything(),
      expect.anything(),
    );
    await act(async () => {
      jest.mocked(get).mockRejectedValue(new Error("Offline"));
      rejectRead?.(new Error("Offline"));
      await expect(refresh).rejects.toThrow("Offline");
    });
    expect(result.current.data).toMatchObject({
      value: { title: story.title },
    });
    await waitFor(() => {
      expect(result.current.refreshError?.message).toBe("Offline");
    });
  });
  it("hides cached data immediately when refresh denies access", async () => {
    const { result } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    await loaded(result);
    jest
      .mocked(get)
      .mockRejectedValue(
        Object.assign(new Error("Not found"), { status: 404 }),
      );
    await act(async () => {
      await expect(result.current.refresh()).rejects.toThrow(
        "no longer available",
      );
    });
    await waitFor(() => {
      expect(result.current.unavailable).toBe(true);
    });
    expect(result.current.data).toBeNull();
  });
  it("does not populate full caches from an old identity's late response", async () => {
    let resolvePrevious: ((value: unknown) => void) | undefined;
    jest.mocked(get).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolvePrevious = resolve;
        }),
    );
    const { result, rerender } = renderHook(() => useNativeLink(target, true), {
      wrapper,
    });
    await waitFor(() => {
      expect(resolvePrevious).toBeDefined();
    });
    jest.mocked(useSession).mockReturnValue(session("reader-b"));
    jest
      .mocked(get)
      .mockResolvedValue(response({ ...story, title: "Current identity" }));
    rerender();
    await loaded(result);
    await act(async () => {
      resolvePrevious?.(response({ ...story, title: "Old identity" }));
    });
    expect(result.current.data).toMatchObject({
      value: { title: "Current identity" },
    });
    expect(
      client.getQueryData(storyKeys.detail("other", story.id)),
    ).toBeUndefined();
  });
  it("projects pending edits even when no full detail has been opened", async () => {
    let resolveMutation: ((value: null) => void) | undefined;
    const { result } = renderHook(
      () => ({
        preview: useNativeLink(target, true),
        mutation: useMutation({
          mutationFn: (_variables: {
            storyId: string;
            payload: { title: string };
          }) =>
            new Promise<null>((resolve) => {
              resolveMutation = resolve;
            }),
        }),
      }),
      { wrapper },
    );
    await waitFor(() => {
      expect(result.current.preview.data).not.toBeNull();
    });
    act(() => {
      result.current.mutation.mutate({
        storyId: story.id,
        payload: { title: "Inline edit" },
      });
    });
    await waitFor(() => {
      expect(result.current.preview.data).toMatchObject({
        value: { title: "Inline edit" },
      });
    });
    await act(async () => {
      resolveMutation?.(null);
    });
    await waitFor(() => {
      expect(result.current.preview.data).toMatchObject({
        value: { title: "Inline edit" },
      });
    });
    expect(
      client.getQueryData(storyKeys.detail("other", story.id)),
    ).toBeUndefined();
  });
  it("updates duplicate document previews from editor writes and canonical invalidation", async () => {
    const document = {
      id: "document",
      title: "Old document",
      contentText: "Summary",
      revision: 1,
    };
    const documentTarget: NativeLinkTarget = {
      ...target,
      kind: "document",
      id: document.id,
    };
    jest
      .mocked(get)
      .mockResolvedValue({ data: { kind: "document", value: document } });
    const { result } = renderHook(() => useNativeLink(documentTarget, true), {
      wrapper,
    });
    await loaded(result);
    const duplicate = renderHook(() => useNativeLink(documentTarget, true), {
      wrapper,
    });
    await loaded(duplicate.result);
    expect(get).toHaveBeenCalledTimes(1);
    act(() => {
      client.setQueryData(
        documentKeys.detail("other", document.id),
        {
          ...document,
          title: "Renamed document",
          contentText: "Edited document body",
          revision: 2,
        },
        { updatedAt: Date.now() + 10 },
      );
    });
    await waitFor(() => {
      for (const preview of [result, duplicate.result]) {
        expect(preview.current.data).toMatchObject({
          value: {
            title: "Renamed document",
            contentText: "Edited document body",
            revision: 2,
          },
        });
      }
    });
    jest.mocked(get).mockResolvedValue({
      data: {
        kind: "document",
        value: {
          ...document,
          title: "Collaborator's title",
          contentText: "Collaborator's body",
          revision: 3,
        },
      },
    });
    await act(async () => {
      await client.invalidateQueries({
        queryKey: documentKeys.detail("other", document.id),
      });
    });
    await waitFor(() => {
      for (const preview of [result, duplicate.result]) {
        expect(preview.current.data).toMatchObject({
          value: {
            title: "Collaborator's title",
            contentText: "Collaborator's body",
            revision: 3,
          },
        });
      }
    });
    expect(get).toHaveBeenCalledTimes(2);
  });
  it("picks up another client's document revision on the visible access check", async () => {
    jest.useFakeTimers();
    const document = {
      id: "document",
      title: "Initial title",
      contentText: "Initial body",
      revision: 1,
    };
    const documentTarget: NativeLinkTarget = {
      ...target,
      kind: "document",
      id: document.id,
    };
    jest.mocked(get).mockResolvedValue({
      data: { kind: "document", value: document },
    });
    const { result, unmount } = renderHook(
      () => useNativeLink(documentTarget, true),
      { wrapper },
    );
    try {
      await loaded(result);
      act(() => {
        client.setQueryData(
          documentKeys.detail("other", document.id),
          document,
        );
      });
      jest.mocked(get).mockResolvedValue({
        data: {
          kind: "document",
          value: {
            ...document,
            title: "Remote title",
            contentText: "Remote body",
            revision: 2,
          },
        },
      });
      await act(async () => {
        jest.advanceTimersByTime(PREVIEW_CACHE_TTL_MS);
      });
      await waitFor(() => {
        expect(result.current.data).toMatchObject({
          value: {
            title: "Remote title",
            contentText: "Remote body",
            revision: 2,
          },
        });
      });
      expect(get).toHaveBeenCalledTimes(2);
      expect(get).toHaveBeenLastCalledWith(
        "previews/native/document/document",
        expect.anything(),
        expect.anything(),
      );
    } finally {
      unmount();
      jest.useRealTimers();
    }
  });
  it("stops periodic checks offscreen and revalidates when visible again", async () => {
    jest.useFakeTimers();
    const { result, rerender, unmount } = renderHook(
      ({ active }) => useNativeLink(target, true, active),
      { initialProps: { active: true }, wrapper },
    );
    try {
      await loaded(result);
      rerender({ active: false });
      await act(async () => {
        jest.advanceTimersByTime(PREVIEW_CACHE_TTL_MS + 1);
      });
      expect(get).toHaveBeenCalledTimes(1);
      rerender({ active: true });
      await waitFor(() => {
        expect(get).toHaveBeenCalledTimes(2);
      });
    } finally {
      unmount();
      jest.useRealTimers();
    }
  });
  it("does not fetch unsupported paths or unauthenticated users", () => {
    renderHook(() => useNativeLink({ ...target, kind: "unsupported" }, true), {
      wrapper,
    });
    jest
      .mocked(useSession)
      .mockReturnValue({ data: null, status: "unauthenticated" });
    renderHook(() => useNativeLink(target, true), { wrapper });
    expect(get).not.toHaveBeenCalled();
  });
  it("treats deleted stories and wrong objective teams as unavailable", async () => {
    jest
      .mocked(get)
      .mockResolvedValue(response({ ...story, deletedAt: "2026-10-17" }));
    expect(
      await fetchNativeLink(target, { workspaceSlug: "other" }),
    ).toBeNull();
    jest.mocked(get).mockResolvedValue({
      data: { kind: "objective", value: { teamId: "wrong" } },
    });
    expect(
      await fetchNativeLink(
        { ...target, kind: "objective", id: "objective", teamId: "expected" },
        { workspaceSlug: "other" },
      ),
    ).toBeNull();
  });
});
