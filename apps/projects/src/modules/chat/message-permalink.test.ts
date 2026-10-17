import { getChatMessageLink } from "./message-permalink";

const message = {
  id: "message-id",
  conversationId: "conversation-id",
  threadRootId: null,
};

describe("getChatMessageLink", () => {
  it.each([
    ["http://localhost:3000", "/second/chat"],
    ["https://cloud.fortyone.app", "/second/chat"],
    ["https://second.fortyone.app", "/chat"],
  ])("keeps the workspace route on %s", (origin, path) => {
    expect(getChatMessageLink(message, path, origin)).toBe(
      `${origin}${path}?conversation=conversation-id&message=message-id`,
    );
  });

  it("targets both the parent thread and the specific reply", () => {
    const link = getChatMessageLink(
      { ...message, id: "reply", threadRootId: "root" },
      "/chat",
      "https://second.fortyone.app",
    );
    expect(link).toBe(
      "https://second.fortyone.app/chat?conversation=conversation-id&message=reply&thread=root",
    );
  });

  it("encodes identifiers and removes unrelated query and fragment state", () => {
    const link = new URL(
      getChatMessageLink(
        { ...message, id: "message & other", conversationId: "room?one" },
        "/second/chat?new=1&search=private&theme=dark#draft",
        "http://localhost:3000",
      ),
    );
    expect(link.searchParams.get("message")).toBe("message & other");
    expect(link.searchParams.get("conversation")).toBe("room?one");
    expect(Array.from(link.searchParams.keys())).toEqual([
      "conversation",
      "message",
    ]);
    expect(link.hash).toBe("");
  });
});
