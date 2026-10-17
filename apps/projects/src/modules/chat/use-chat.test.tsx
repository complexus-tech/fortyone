/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { randomUUID } from "node:crypto";
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import type {
  ChatConversation,
  ChatMessage,
  ChatMessagePage,
  ChatUnreadThread,
} from "./types";
import { useChat } from "./use-chat";
import { chatKeys } from "./constants";
import * as api from "./api";
import { reconcileChatQueries } from "./realtime";
import { recordChatMessageVisible } from "./alerts/active-view";

const params = new URLSearchParams();
const replace = jest.fn();
let mockSubdomain = false;
let mockAccountId = "me";
let mockWorkspaceSlug = "workspace";
jest.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  useSearchParams: () => params,
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: mockWorkspaceSlug,
    withWorkspace: (path: string) =>
      mockSubdomain ? path : `/workspace${path}`,
  }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: { user: { id: mockAccountId, name: "Joseph", image: null } },
  }),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("./api");
jest.mock("./realtime", () => ({
  reconcileChatQueries: jest.fn().mockResolvedValue(undefined),
}));
jest.mock("./alerts/active-view", () => ({
  recordChatMessageVisible: jest.fn(),
}));
const mocked = jest.mocked(api);
const person = {
  id: "me",
  fullName: "Joseph",
  avatarUrl: null,
  isSystem: false,
};
const room: ChatConversation = {
  id: "team",
  kind: "team",
  teamId: "team-id",
  name: "Team",
  description: "",
  color: null,
  members: [person],
  memberCount: 1,
  unreadCount: 0,
  unreadRootCount: 0,
  unreadReplyCount: 0,
  lastReadSequence: 0,
  lastMessageAt: null,
  lastMessageText: "",
};
const message: ChatMessage = {
  id: "root",
  conversationId: room.id,
  threadRootId: null,
  creationId: "creation",
  sequence: 1,
  version: 1,
  author: person,
  text: "Hello",
  richContent: null,
  createdAt: "2026-10-16T09:00:00Z",
  editedAt: null,
  deletedAt: null,
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
  attachments: [],
  links: [],
  taskRefs: [],
  reactions: [],
  replyCount: 0,
  unreadReplyCount: 0,
  replyAuthors: [],
  lastReplyAt: null,
};
const page: ChatMessagePage = {
  items: [message],
  olderCursor: null,
  newerCursor: null,
  hasOlder: false,
  hasNewer: false,
};
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};
let client: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);
beforeEach(() => {
  mockSubdomain = false;
  mockAccountId = "me";
  mockWorkspaceSlug = "workspace";
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: jest.fn().mockReturnValue({ matches: false }),
  });
  Object.defineProperty(globalThis.crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
  client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
  params.delete("conversation");
  params.delete("message");
  params.delete("new");
  params.delete("thread");
  mocked.getConversations.mockResolvedValue({
    items: [room],
    nextCursor: null,
  });
  mocked.getConversation.mockResolvedValue(room);
  mocked.getMessages.mockResolvedValue(page);
  mocked.getMessage.mockResolvedValue(message);
  mocked.editMessage.mockResolvedValue(message);
  mocked.markRead.mockResolvedValue({});
  mocked.getReplies.mockResolvedValue({ ...page, items: [] });
  mocked.createMessage.mockResolvedValue({
    ...message,
    id: "sent",
    sequence: 2,
  });
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.clearAllMocks();
});

describe("live chat state", () => {
  it("keeps team, direct-message, new-message, and deep-link navigation on clean subdomain URLs", async () => {
    mockSubdomain = true;
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });

    for (const conversationId of ["team", "direct-message"]) {
      act(() => {
        result.current.selectConversation(conversationId);
      });
      expect(replace).toHaveBeenLastCalledWith(
        `/chat?conversation=${conversationId}`,
        { scroll: false },
      );
    }
    act(() => {
      result.current.startNewMessage();
    });
    expect(replace).toHaveBeenLastCalledWith("/chat?new=1", { scroll: false });
    act(() => {
      result.current.showList();
    });
    expect(replace).toHaveBeenLastCalledWith("/chat", { scroll: false });

    await act(async () => {
      await result.current.viewMessage("root");
    });
    expect(replace).toHaveBeenLastCalledWith(
      "/chat?conversation=team&message=root",
      { scroll: false },
    );
  });

  it("opens the new-message route without loading a default conversation", async () => {
    params.set("new", "1");
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversations).toHaveLength(1);
    });
    expect(result.current.isComposing).toBe(true);
    expect(result.current.showConversation).toBe(true);
    expect(result.current.activeId).toBe("");
    expect(result.current.conversation).toBeUndefined();
    expect(mocked.getConversation).not.toHaveBeenCalled();
    expect(mocked.getMessages).not.toHaveBeenCalled();
  });

  it("carries a new draft into the selected team without sending or replacing its existing draft", async () => {
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    act(() => {
      result.current.updateDraft("team", {
        text: "Existing notes",
        content: null,
      });
      result.current.startNewMessage();
    });
    expect(result.current.isComposing).toBe(true);
    expect(result.current.activeId).toBe("");
    expect(replace).toHaveBeenLastCalledWith("/workspace/chat?new=1", {
      scroll: false,
    });
    act(() => {
      result.current.updateNewMessageDraft({
        text: "New thought",
        content: null,
      });
    });
    await act(async () => {
      await result.current.sendMessage("new-message");
    });
    expect(mocked.createMessage).not.toHaveBeenCalled();
    act(() => {
      result.current.chooseNewMessageRecipient("team");
    });
    expect(result.current.isComposing).toBe(false);
    expect(result.current.draft.text).toBe("Existing notes\n\nNew thought");
    expect(result.current.newMessageDraft.text).toBe("");
    expect(mocked.createMessage).not.toHaveBeenCalled();
  });

  it("keeps the new-message draft when canceled and ignores a late recipient selection", async () => {
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    act(() => {
      result.current.startNewMessage();
    });
    act(() => {
      result.current.updateNewMessageDraft({
        text: "Keep this thought",
        content: null,
      });
    });
    act(() => {
      result.current.showList();
    });
    act(() => {
      result.current.chooseNewMessageRecipient("late-dm");
    });
    expect(result.current.isComposing).toBe(false);
    expect(result.current.showConversation).toBe(false);
    expect(result.current.activeId).toBe("team");
    expect(result.current.newMessageDraft.text).toBe("Keep this thought");
    expect(replace).toHaveBeenLastCalledWith("/workspace/chat", {
      scroll: false,
    });
    act(() => {
      result.current.startNewMessage();
    });
    expect(result.current.newMessageDraft.text).toBe("Keep this thought");
  });

  it.each(["succeeds", "fails"] as const)(
    "preserves a transferred new draft when an earlier send to that conversation %s",
    async (outcome) => {
      const send = deferred<ChatMessage>();
      mocked.createMessage.mockReturnValueOnce(send.promise);
      const { result } = renderHook(useChat, { wrapper });
      await waitFor(() => {
        expect(result.current.conversation?.id).toBe("team");
      });
      act(() => {
        result.current.updateDraft("team", {
          text: "Earlier message",
          content: null,
        });
      });
      let sending = Promise.resolve();
      act(() => {
        sending = result.current.sendMessage("team");
      });
      expect(result.current.draft.text).toBe("");
      await waitFor(() => {
        expect(
          result.current.conversation?.messages.at(-1)?.deliveryStatus,
        ).toBe("sending");
      });
      act(() => {
        result.current.startNewMessage();
      });
      act(() => {
        result.current.updateNewMessageDraft({
          text: "New unsent thought",
          content: null,
        });
      });
      act(() => {
        result.current.chooseNewMessageRecipient("team");
      });
      expect(result.current.draft.text).toBe("New unsent thought");

      await act(async () => {
        if (outcome === "succeeds") {
          send.resolve({
            ...message,
            id: "sent",
            text: "Earlier message",
            sequence: 2,
          });
        } else {
          send.reject(new Error("Disconnected"));
        }
        await sending;
      });

      expect(mocked.createMessage).toHaveBeenCalledTimes(1);
      expect(mocked.createMessage.mock.calls[0][2].text).toBe(
        "Earlier message",
      );
      expect(result.current.draft.text).toBe("New unsent thought");
      if (outcome === "fails") {
        await waitFor(() => {
          expect(
            result.current.conversation?.messages.at(-1)?.deliveryStatus,
          ).toBe("failed");
        });
      }
      expect(result.current.newMessageDraft.text).toBe("");
    },
  );

  it("does not replace the new-message pane when an earlier message lookup finishes", async () => {
    const lookup = deferred<ChatMessage>();
    mocked.getMessage.mockReturnValueOnce(lookup.promise);
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    let opening = Promise.resolve();
    act(() => {
      opening = result.current.viewMessage("old-result");
    });
    act(() => {
      result.current.startNewMessage();
    });
    act(() => {
      result.current.updateNewMessageDraft({
        text: "Keep writing",
        content: null,
      });
    });

    await act(async () => {
      lookup.resolve({ ...message, id: "old-result" });
      await opening;
    });

    expect(result.current.isComposing).toBe(true);
    expect(result.current.activeId).toBe("");
    expect(result.current.threadId).toBeNull();
    expect(result.current.newMessageDraft.text).toBe("Keep writing");
    expect(replace).toHaveBeenLastCalledWith("/workspace/chat?new=1", {
      scroll: false,
    });
  });

  it("handles empty/loading query data and starts fetching only after a conversation is selected", async () => {
    const { result } = renderHook(useChat, { wrapper });
    expect(result.current.conversation).toBeUndefined();
    expect(result.current.replies).toEqual([]);
    await waitFor(() => {
      expect(result.current.conversation?.messages).toHaveLength(1);
    });
    expect(
      mocked.getMessages.mock.calls.every((call) => call[1] === "team"),
    ).toBe(true);
  });
  it("keeps failed messages and reuses creationId across retries, preventing duplicate sends", async () => {
    mocked.createMessage.mockRejectedValueOnce(
      new Error("Network disconnected"),
    );
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    act(() => {
      result.current.updateDraft("team", {
        text: "Retry safely",
        content: null,
      });
    });
    await act(async () => {
      await result.current.sendMessage("team");
    });
    expect(result.current.draft.text).toBe("");
    await waitFor(() => {
      expect(result.current.conversation?.messages.at(-1)?.deliveryStatus).toBe(
        "failed",
      );
    });
    const failedId = result.current.conversation!.messages.at(-1)!.id;
    const firstId = mocked.createMessage.mock.calls[0][2].creationId;
    await act(async () => {
      await Promise.all([
        result.current.retryMessage(failedId),
        result.current.retryMessage(failedId),
      ]);
    });
    expect(mocked.createMessage).toHaveBeenCalledTimes(2);
    expect(mocked.createMessage.mock.calls[1][2].creationId).toBe(firstId);
    expect(result.current.draft.text).toBe("");
  });
  it("inserts root messages immediately and allows another send before either request finishes", async () => {
    const first = deferred<ChatMessage>();
    const second = deferred<ChatMessage>();
    mocked.createMessage
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.messages).toHaveLength(1);
    });
    let firstSend = Promise.resolve();
    let secondSend = Promise.resolve();
    act(() => {
      result.current.updateDraft("team", { text: "First", content: null });
      firstSend = result.current.sendMessage("team");
      void result.current.sendMessage("team");
    });
    expect(result.current.draft.text).toBe("");
    await waitFor(() => {
      expect(result.current.conversation?.messages.at(-1)?.text).toBe("First");
    });
    expect(mocked.createMessage).toHaveBeenCalledTimes(1);
    act(() => {
      result.current.updateDraft("team", { text: "Second", content: null });
      secondSend = result.current.sendMessage("team");
    });
    await waitFor(() => {
      expect(
        result.current.conversation?.messages
          .slice(-2)
          .map((item) => item.text),
      ).toEqual(["First", "Second"]);
    });
    expect(mocked.createMessage).toHaveBeenCalledTimes(2);
    const creationIds = mocked.createMessage.mock.calls.map(
      (call) => call[2].creationId,
    );
    await act(async () => {
      second.resolve({
        ...message,
        id: "second",
        text: "Second",
        creationId: creationIds[1],
        sequence: 3,
      });
      first.resolve({
        ...message,
        id: "first",
        text: "First",
        creationId: creationIds[0],
        sequence: 2,
      });
      await Promise.all([firstSend, secondSend]);
    });
    await waitFor(() => {
      expect(
        result.current.conversation?.messages.map((item) => item.id),
      ).toEqual(["root", "first", "second"]);
    });
    expect(result.current.draft.text).toBe("");
  });

  it("inserts replies with attachments immediately, excludes them from read receipts, and dedupes a realtime response before HTTP finishes", async () => {
    const send = deferred<ChatMessage>();
    mocked.createMessage.mockReturnValueOnce(send.promise);
    const focused = jest.spyOn(document, "hasFocus").mockReturnValue(true);
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.messages).toHaveLength(1);
    });
    act(() => {
      result.current.openThread(message, document.createElement("button"));
    });
    await waitFor(() => {
      expect(result.current.repliesQuery.isSuccess).toBe(true);
    });
    const attachment = {
      id: "file",
      fileName: "brief.pdf",
      mediaType: "application/pdf",
      sizeBytes: 100,
      url: "",
    };
    let sending = Promise.resolve();
    act(() => {
      result.current.updateDraft(result.current.threadKey, {
        text: "",
        content: null,
        attachments: [attachment],
      });
      sending = result.current.sendMessage(result.current.threadKey, "root");
    });
    expect(result.current.threadDraft.text).toBe("");
    expect(result.current.threadDraft.attachments).toBeUndefined();
    await waitFor(() => {
      expect(result.current.replies).toHaveLength(1);
    });
    const pending = result.current.replies[0];
    expect(pending).toMatchObject({
      threadRootId: "root",
      attachments: [attachment],
      deliveryStatus: "sending",
    });
    act(() => {
      result.current.markVisible(pending);
      result.current.recordVisibleMessages([pending]);
    });
    expect(mocked.markRead).not.toHaveBeenCalled();
    expect(recordChatMessageVisible).not.toHaveBeenCalled();
    const confirmed = {
      ...message,
      id: "reply",
      threadRootId: "root",
      text: "",
      attachments: [attachment],
      creationId: pending.creationId,
      sequence: 2,
    };
    act(() => {
      client.setQueryData(
        [...chatKeys.replies("workspace", "team", "root"), "start"],
        { pages: [{ ...page, items: [confirmed] }], pageParams: [{}] },
      );
    });
    await waitFor(() => {
      expect(result.current.replies).toEqual([confirmed]);
    });
    await act(async () => {
      send.resolve(confirmed);
      await sending;
    });
    expect(result.current.replies).toEqual([confirmed]);
    focused.mockRestore();
  });

  it("does not show an outgoing message in another account or workspace", async () => {
    const send = deferred<ChatMessage>();
    mocked.createMessage.mockReturnValueOnce(send.promise);
    const { result, rerender } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    let sending = Promise.resolve();
    act(() => {
      result.current.updateDraft("team", {
        text: "Scoped message",
        content: null,
      });
      sending = result.current.sendMessage("team");
    });
    await waitFor(() => {
      expect(result.current.conversation?.messages.at(-1)?.text).toBe(
        "Scoped message",
      );
    });
    mockAccountId = "another-account";
    rerender();
    expect(
      result.current.conversation?.messages.some(
        (item) => item.text === "Scoped message",
      ),
    ).toBe(false);
    mockAccountId = "me";
    mockWorkspaceSlug = "other-workspace";
    rerender();
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    expect(
      result.current.conversation?.messages.some(
        (item) => item.text === "Scoped message",
      ),
    ).toBe(false);
    mockWorkspaceSlug = "workspace";
    rerender();
    await waitFor(() => {
      expect(result.current.conversation?.messages.at(-1)?.text).toBe(
        "Scoped message",
      );
    });
    await act(async () => {
      send.reject(new Error("Offline"));
      await sending;
    });
  });

  it("treats a realtime-confirmed message as sent even if the HTTP response is lost", async () => {
    const send = deferred<ChatMessage>();
    mocked.createMessage.mockReturnValueOnce(send.promise);
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.messages).toHaveLength(1);
    });
    let sending = Promise.resolve();
    act(() => {
      result.current.updateDraft("team", {
        text: "Confirmed elsewhere",
        content: null,
      });
      sending = result.current.sendMessage("team");
    });
    const creationId = mocked.createMessage.mock.calls[0][2].creationId;
    const confirmed = {
      ...message,
      id: "confirmed",
      text: "Confirmed elsewhere",
      creationId,
      sequence: 2,
    };
    act(() => {
      client.setQueryData(chatKeys.messages("workspace", "team"), {
        pages: [{ ...page, items: [message, confirmed] }],
        pageParams: [{}],
      });
    });
    await waitFor(() => {
      expect(result.current.conversation?.messages).toEqual([
        message,
        confirmed,
      ]);
    });
    await act(async () => {
      send.reject(new Error("Response lost"));
      await sending;
    });
    expect(result.current.conversation?.messages).toEqual([message, confirmed]);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("retains failed sends when Chat unmounts without replacing a newer draft", async () => {
    mocked.createMessage.mockRejectedValueOnce(new Error("Offline"));
    const mounted = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(mounted.result.current.conversation?.id).toBe("team");
    });
    await act(async () => {
      mounted.result.current.updateDraft("team", {
        text: "Keep this",
        content: null,
      });
      await mounted.result.current.sendMessage("team");
    });
    await waitFor(() => {
      expect(
        mounted.result.current.conversation?.messages.at(-1)?.deliveryStatus,
      ).toBe("failed");
    });
    mounted.unmount();
    const remounted = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(
        remounted.result.current.conversation?.messages.at(-1),
      ).toMatchObject({ text: "Keep this", deliveryStatus: "failed" });
    });
  });

  it.each(["account", "workspace"])(
    "keeps late sent messages out of shared caches after a %s switch",
    async (change) => {
      const send = deferred<ChatMessage>();
      mocked.createMessage.mockReturnValueOnce(send.promise);
      const mounted = renderHook(useChat, { wrapper });
      await waitFor(() => {
        expect(mounted.result.current.conversation?.id).toBe("team");
      });
      let sending = Promise.resolve();
      act(() => {
        mounted.result.current.updateDraft("team", {
          text: "Original account",
          content: null,
        });
        sending = mounted.result.current.sendMessage("team");
      });
      const creationId = mocked.createMessage.mock.calls[0][2].creationId;
      if (change === "account") mockAccountId = "other-account";
      else mockWorkspaceSlug = "other-workspace";
      mounted.rerender();
      await act(async () => {
        send.resolve({ ...message, id: "late-sent", creationId, sequence: 2 });
        await sending;
      });
      expect(
        client.getQueryData([
          ...chatKeys.all("workspace"),
          "message",
          "late-sent",
        ]),
      ).toBeUndefined();
      expect(reconcileChatQueries).not.toHaveBeenCalled();
      expect(toast.error).not.toHaveBeenCalled();
      expect(client.getQueryData(["chat-outbox", "workspace", "me"])).toEqual([
        expect.objectContaining({ id: "late-sent", creationId }),
      ]);
    },
  );

  it("retains a send that fails after unmount without notifying the next page", async () => {
    const send = deferred<ChatMessage>();
    mocked.createMessage.mockReturnValueOnce(send.promise);
    const mounted = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(mounted.result.current.conversation?.id).toBe("team");
    });
    let sending = Promise.resolve();
    act(() => {
      mounted.result.current.updateDraft("team", {
        text: "Keep after navigation",
        content: null,
      });
      sending = mounted.result.current.sendMessage("team");
    });
    mounted.unmount();
    await act(async () => {
      send.reject(new Error("Offline"));
      await sending;
    });
    expect(toast.error).not.toHaveBeenCalled();
    const remounted = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(
        remounted.result.current.conversation?.messages.at(-1),
      ).toMatchObject({
        text: "Keep after navigation",
        deliveryStatus: "failed",
      });
    });
  });

  it("resolves deep-linked replies to their root and opens the correct thread", async () => {
    params.set("conversation", "team");
    params.set("message", "reply");
    mocked.getMessage.mockImplementation(async (_ctx, id) =>
      id === "reply"
        ? { ...message, id: "reply", threadRootId: "root", sequence: 2 }
        : message,
    );
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.threadId).toBe("root");
    });
    expect(result.current.targetMessageId).toBe("root");
    expect(result.current.threadTargetId).toBe("reply");
    await waitFor(() => {
      expect(mocked.getReplies).toHaveBeenCalledWith(
        { workspaceSlug: "workspace" },
        "root",
        { around: "reply" },
        expect.any(AbortSignal),
      );
    });
  });
  it("allows attachment-only messages and preserves the version captured when editing began", async () => {
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    act(() => {
      result.current.updateDraft("team", {
        text: "",
        content: null,
        attachments: [
          {
            id: "file",
            fileName: "brief.pdf",
            mediaType: "application/pdf",
            sizeBytes: 100,
            url: "",
          },
        ],
      });
    });
    await act(async () => {
      await result.current.sendMessage("team");
    });
    expect(mocked.createMessage.mock.calls[0][2]).toMatchObject({
      text: "",
      attachmentIds: ["file"],
    });
    await act(async () => {
      await result.current.editMessage("root", {
        text: "Changed",
        content: null,
        baseVersion: 7,
      });
    });
    expect(mocked.editMessage.mock.calls[0][2].version).toBe(7);
  });

  it("does not mark background messages read and advances only through a focused visible message", async () => {
    const focused = jest.spyOn(document, "hasFocus").mockReturnValue(false);
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    act(() => {
      result.current.markVisible(message);
    });
    expect(mocked.markRead).not.toHaveBeenCalled();
    expect(recordChatMessageVisible).not.toHaveBeenCalled();
    focused.mockReturnValue(true);
    act(() => {
      result.current.markVisible(message);
    });
    await waitFor(() => {
      expect(mocked.markRead).toHaveBeenCalledWith(
        { workspaceSlug: "workspace" },
        "team",
        "root",
        undefined,
      );
    });
    focused.mockRestore();
    expect(recordChatMessageVisible).toHaveBeenCalledWith(
      "me",
      "workspace",
      message,
    );
    expect(reconcileChatQueries).toHaveBeenCalledWith(
      client,
      "workspace",
      "team",
    );
  });

  it("opens unread replies in an old thread outside the loaded conversation page", async () => {
    const { result, rerender } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    const olderRoot: ChatUnreadThread = {
      ...message,
      id: "old-root",
      conversationId: "old-team",
      unreadReplyCount: 100,
      conversationName: "Old team",
      conversationKind: "team",
      firstUnreadReplyId: "first-unread",
    };
    mocked.getMessage.mockResolvedValue({
      ...olderRoot,
      id: "first-unread",
      threadRootId: "old-root",
      sequence: 800,
    });
    const trigger = document.createElement("button");
    act(() => {
      result.current.openUnreadThread(olderRoot, trigger);
    });
    expect(replace).toHaveBeenLastCalledWith(
      "/workspace/chat?conversation=old-team&message=first-unread&thread=old-root",
      { scroll: false },
    );
    params.set("conversation", "old-team");
    params.set("message", "first-unread");
    params.set("thread", "old-root");
    rerender();
    await waitFor(() => {
      expect(result.current.thread?.id).toBe("old-root");
    });
    expect(result.current.activeId).toBe("old-team");
    expect(result.current.threadTargetId).toBe("first-unread");
    expect(mocked.getReplies).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      "old-root",
      { around: "first-unread" },
      expect.any(AbortSignal),
    );
    expect(mocked.markRead).not.toHaveBeenCalled();
  });

  it("records all visible incoming messages without changing read cursors or acknowledging background content", async () => {
    const focused = jest.spyOn(document, "hasFocus").mockReturnValue(false);
    const { result } = renderHook(useChat, { wrapper });
    await waitFor(() => {
      expect(result.current.conversation?.id).toBe("team");
    });
    const incoming = [message, { ...message, id: "second", sequence: 2 }];
    act(() => {
      result.current.recordVisibleMessages(incoming);
    });
    expect(recordChatMessageVisible).not.toHaveBeenCalled();
    focused.mockReturnValue(true);
    act(() => {
      result.current.recordVisibleMessages(incoming);
    });
    expect(recordChatMessageVisible).toHaveBeenCalledTimes(2);
    expect(recordChatMessageVisible).toHaveBeenLastCalledWith(
      "me",
      "workspace",
      incoming[1],
    );
    expect(mocked.markRead).not.toHaveBeenCalled();
    focused.mockRestore();
  });
});
