import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { ChatIdentityProvider } from "./chat-context";
import { MessageActions } from "./message-actions";
import type { ChatMessage } from "./types";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: "second",
    withWorkspace: (path: string) => `/second${path}`,
  }),
}));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));

const writeText = jest.fn();
const message: ChatMessage = {
  id: "message",
  conversationId: "conversation",
  threadRootId: null,
  creationId: "creation",
  sequence: 5,
  author: {
    id: "sarah",
    fullName: "Sarah",
    avatarUrl: null,
    isSystem: false,
  },
  text: "Please review this.\nThanks @Maya!",
  richContent: null,
  createdAt: "2026-10-17T09:00:00Z",
  editedAt: null,
  deletedAt: null,
  version: 1,
  replyCount: 0,
  unreadReplyCount: 0,
  lastReplyAt: null,
  replyAuthors: [],
  reactions: [],
  attachments: [],
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
};

const openActions = (
  value = message,
  userId = "me",
  handlers = {
    onEdit: jest.fn(),
    onDelete: jest.fn(),
    onMarkUnread: jest.fn(),
  },
) => {
  render(
    <ChatIdentityProvider
      value={{
        userId,
        conversationId: value.conversationId,
        members: [],
        mayaEnabled: false,
      }}
    >
      <MessageActions
        message={value}
        onToggleReaction={jest.fn()}
        {...handlers}
      />
    </ChatIdentityProvider>,
  );
  const trigger = screen.getByRole("button", {
    name: "More message actions",
  });
  fireEvent.keyDown(trigger, { key: "Enter" });
  return { trigger, ...handlers };
};

beforeEach(() => {
  jest.clearAllMocks();
  writeText.mockResolvedValue(undefined);
  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    value: { writeText },
  });
});

describe("MessageActions", () => {
  it("copies another person's message as text, preserving line breaks", async () => {
    const { trigger } = openActions();
    fireEvent.click(screen.getByRole("menuitem", { name: "Copy message" }));
    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith(message.text);
      expect(toast.success).toHaveBeenCalledWith("Message copied");
      expect(trigger).toHaveFocus();
    });
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("copies a link to the exact reply and its thread", async () => {
    openActions({ ...message, id: "reply", threadRootId: "root" });
    fireEvent.click(
      screen.getByRole("menuitem", { name: "Copy message link" }),
    );
    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith(
        `${window.location.origin}/second/chat?conversation=conversation&message=reply&thread=root`,
      );
      expect(toast.success).toHaveBeenCalledWith("Message link copied");
    });
  });

  it("reports a clipboard failure without claiming success", async () => {
    writeText.mockRejectedValue(new Error("Clipboard unavailable"));
    openActions();
    fireEvent.click(screen.getByRole("menuitem", { name: "Copy message" }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "Could not copy. Please try again.",
      );
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("keeps Edit and Delete restricted to the author", () => {
    const { onMarkUnread } = openActions();
    expect(
      screen.queryByRole("menuitem", { name: "Edit message" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("menuitem", { name: "Delete message" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("menuitem", { name: "Mark unread" }));
    expect(onMarkUnread).toHaveBeenCalledWith(message);
  });

  it("retains the author's Edit and Delete actions after the shared actions", () => {
    const { onEdit } = openActions(message, message.author.id);
    expect(
      screen.getAllByRole("menuitem").map((item) => item.textContent),
    ).toEqual([
      " Copy message",
      " Copy message link",
      " Edit message",
      " Delete message",
    ]);
    fireEvent.click(screen.getByRole("menuitem", { name: "Edit message" }));
    expect(onEdit).toHaveBeenCalledTimes(1);
  });

  it.each([
    { deletedAt: "2026-10-17T10:00:00Z" },
    { deliveryStatus: "sending" as const },
    { deliveryStatus: "failed" as const },
  ])("does not mark an unavailable message unread: %j", (overrides) => {
    openActions({ ...message, ...overrides });
    expect(
      screen.queryByRole("menuitem", { name: "Mark unread" }),
    ).not.toBeInTheDocument();
  });

  it("does not copy empty text for an attachment-only message", () => {
    openActions({ ...message, text: "" });
    expect(
      screen.getByRole("menuitem", { name: "Copy message" }),
    ).toHaveAttribute("aria-disabled", "true");
    expect(
      screen.getByRole("menuitem", { name: "Copy message link" }),
    ).not.toHaveAttribute("aria-disabled");
  });
});
