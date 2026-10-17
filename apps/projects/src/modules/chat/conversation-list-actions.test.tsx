/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import { readChatAlertPreferences } from "./alerts/preferences";
import { ConversationListActions } from "./conversation-list-actions";

const mockMarkAllRead = jest.fn();
let mockPending = false;
let mockUnreadThreads = 3;

jest.mock("./use-mark-all-threads-read", () => ({
  useMarkAllThreadsRead: () => ({
    mutateAsync: mockMarkAllRead,
    isPending: mockPending,
  }),
}));
jest.mock("./use-chat-unread", () => ({
  useChatUnread: () => ({ data: { unreadThreads: mockUnreadThreads } }),
}));
jest.mock("./chat-context", () => ({
  useChatIdentity: () => ({ userId: "me", conversationId: "room" }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: "acme",
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));

const openMenu = async () => {
  fireEvent.keyDown(screen.getByRole("button", { name: "Chat menu" }), {
    key: "Enter",
  });
  return screen.findByRole("menu");
};

beforeEach(() => {
  jest.clearAllMocks();
  mockPending = false;
  mockUnreadThreads = 3;
  mockMarkAllRead.mockResolvedValue(undefined);
});
afterEach(() => {
  localStorage.clear();
});

describe("chat header actions", () => {
  it("places New message before the menu and links directly to Chat notification settings", async () => {
    const onNewMessage = jest.fn();
    render(<ConversationListActions onNewMessage={onNewMessage} />);
    expect(
      screen
        .getAllByRole("button")
        .map((button) => button.getAttribute("aria-label")),
    ).toEqual(["New message", "Chat menu"]);
    fireEvent.click(screen.getByRole("button", { name: "New message" }));
    expect(onNewMessage).toHaveBeenCalledTimes(1);
    await openMenu();
    expect(
      screen.getByRole("menuitem", { name: "Notification settings" }),
    ).toHaveAttribute("href", "/acme/settings/account/notifications?tab=chat");
  });

  it("marks all threads read and closes the menu only after success", async () => {
    let finish!: () => void;
    mockMarkAllRead.mockReturnValue(
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
    );
    render(<ConversationListActions onNewMessage={jest.fn()} />);
    await openMenu();
    fireEvent.click(
      screen.getByRole("menuitem", { name: "Mark all threads as read" }),
    );
    expect(mockMarkAllRead).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("menu")).toBeInTheDocument();
    await act(async () => {
      finish();
    });
    await waitFor(() => {
      expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    });
    expect(toast.success).toHaveBeenCalledWith("All threads marked as read.");
  });

  it("keeps failures retryable", async () => {
    mockMarkAllRead.mockRejectedValue(new Error("Offline"));
    render(<ConversationListActions onNewMessage={jest.fn()} />);
    await openMenu();
    fireEvent.click(
      screen.getByRole("menuitem", { name: "Mark all threads as read" }),
    );
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled();
    });
    expect(screen.getByRole("menu")).toBeInTheDocument();
  });

  it("shows a disabled pending action while the request runs", async () => {
    mockPending = true;
    render(<ConversationListActions onNewMessage={jest.fn()} />);
    await openMenu();
    expect(
      screen.getByRole("menuitem", { name: "Marking threads as read…" }),
    ).toHaveAttribute("aria-disabled", "true");
  });

  it("disables mark-all when there are no unread threads", async () => {
    mockUnreadThreads = 0;
    render(<ConversationListActions onNewMessage={jest.fn()} />);
    await openMenu();
    expect(
      screen.getByRole("menuitem", { name: "Mark all threads as read" }),
    ).toHaveAttribute("aria-disabled", "true");
  });

  it("preserves muting the current conversation in the menu", async () => {
    render(<ConversationListActions onNewMessage={jest.fn()} />);
    await openMenu();
    fireEvent.click(
      screen.getByRole("menuitemcheckbox", { name: "Mute this conversation" }),
    );
    expect(readChatAlertPreferences("me", "acme").mutedConversations).toEqual([
      "room",
    ]);
  });
});
