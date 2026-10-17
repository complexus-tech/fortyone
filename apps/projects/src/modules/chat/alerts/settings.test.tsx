/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ComponentProps, ReactNode } from "react";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { chatKeys } from "../constants";
import { ChatNotificationSettings } from "./settings";
import {
  DEFAULT_CHAT_ALERT_PREFERENCES,
  chatAlertPreferencesKey,
  readChatAlertPreferences,
  saveChatAlertPreferences,
} from "./preferences";
import { enableChatSound, playChatSound } from "./sound";

const mockNative = {
  chatNotificationLinks: true,
  getPreferences: jest.fn().mockResolvedValue({ notifications: false }),
  requestNotificationPermission: jest.fn<
    Promise<"granted" | "denied" | "prompt">,
    []
  >(),
  updatePreferences: jest.fn().mockResolvedValue({ notifications: true }),
};
let mockBridge: typeof mockNative | null = mockNative;
let mockUserId: string | null = "me";
let mockWorkspaceSlug = "Acme";
const mockBrowserPermission = jest.fn();
const clients: QueryClient[] = [];
const renderSettings = () => {
  const client = new QueryClient();
  clients.push(client);
  render(
    <QueryClientProvider client={client}>
      <ChatNotificationSettings />
    </QueryClientProvider>,
  );
  return client;
};
const originalNotification = Object.getOwnPropertyDescriptor(
  window,
  "Notification",
);

jest.mock("@/shell/desktop/bridge", () => ({
  getDesktopBridge: () => mockBridge,
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspaceSlug }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: mockUserId ? { user: { id: mockUserId } } : null,
  }),
}));
jest.mock("@/components/ui/section-header", () => ({
  SectionHeader: ({
    title,
    description,
  }: {
    title: string;
    description: string;
  }) => (
    <header>
      <h2>{title}</h2>
      <p>{description}</p>
    </header>
  ),
}));
jest.mock("./sound", () => ({
  enableChatSound: jest.fn(),
  playChatSound: jest.fn(),
}));
jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Box: Container,
    Flex: Container,
    Button: ({
      children,
      onClick,
      disabled,
      "aria-label": label,
    }: ComponentProps<"button">) => (
      <button
        aria-label={label}
        disabled={disabled}
        onClick={onClick}
        type="button"
      >
        {children}
      </button>
    ),
    Input: (props: ComponentProps<"input">) => <input {...props} />,
    Text: ({ children, id, role }: ComponentProps<"span">) => (
      <span id={id} role={role}>
        {children}
      </span>
    ),
    Switch: ({
      checked,
      disabled,
      onCheckedChange,
      "aria-labelledby": label,
    }: {
      checked: boolean;
      disabled: boolean;
      onCheckedChange: (value: boolean) => void;
      "aria-labelledby": string;
    }) => (
      <input
        aria-labelledby={label}
        checked={checked}
        disabled={disabled}
        onChange={(event) => {
          onCheckedChange(event.target.checked);
        }}
        role="switch"
        type="checkbox"
      />
    ),
  };
});

const seed = (patch = {}) => {
  saveChatAlertPreferences("me", "Acme", {
    ...DEFAULT_CHAT_ALERT_PREFERENCES,
    desktop: false,
    ...patch,
  });
};

beforeEach(() => {
  mockBridge = mockNative;
  mockUserId = "me";
  mockWorkspaceSlug = "Acme";
  localStorage.clear();
  mockNative.getPreferences.mockResolvedValue({ notifications: false });
  mockNative.requestNotificationPermission.mockResolvedValue("granted");
  mockNative.updatePreferences.mockResolvedValue({ notifications: true });
  mockBrowserPermission.mockResolvedValue("granted");
  jest.mocked(enableChatSound).mockResolvedValue(true);
  jest.mocked(playChatSound).mockReturnValue(true);
  Object.defineProperty(window, "Notification", {
    configurable: true,
    value: { permission: "default", requestPermission: mockBrowserPermission },
  });
});

afterEach(() => {
  cleanup();
  clients.splice(0).forEach((client) => {
    client.clear();
  });
  jest.restoreAllMocks();
  jest.clearAllMocks();
  if (originalNotification)
    Object.defineProperty(window, "Notification", originalNotification);
  else Reflect.deleteProperty(window, "Notification");
});

describe("ChatNotificationSettings", () => {
  it("requests native permission, enables native notifications, then saves the scoped desktop preference", async () => {
    seed();
    let grant!: (permission: "granted") => void;
    mockNative.requestNotificationPermission.mockReturnValue(
      new Promise((resolve) => {
        grant = resolve;
      }),
    );
    const saved = jest.spyOn(Storage.prototype, "setItem");
    const client = renderSettings();
    fireEvent.click(
      screen.getByRole("switch", { name: "Desktop notifications" }),
    );
    expect(screen.getByRole("status")).toHaveTextContent("Saving");
    expect(mockNative.updatePreferences).not.toHaveBeenCalled();
    expect(readChatAlertPreferences("me", "Acme").desktop).toBe(false);
    await act(async () => {
      grant("granted");
    });
    await waitFor(() => {
      expect(
        screen.getByRole("switch", { name: "Desktop notifications" }),
      ).toBeChecked();
    });
    expect(mockNative.updatePreferences).toHaveBeenCalledWith({
      notifications: true,
    });
    expect(saved).toHaveBeenCalledWith(
      chatAlertPreferencesKey("me", "Acme"),
      expect.any(String),
    );
    expect(
      mockNative.requestNotificationPermission.mock.invocationCallOrder[0],
    ).toBeLessThan(mockNative.updatePreferences.mock.invocationCallOrder[0]);
    expect(
      mockNative.updatePreferences.mock.invocationCallOrder[0],
    ).toBeLessThan(saved.mock.invocationCallOrder[0]);
    expect(readChatAlertPreferences("me", "acme").desktop).toBe(true);
    expect(client.getQueryData(["desktop", "preferences"])).toEqual({
      notifications: true,
    });
    expect(
      localStorage.getItem(chatAlertPreferencesKey("other", "acme")),
    ).toBeNull();
  });

  it.each(["native", "browser"] as const)(
    "surfaces denied %s permission without enabling notifications",
    async (platform) => {
      seed();
      if (platform === "browser") mockBridge = null;
      mockNative.requestNotificationPermission.mockResolvedValue("denied");
      mockBrowserPermission.mockResolvedValue("denied");
      renderSettings();
      fireEvent.click(
        screen.getByRole("switch", { name: "Desktop notifications" }),
      );
      expect(await screen.findByRole("alert")).toHaveTextContent(
        platform === "native" ? "system settings" : "browser settings",
      );
      expect(
        screen.getByRole("switch", { name: "Desktop notifications" }),
      ).not.toBeChecked();
      expect(mockNative.updatePreferences).not.toHaveBeenCalled();
      expect(readChatAlertPreferences("me", "Acme").desktop).toBe(false);
    },
  );

  it("enables browser notifications only after granted permission", async () => {
    seed();
    mockBridge = null;
    renderSettings();
    fireEvent.click(
      screen.getByRole("switch", { name: "Desktop notifications" }),
    );
    await waitFor(() => {
      expect(
        screen.getByRole("switch", { name: "Desktop notifications" }),
      ).toBeChecked();
    });
    expect(mockBrowserPermission).toHaveBeenCalledTimes(1);
    expect(readChatAlertPreferences("me", "Acme").desktop).toBe(true);
    expect(mockNative.updatePreferences).not.toHaveBeenCalled();
  });

  it("changes sound without altering conversation mutes, other preferences, or unread caches", async () => {
    seed({
      sound: false,
      mutedConversations: ["other-room"],
      quietHours: true,
      quietStart: "21:30",
      quietEnd: "07:15",
    });
    saveChatAlertPreferences("other", "Acme", {
      ...DEFAULT_CHAT_ALERT_PREFERENCES,
      sound: false,
    });
    const unrelated = localStorage.getItem(
      chatAlertPreferencesKey("other", "Acme"),
    );
    const client = new QueryClient();
    const summary = { totalUnread: 500, unreadThreads: 3, conversations: [] };
    const conversations = {
      pages: [{ items: [{ id: "current-room", unreadCount: 100 }] }],
    };
    client.setQueryData(chatKeys.unread("Acme"), summary);
    client.setQueryData(chatKeys.conversations("Acme"), conversations);
    render(
      <QueryClientProvider client={client}>
        <ChatNotificationSettings />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByRole("switch", { name: "Message sounds" }));
    await waitFor(() => {
      expect(readChatAlertPreferences("me", "Acme").sound).toBe(true);
    });
    expect(enableChatSound).toHaveBeenCalledTimes(1);
    expect(
      screen.queryByRole("switch", { name: "Mute this conversation" }),
    ).not.toBeInTheDocument();
    expect(readChatAlertPreferences("me", "Acme")).toMatchObject({
      desktop: false,
      sound: true,
      quietHours: true,
      quietStart: "21:30",
      quietEnd: "07:15",
      mutedConversations: ["other-room"],
    });
    expect(localStorage.getItem(chatAlertPreferencesKey("other", "Acme"))).toBe(
      unrelated,
    );
    expect(client.getQueryData(chatKeys.unread("Acme"))).toEqual(summary);
    expect(client.getQueryData(chatKeys.conversations("Acme"))).toEqual(
      conversations,
    );
    expect(client.getQueryState(chatKeys.unread("Acme"))?.isInvalidated).toBe(
      false,
    );
    client.clear();
  });

  it("saves quiet hours and their times for the current workspace", async () => {
    seed();
    renderSettings();

    fireEvent.click(screen.getByRole("switch", { name: "Quiet hours" }));
    expect(await screen.findByLabelText("From")).toHaveValue("22:00");
    fireEvent.change(screen.getByLabelText("From"), {
      target: { value: "21:30" },
    });
    fireEvent.change(screen.getByLabelText("Until"), {
      target: { value: "07:15" },
    });

    expect(readChatAlertPreferences("me", "Acme")).toMatchObject({
      quietHours: true,
      quietStart: "21:30",
      quietEnd: "07:15",
    });
  });

  it("loads and updates the authenticated account and workspace after navigation", async () => {
    mockBridge = null;
    seed({ sound: false });
    saveChatAlertPreferences("other", "Design", {
      ...DEFAULT_CHAT_ALERT_PREFERENCES,
      sound: true,
    });
    const client = new QueryClient();
    clients.push(client);
    const view = render(
      <QueryClientProvider client={client}>
        <ChatNotificationSettings />
      </QueryClientProvider>,
    );
    expect(
      screen.getByRole("switch", { name: "Message sounds" }),
    ).not.toBeChecked();

    mockUserId = "other";
    mockWorkspaceSlug = "Design";
    view.rerender(
      <QueryClientProvider client={client}>
        <ChatNotificationSettings />
      </QueryClientProvider>,
    );
    expect(
      screen.getByRole("switch", { name: "Message sounds" }),
    ).toBeChecked();
    fireEvent.click(screen.getByRole("switch", { name: "Message sounds" }));

    await waitFor(() => {
      expect(readChatAlertPreferences("other", "Design").sound).toBe(false);
    });
    expect(readChatAlertPreferences("me", "Acme").sound).toBe(false);
    expect(
      localStorage.getItem(chatAlertPreferencesKey("other", "Acme")),
    ).toBeNull();
  });

  it("does not expose editable preferences without an authenticated account", () => {
    mockUserId = null;
    renderSettings();

    expect(screen.queryByRole("switch")).not.toBeInTheDocument();
    expect(localStorage.length).toBe(0);
    expect(mockNative.getPreferences).not.toHaveBeenCalled();
  });

  it("does not save an enabled sound preference when browser audio cannot be unlocked", async () => {
    seed({ sound: false });
    jest.mocked(enableChatSound).mockResolvedValue(false);
    renderSettings();
    fireEvent.click(screen.getByRole("switch", { name: "Message sounds" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Sound is unavailable",
    );
    expect(
      screen.getByRole("switch", { name: "Message sounds" }),
    ).not.toBeChecked();
    expect(readChatAlertPreferences("me", "Acme").sound).toBe(false);
  });

  it.each(["unavailable", "rejected"] as const)(
    "shows an error when sound preview is %s",
    async (failure) => {
      seed();
      if (failure === "unavailable")
        jest.mocked(enableChatSound).mockResolvedValue(false);
      else
        jest
          .mocked(enableChatSound)
          .mockRejectedValue(new Error("Audio failed"));
      renderSettings();
      fireEvent.click(screen.getByRole("button", { name: "Preview sound" }));
      expect(await screen.findByRole("alert")).toHaveTextContent(
        failure === "unavailable"
          ? "Sound is unavailable"
          : "Could not play the message sound",
      );
      expect(playChatSound).not.toHaveBeenCalled();
    },
  );
});
