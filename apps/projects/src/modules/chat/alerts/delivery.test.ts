import { update } from "idb-keyval";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { bindDesktopSession } from "@/shell/desktop/session-binding";
import { wasChatMessageVisible } from "./active-view";
import type { ChatNotificationEvent } from "./contract";
import {
  createChatAlertDelivery,
  isLiveChatNotification,
  chatNotificationPath,
} from "./delivery";
import {
  DEFAULT_CHAT_ALERT_PREFERENCES,
  readChatAlertPreferences,
  isChatQuietTime,
} from "./preferences";
import { canPlayChatSound, playChatSound } from "./sound";

jest.mock("idb-keyval", () => ({ update: jest.fn() }));
jest.mock("@/shell/desktop/bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("@/shell/desktop/session-binding", () => ({
  bindDesktopSession: jest.fn(),
}));
jest.mock("./active-view", () => ({ wasChatMessageVisible: jest.fn() }));
jest.mock("./preferences", () => ({
  ...jest.requireActual("./preferences"),
  readChatAlertPreferences: jest.fn(),
  isChatQuietTime: jest.fn(),
}));
jest.mock("./sound", () => ({
  canPlayChatSound: jest.fn(),
  playChatSound: jest.fn(),
}));

const scope = {
  accountId: "reader",
  workspaceId: "workspace",
  workspaceSlug: "engineering",
};
const now = new Date("2026-10-17T12:00:00Z");
const message = (
  id = "one",
  change: Partial<ChatNotificationEvent> = {},
): ChatNotificationEvent => ({
  kind: "chat-notification",
  notificationId: `receipt-${id}`,
  eventId: `event-${id}`,
  workspaceId: scope.workspaceId,
  userId: scope.accountId,
  actorId: "sender",
  conversationId: "conversation",
  messageId: id,
  threadRootId: null,
  createdAt: now.toISOString(),
  occurredAt: now.toISOString(),
  title: "New chat message",
  body: `Message ${id}`,
  ...change,
});
const browserNotification = jest.fn();
const originalNotification = Object.getOwnPropertyDescriptor(
  globalThis,
  "Notification",
);
const disposals: (() => void)[] = [];
let hasFocus: jest.SpyInstance<boolean, []>;
const makeDelivery = (current = () => true, onError = jest.fn()) => {
  const delivery = createChatAlertDelivery(scope, current, onError);
  disposals.push(() => {
    delivery.dispose();
  });
  return delivery;
};
const native = (changes: Record<string, unknown> = {}) => {
  const bridge = {
    chatNotificationLinks: true,
    deliverNotification: jest.fn().mockResolvedValue(undefined),
    getPreferences: jest.fn().mockResolvedValue({ notifications: true }),
    ...changes,
  };
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue(
      bridge as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>,
    );
  return bridge;
};

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(now);
  jest.clearAllMocks();
  const records = new Map<IDBValidKey, unknown>();
  jest.mocked(update).mockImplementation(async (key, updater) => {
    // The updater executes synchronously, matching the serialized IndexedDB
    // read/write transaction shared by distinct live tab controllers.
    records.set(key, updater(records.get(key)));
  });
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  jest.mocked(bindDesktopSession).mockResolvedValue(true);
  jest.mocked(wasChatMessageVisible).mockReturnValue(false);
  jest
    .mocked(readChatAlertPreferences)
    .mockReturnValue({ ...DEFAULT_CHAT_ALERT_PREFERENCES });
  jest.mocked(isChatQuietTime).mockReturnValue(false);
  jest.mocked(canPlayChatSound).mockReturnValue(true);
  jest.mocked(playChatSound).mockReturnValue(true);
  hasFocus = jest.spyOn(document, "hasFocus").mockReturnValue(true);
  let identity = 0;
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: () => `claim-${identity++}`,
  });
  browserNotification.mockImplementation(() => ({
    close: jest.fn(),
    onclick: null,
  }));
  Object.defineProperty(globalThis, "Notification", {
    configurable: true,
    value: Object.assign(browserNotification, { permission: "granted" }),
  });
});
afterEach(() => {
  disposals.splice(0).forEach((dispose) => {
    dispose();
  });
  jest.restoreAllMocks();
  jest.useRealTimers();
});
afterAll(() => {
  if (originalNotification)
    Object.defineProperty(globalThis, "Notification", originalNotification);
  else Reflect.deleteProperty(globalThis, "Notification");
});

it.each([
  { actorId: scope.accountId },
  { userId: "another" },
  { workspaceId: "another" },
  { occurredAt: "2026-10-17T11:58:00Z" },
  { occurredAt: "invalid" },
  { occurredAt: "2026-10-17T12:00:06Z" },
])("rejects own, stale and out-of-scope messages: %o", async (change) => {
  const event = message("one", change);
  expect(isLiveChatNotification(event, scope)).toBe(false);
  makeDelivery().receive(event);
  await jest.advanceTimersByTimeAsync(1_000);
  expect(browserNotification).not.toHaveBeenCalled();
  expect(playChatSound).not.toHaveBeenCalled();
});

it("groups a burst once, targets the latest unseen message, and counts only unseen messages", async () => {
  const bridge = native();
  jest
    .mocked(wasChatMessageVisible)
    .mockImplementation((_account, _workspace, id) => id === "seen");
  const delivery = makeDelivery();
  delivery.receive(message("first"));
  delivery.receive(
    message("second", {
      occurredAt: new Date(now.getTime() + 100).toISOString(),
      threadRootId: "thread",
    }),
  );
  delivery.receive(
    message("seen", {
      occurredAt: new Date(now.getTime() + 200).toISOString(),
    }),
  );
  delivery.receive(
    message("second", {
      occurredAt: new Date(now.getTime() + 100).toISOString(),
      threadRootId: "thread",
    }),
  );
  await jest.advanceTimersByTimeAsync(800);
  expect(bridge.deliverNotification).toHaveBeenCalledTimes(1);
  expect(bridge.deliverNotification).toHaveBeenCalledWith(
    expect.objectContaining({
      notificationId: "receipt-second",
      entityId: "second",
      threadRootId: "thread",
      body: "2 new messages in this conversation.",
      silent: false,
    }),
  );
  expect(playChatSound).not.toHaveBeenCalled();
});

it("shares banner and audio receipts across two tabs and replay", async () => {
  const one = makeDelivery();
  const two = makeDelivery();
  one.receive(message());
  two.receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(browserNotification).toHaveBeenCalledTimes(1);
  expect(browserNotification).toHaveBeenCalledWith(
    "New chat message",
    expect.objectContaining({ silent: true }),
  );
  expect(playChatSound).toHaveBeenCalledTimes(1);
  await jest.advanceTimersByTimeAsync(4_000);
  two.receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(browserNotification).toHaveBeenCalledTimes(1);
  expect(playChatSound).toHaveBeenCalledTimes(1);
});

it("a tab without unlocked audio cannot consume another tab's sound", async () => {
  jest
    .mocked(canPlayChatSound)
    .mockReturnValueOnce(false)
    .mockReturnValue(true);
  makeDelivery().receive(message());
  makeDelivery().receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(browserNotification).toHaveBeenCalledTimes(1);
  expect(playChatSound).toHaveBeenCalledTimes(1);
});

it("lets a focused tab claim audio before an unlocked background tab", async () => {
  hasFocus.mockReturnValueOnce(false).mockReturnValue(true);
  makeDelivery().receive(message());
  makeDelivery().receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(playChatSound).toHaveBeenCalledTimes(1);
  await jest.advanceTimersByTimeAsync(150);
  expect(playChatSound).toHaveBeenCalledTimes(1);
});

it.each(["muted", "quiet"])(
  "%s conversations do not consume the sound cooldown",
  async (reason) => {
    const delivery = makeDelivery();
    if (reason === "muted")
      jest.mocked(readChatAlertPreferences).mockReturnValue({
        ...DEFAULT_CHAT_ALERT_PREFERENCES,
        mutedConversations: ["conversation"],
      });
    else jest.mocked(isChatQuietTime).mockReturnValue(true);
    delivery.receive(message());
    await jest.advanceTimersByTimeAsync(800);
    expect(browserNotification).not.toHaveBeenCalled();
    expect(playChatSound).not.toHaveBeenCalled();
    jest
      .mocked(readChatAlertPreferences)
      .mockReturnValue({ ...DEFAULT_CHAT_ALERT_PREFERENCES });
    jest.mocked(isChatQuietTime).mockReturnValue(false);
    delivery.receive(message("other", { conversationId: "unmuted" }));
    await jest.advanceTimersByTimeAsync(800);
    expect(playChatSound).toHaveBeenCalledTimes(1);
  },
);

it("suppresses banners for messages already visible in any tab while retaining one incoming-message sound", async () => {
  jest.mocked(wasChatMessageVisible).mockReturnValue(true);
  makeDelivery().receive(message());
  makeDelivery().receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(browserNotification).not.toHaveBeenCalled();
  expect(playChatSound).toHaveBeenCalledTimes(1);
});

it("honors silent native notifications and keeps desktop-off incoming sounds", async () => {
  const bridge = native();
  jest
    .mocked(readChatAlertPreferences)
    .mockReturnValue({ ...DEFAULT_CHAT_ALERT_PREFERENCES, sound: false });
  const delivery = makeDelivery();
  delivery.receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(bridge.deliverNotification).toHaveBeenCalledWith(
    expect.objectContaining({ silent: true }),
  );
  expect(playChatSound).not.toHaveBeenCalled();
  jest
    .mocked(readChatAlertPreferences)
    .mockReturnValue({ ...DEFAULT_CHAT_ALERT_PREFERENCES, desktop: false });
  delivery.receive(message("two"));
  await jest.advanceTimersByTimeAsync(800);
  expect(bridge.deliverNotification).toHaveBeenCalledTimes(1);
  expect(playChatSound).toHaveBeenCalledTimes(1);
});

it("uses in-app audio when native system notifications are disabled", async () => {
  const bridge = native({
    getPreferences: jest.fn().mockResolvedValue({ notifications: false }),
  });
  makeDelivery().receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(bridge.deliverNotification).not.toHaveBeenCalled();
  expect(playChatSound).toHaveBeenCalledTimes(1);
});

it("releases failed native claims and retries with the same durable receipt", async () => {
  const bridge = native();
  bridge.deliverNotification.mockRejectedValueOnce(new Error("OS unavailable"));
  const onError = jest.fn();
  makeDelivery(() => true, onError).receive(message());
  await jest.advanceTimersByTimeAsync(800);
  expect(onError).toHaveBeenCalledTimes(1);
  await jest.advanceTimersByTimeAsync(2_000);
  expect(bridge.deliverNotification).toHaveBeenCalledTimes(2);
  expect(bridge.deliverNotification).toHaveBeenLastCalledWith(
    expect.objectContaining({ notificationId: "receipt-one", silent: false }),
  );
});

it("does not let an asynchronous session bind deliver after disposal", async () => {
  const bridge = native();
  let finish: ((bound: boolean) => void) | undefined;
  jest.mocked(bindDesktopSession).mockReturnValue(
    new Promise((resolve) => {
      finish = resolve;
    }),
  );
  const delivery = makeDelivery();
  delivery.receive(message());
  await jest.advanceTimersByTimeAsync(800);
  delivery.dispose();
  finish?.(true);
  await jest.advanceTimersByTimeAsync(1_000);
  expect(bridge.deliverNotification).not.toHaveBeenCalled();
  expect(playChatSound).not.toHaveBeenCalled();
});

it("cancels pending bursts, retries and obsolete scope deliveries", async () => {
  const delivery = makeDelivery();
  delivery.receive(message());
  delivery.dispose();
  await jest.advanceTimersByTimeAsync(800);
  expect(browserNotification).not.toHaveBeenCalled();
  let current = true;
  const scoped = makeDelivery(() => current);
  scoped.receive(message());
  current = false;
  await jest.advanceTimersByTimeAsync(800);
  expect(browserNotification).not.toHaveBeenCalled();
});

it("preserves the exact thread/message destination", () => {
  expect(
    chatNotificationPath(
      message("message", { threadRootId: "thread" }),
      scope.workspaceSlug,
    ),
  ).toBe(
    "/engineering/chat?conversation=conversation&message=message&thread=thread",
  );
});
