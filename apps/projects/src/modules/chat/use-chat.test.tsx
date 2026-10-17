/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { randomUUID } from "node:crypto";
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ChatConversation, ChatMessage, ChatMessagePage } from "./types";
import { useChat } from "./use-chat";
import * as api from "./api";

const params = new URLSearchParams();
const replace = jest.fn();
jest.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  useSearchParams: () => params,
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "me" } } }),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("./api");
jest.mock("./realtime", () => ({
  reconcileChatQueries: jest.fn().mockResolvedValue(undefined),
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
      expect(result.current.sendingKeys.has("team")).toBe(true);
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
      expect(result.current.draft.text).toBe(
        "Earlier message\n\nNew unsent thought",
      );

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
      expect(result.current.sendingKeys.has("team")).toBe(false);
      expect(result.current.draft.text).toBe(
        outcome === "succeeds"
          ? "New unsent thought"
          : "Earlier message\n\nNew unsent thought",
      );
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
  it("keeps failed drafts and reuses creationId across retries, preventing duplicate sends", async () => {
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
    expect(result.current.draft.text).toBe("Retry safely");
    const firstId = mocked.createMessage.mock.calls[0][2].creationId;
    await act(async () => {
      await Promise.all([
        result.current.sendMessage("team"),
        result.current.sendMessage("team"),
      ]);
    });
    expect(mocked.createMessage).toHaveBeenCalledTimes(2);
    expect(mocked.createMessage.mock.calls[1][2].creationId).toBe(firstId);
    expect(result.current.draft.text).toBe("");
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
  });
});
