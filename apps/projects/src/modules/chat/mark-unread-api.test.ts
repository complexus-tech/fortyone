import { post } from "@/lib/http";
import { markUnread } from "./api";
import type { ChatMarkUnreadResult } from "./types";

jest.mock("@/lib/http", () => ({ post: jest.fn() }));

const result: ChatMarkUnreadResult = {
  conversationId: "room",
  threadRootId: null,
  lastReadSequence: 4,
  unreadCount: 1,
  firstUnreadReplyId: null,
  summary: {
    totalUnread: 1,
    unreadThreads: 0,
    conversations: [
      {
        conversationId: "room",
        unreadCount: 1,
        unreadRootCount: 1,
        unreadReplyCount: 0,
      },
    ],
  },
};

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(post).mockResolvedValue(result);
});

it.each([undefined, "root"])(
  "requests one explicit unread rewind without transport retries (thread: %s)",
  async (threadRootId) => {
    const scope = { workspaceSlug: "workspace" };
    await markUnread(scope, "room", "message", threadRootId);
    expect(post).toHaveBeenCalledWith(
      "chat/conversations/room/unread",
      { fromMessageId: "message", threadRootId },
      scope,
      { retry: 0 },
      expect.any(Function),
    );
  },
);

it("decodes authoritative cursors and counts and rejects malformed responses", async () => {
  await markUnread({ workspaceSlug: "workspace" }, "room", "message");
  const decode = jest.mocked(post).mock.calls[0][4]!;
  expect(decode({ data: result })).toEqual(result);
  expect(() => decode({ data: { ...result, lastReadSequence: -1 } })).toThrow();
  expect(() =>
    decode({
      data: { ...result, summary: { ...result.summary, totalUnread: -1 } },
    }),
  ).toThrow();
  expect(() => decode({ error: { message: "Not a member" } })).toThrow(
    "Not a member",
  );
});

it("surfaces a failed write without submitting a second rewind", async () => {
  jest.mocked(post).mockRejectedValue(new Error("Offline"));
  await expect(
    markUnread({ workspaceSlug: "workspace" }, "room", "message"),
  ).rejects.toThrow("Offline");
  expect(post).toHaveBeenCalledTimes(1);
});
