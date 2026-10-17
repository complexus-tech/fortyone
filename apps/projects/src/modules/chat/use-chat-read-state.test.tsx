import type { ReactNode } from "react";
import { act, cleanup, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import type {
  ChatConversation,
  ChatMarkUnreadResult,
  ChatMessage,
  ChatUnreadSummary,
} from "./types";
import * as api from "./api";
import { chatKeys } from "./constants";
import { recordChatMessageVisible } from "./alerts/active-view";
import { reconcileChatQueries } from "./realtime";
import { useChatReadState } from "./use-chat-read-state";

let mockWorkspaceSlug = "workspace";
let mockAccountId = "me";
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspaceSlug }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockAccountId } } }),
}));
jest.mock("./api", () => ({ markRead: jest.fn(), markUnread: jest.fn() }));
jest.mock("./realtime", () => ({ reconcileChatQueries: jest.fn() }));
jest.mock("./alerts/active-view", () => ({
  recordChatMessageVisible: jest.fn(),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));

const person = {
  id: "sarah",
  fullName: "Sarah",
  avatarUrl: null,
  isSystem: false,
};
const root: ChatMessage = {
  id: "root",
  conversationId: "room",
  threadRootId: null,
  creationId: "creation",
  sequence: 5,
  author: person,
  text: "Please review",
  richContent: null,
  createdAt: "2026-10-17T10:00:00Z",
  editedAt: null,
  deletedAt: null,
  version: 1,
  replyCount: 1,
  unreadReplyCount: 0,
  lastReplyAt: null,
  replyAuthors: [person],
  reactions: [],
  attachments: [],
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
};
const reply: ChatMessage = {
  ...root,
  id: "reply",
  sequence: 6,
  threadRootId: root.id,
  replyCount: 0,
};
const room: ChatConversation = {
  id: "room",
  kind: "team",
  teamId: "team",
  name: "Product",
  description: "",
  color: null,
  members: [person],
  memberCount: 1,
  unreadCount: 0,
  unreadRootCount: 0,
  unreadReplyCount: 0,
  lastReadSequence: 10,
  lastMessageAt: null,
  lastMessageText: "",
};
const unread: ChatUnreadSummary = {
  totalUnread: 0,
  unreadThreads: 0,
  conversations: [
    {
      conversationId: room.id,
      unreadCount: 0,
      unreadRootCount: 0,
      unreadReplyCount: 0,
    },
  ],
};
const resultFor = (message: ChatMessage): ChatMarkUnreadResult => ({
  conversationId: message.conversationId,
  threadRootId: message.threadRootId,
  lastReadSequence: message.sequence - 1,
  unreadCount: 1,
  firstUnreadReplyId: message.threadRootId ? message.id : null,
  summary: {
    totalUnread: 1,
    unreadThreads: message.threadRootId ? 1 : 0,
    conversations: [
      {
        conversationId: room.id,
        unreadCount: 1,
        unreadRootCount: message.threadRootId ? 0 : 1,
        unreadReplyCount: message.threadRootId ? 1 : 0,
      },
    ],
  },
});
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
};
let client: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);
type ReadProps = Parameters<typeof useChatReadState>[0];
const initialProps: ReadProps = {
  conversationId: room.id,
  threadRootId: root.id,
};
const mount = () =>
  renderHook((props: ReadProps) => useChatReadState(props), {
    wrapper,
    initialProps,
  });
const summaryKey = (workspace = "workspace", account = "me") => [
  ...chatKeys.unread(workspace),
  account,
];
const summary = (workspace = "workspace", account = "me") =>
  client.getQueryData<ChatUnreadSummary>(summaryKey(workspace, account));
const page = (items: ChatMessage[]) => ({
  pages: [
    {
      items,
      hasOlder: false,
      hasNewer: false,
      olderCursor: null,
      newerCursor: null,
    },
  ],
  pageParams: [{}],
});
const seed = (workspace = "workspace", account = "me") => {
  client.setQueryData(summaryKey(workspace, account), unread);
  client.setQueryData(chatKeys.conversations(workspace), {
    pages: [{ items: [room], nextCursor: null }],
    pageParams: [undefined],
  });
  client.setQueryData(chatKeys.conversation(workspace, room.id), room);
  client.setQueryData(chatKeys.messages(workspace, room.id), page([root]));
  client.setQueryData(
    [...chatKeys.replies(workspace, room.id, root.id), "start"],
    page([reply]),
  );
  client.setQueryData([...chatKeys.all(workspace), "message", root.id], root);
  client.setQueryData([...chatKeys.unreadThreads(workspace), account], {
    pages: [{ items: [], nextCursor: null }],
    pageParams: [undefined],
  });
};
const advanceReads = async () => {
  await act(async () => {
    jest.advanceTimersByTime(500);
  });
};

beforeEach(() => {
  jest.useFakeTimers();
  mockWorkspaceSlug = "workspace";
  mockAccountId = "me";
  client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
  jest.spyOn(document, "hasFocus").mockReturnValue(true);
  jest.spyOn(document, "visibilityState", "get").mockReturnValue("visible");
  jest.mocked(api.markRead).mockResolvedValue(undefined);
  jest
    .mocked(api.markUnread)
    .mockImplementation(async (_ctx, _room, id) =>
      resultFor(id === reply.id ? reply : root),
    );
  jest.mocked(reconcileChatQueries).mockResolvedValue(undefined);
  seed();
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.clearAllTimers();
  jest.useRealTimers();
  jest.restoreAllMocks();
  jest.resetAllMocks();
});

describe("explicit unread and automatic visibility reads", () => {
  it.each([root, reply])(
    "updates unread badges before the network for $id and keeps the visible stream unread",
    async (message) => {
      const request = deferred<ChatMarkUnreadResult>();
      jest.mocked(api.markUnread).mockReturnValue(request.promise);
      const { result } = mount();
      let marking!: Promise<void>;
      act(() => {
        marking = result.current.markUnread(message);
      });
      expect(summary()?.totalUnread).toBe(1);
      expect(
        client.getQueryData<ChatConversation>(
          chatKeys.conversation("workspace", room.id),
        )?.unreadCount,
      ).toBe(1);
      if (message.threadRootId) {
        expect(
          client.getQueryData<ChatMessage>([
            ...chatKeys.all("workspace"),
            "message",
            root.id,
          ])?.unreadReplyCount,
        ).toBe(1);
        expect(summary()?.unreadThreads).toBe(1);
      }
      await act(async () => {
        request.resolve(resultFor(message));
        await marking;
      });
      jest.mocked(recordChatMessageVisible).mockClear();
      act(() => {
        result.current.markVisible(message);
        result.current.markVisible(message);
        result.current.recordVisibleMessages([message]);
      });
      await advanceReads();
      expect(api.markRead).not.toHaveBeenCalled();
      expect(recordChatMessageVisible).not.toHaveBeenCalled();
      expect(summary()?.totalUnread).toBe(1);
    },
  );

  it("cancels a scheduled visibility read when a message is marked unread", async () => {
    const { result } = mount();
    act(() => {
      result.current.markVisible(root);
    });
    await act(async () => {
      await result.current.markUnread(root);
    });
    await advanceReads();
    expect(api.markRead).not.toHaveBeenCalled();
    expect(api.markUnread).toHaveBeenCalledTimes(1);
  });

  it("finishes an in-flight read before sending the explicit unread rewind", async () => {
    const pendingRead = deferred<Awaited<ReturnType<typeof api.markRead>>>();
    jest.mocked(api.markRead).mockReturnValue(pendingRead.promise);
    const { result } = mount();
    act(() => {
      result.current.markVisible(root);
    });
    await advanceReads();
    expect(api.markRead).toHaveBeenCalledTimes(1);
    let marking!: Promise<void>;
    await act(async () => {
      marking = result.current.markUnread(root);
    });
    expect(summary()?.totalUnread).toBe(1);
    expect(api.markUnread).not.toHaveBeenCalled();
    await act(async () => {
      pendingRead.resolve(undefined);
      await marking;
    });
    expect(api.markUnread).toHaveBeenCalledTimes(1);
    expect(jest.mocked(api.markRead).mock.invocationCallOrder[0]).toBeLessThan(
      jest.mocked(api.markUnread).mock.invocationCallOrder[0],
    );
  });

  it.each([root, reply])(
    "resumes reads after leaving and reopening the $id stream",
    async (message) => {
      const { result, rerender } = mount();
      await act(async () => {
        await result.current.markUnread(message);
      });
      rerender(
        message.threadRootId
          ? { ...initialProps, threadRootId: null }
          : { conversationId: "other", threadRootId: null },
      );
      rerender(initialProps);
      act(() => {
        result.current.markVisible(message);
      });
      await advanceReads();
      expect(api.markRead).toHaveBeenCalledWith(
        { workspaceSlug: "workspace" },
        room.id,
        message.id,
        message.threadRootId ?? undefined,
      );
    },
  );

  it("rolls back a failed rewind and releases its visibility pause", async () => {
    const failure = deferred<ChatMarkUnreadResult>();
    jest.mocked(api.markUnread).mockReturnValue(failure.promise);
    const { result } = mount();
    let marking!: Promise<void>;
    act(() => {
      marking = result.current.markUnread(reply);
    });
    expect(summary()?.totalUnread).toBe(1);
    await act(async () => {
      failure.reject(new Error("Offline"));
      await marking;
    });
    expect(summary()?.totalUnread).toBe(0);
    expect(toast.error).toHaveBeenCalledWith(
      "Could not mark this message as unread",
      { description: "Offline" },
    );
    act(() => {
      result.current.markVisible(reply);
    });
    await advanceReads();
    expect(api.markRead).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      room.id,
      reply.id,
      root.id,
    );
  });

  it("keeps the other visible stream readable while a thread is manually unread", async () => {
    const { result } = mount();
    await act(async () => {
      await result.current.markUnread(reply);
    });
    act(() => {
      result.current.markVisible(root);
      result.current.markVisible(reply);
    });
    await advanceReads();
    expect(api.markRead).toHaveBeenCalledTimes(1);
    expect(api.markRead).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      room.id,
      root.id,
      undefined,
    );
  });

  it("releases the pause when two queued attempts both fail", async () => {
    const first = deferred<ChatMarkUnreadResult>();
    const second = deferred<ChatMarkUnreadResult>();
    jest
      .mocked(api.markUnread)
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const { result } = mount();
    let firstMark!: Promise<void>;
    let secondMark!: Promise<void>;
    await act(async () => {
      firstMark = result.current.markUnread(reply);
      secondMark = result.current.markUnread(reply);
    });
    await act(async () => {
      first.reject(new Error("Offline"));
      await firstMark;
    });
    await act(async () => {
      second.reject(new Error("Still offline"));
      await secondMark;
    });
    expect(summary()?.totalUnread).toBe(0);
    act(() => {
      result.current.markVisible(reply);
    });
    await advanceReads();
    expect(api.markRead).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      room.id,
      reply.id,
      root.id,
    );
  });

  it("keeps an earlier successful unread pause when a later attempt fails", async () => {
    const { result } = mount();
    await act(async () => {
      await result.current.markUnread(reply);
    });
    jest.mocked(api.markUnread).mockRejectedValueOnce(new Error("Offline"));
    await act(async () => {
      await result.current.markUnread(reply);
    });
    expect(summary()?.totalUnread).toBe(1);
    act(() => {
      result.current.markVisible(reply);
    });
    await advanceReads();
    expect(api.markRead).not.toHaveBeenCalled();
  });

  it("does not pause a thread opened after marking an unopened search result unread", async () => {
    const { result, rerender } = mount();
    rerender({ ...initialProps, threadRootId: null });
    await act(async () => {
      await result.current.markUnread(reply);
    });
    expect(summary()?.totalUnread).toBe(1);
    rerender(initialProps);
    act(() => {
      result.current.markVisible(reply);
    });
    await advanceReads();
    expect(api.markRead).toHaveBeenCalledWith(
      { workspaceSlug: "workspace" },
      room.id,
      reply.id,
      root.id,
    );
  });

  it.each(["account", "workspace"])(
    "ignores a late failure after changing %s and leaves the current scope usable",
    async (change) => {
      const failure = deferred<ChatMarkUnreadResult>();
      jest.mocked(api.markUnread).mockReturnValueOnce(failure.promise);
      const { result, rerender } = mount();
      let marking!: Promise<void>;
      await act(async () => {
        marking = result.current.markUnread(reply);
      });
      expect(api.markUnread).toHaveBeenCalledTimes(1);
      if (change === "account") mockAccountId = "other-account";
      else mockWorkspaceSlug = "other-workspace";
      rerender(initialProps);
      seed(mockWorkspaceSlug, mockAccountId);
      await act(async () => {
        failure.reject(new Error("Expired session"));
        await marking;
      });
      expect(toast.error).not.toHaveBeenCalled();
      expect(toast.success).not.toHaveBeenCalled();
      expect(summary(mockWorkspaceSlug, mockAccountId)?.totalUnread).toBe(0);
      act(() => {
        result.current.markVisible(reply);
      });
      await advanceReads();
      expect(api.markRead).toHaveBeenCalledWith(
        { workspaceSlug: mockWorkspaceSlug },
        room.id,
        reply.id,
        root.id,
      );
    },
  );

  it("ignores callbacks from a disabled or departed stream", async () => {
    const { result, rerender } = mount();
    rerender({ conversationId: "other", threadRootId: null });
    act(() => {
      result.current.markVisible(root);
      result.current.recordVisibleMessages([root]);
    });
    await advanceReads();
    rerender({ ...initialProps, enabled: false });
    act(() => {
      result.current.markVisible(reply);
      result.current.recordVisibleMessages([reply]);
    });
    await advanceReads();
    expect(api.markRead).not.toHaveBeenCalled();
    expect(recordChatMessageVisible).not.toHaveBeenCalled();
  });
});
