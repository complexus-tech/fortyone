import { update } from "idb-keyval";
import type { DesktopNotificationScope } from "@/shell/desktop/bridge";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { bindDesktopSession } from "@/shell/desktop/session-binding";
import { wasChatMessageVisible } from "./active-view";
import type { ChatNotificationEvent } from "./contract";
import { isChatQuietTime, readChatAlertPreferences } from "./preferences";
import { canPlayChatSound, playChatSound } from "./sound";

const MAX_EVENT_AGE_MS = 60_000;
const BURST_WINDOW_MS = 800;
const SOUND_COOLDOWN_MS = 3_000;
const BACKGROUND_SOUND_DELAY_MS = 150;
const RETRY_DELAY_MS = 2_000;

export const isLiveChatNotification = (
  event: ChatNotificationEvent,
  scope: DesktopNotificationScope,
  now = Date.now(),
) => {
  const age = now - Date.parse(event.occurredAt);
  return (
    event.workspaceId === scope.workspaceId &&
    event.userId === scope.accountId &&
    event.actorId !== scope.accountId &&
    age >= -5_000 &&
    age <= MAX_EVENT_AGE_MS
  );
};

export const chatNotificationPath = (
  event: ChatNotificationEvent,
  workspaceSlug: string,
) => {
  const params = new URLSearchParams({
    conversation: event.conversationId,
    message: event.messageId,
  });
  if (event.threadRootId) params.set("thread", event.threadRootId);
  const hostname =
    typeof window === "undefined" ? "" : window.location.hostname;
  const prefix =
    hostname.endsWith(".fortyone.app") && hostname !== "cloud.fortyone.app"
      ? ""
      : `/${encodeURIComponent(workspaceSlug)}`;
  return `${prefix}/chat?${params.toString()}`;
};

type Receipt = { at: number; owner: string };
type Receipts = {
  events: Record<string, Receipt>;
  lastSound?: Receipt;
};
type Claim = {
  key: string;
  owner: string;
  events: ChatNotificationEvent[];
  sound: boolean;
  previousSound?: Receipt;
};

/** Banner and sound claims are independent: a silent tab cannot consume audio. */
const claim = async (
  kind: "banner" | "sound",
  events: ChatNotificationEvent[],
  scope: DesktopNotificationScope,
): Promise<Claim> => {
  const result: Claim = {
    key: `chat-alert-${kind}:v2:${scope.accountId}:${scope.workspaceId}`,
    owner: crypto.randomUUID(),
    events: [],
    sound: false,
  };
  const now = Date.now();
  await update<Receipts>(result.key, (stored) => {
    const receipts = Object.fromEntries(
      Object.entries(stored?.events ?? {}).filter(
        ([, receipt]) => now - receipt.at < MAX_EVENT_AGE_MS * 2,
      ),
    );
    for (const event of events) {
      if (
        Object.hasOwn(receipts, event.eventId) ||
        !isLiveChatNotification(event, scope, now)
      )
        continue;
      receipts[event.eventId] = { at: now, owner: result.owner };
      result.events.push(event);
    }
    result.previousSound = stored?.lastSound;
    result.sound =
      kind === "sound" &&
      result.events.length > 0 &&
      now - (stored?.lastSound?.at ?? 0) >= SOUND_COOLDOWN_MS;
    return {
      events: receipts,
      lastSound: result.sound
        ? { at: now, owner: result.owner }
        : stored?.lastSound,
    };
  });
  return result;
};

/** Failed presentation can retry without clearing a different tab's receipt. */
const release = async (claimed: Claim) => {
  await update<Receipts>(claimed.key, (stored) => ({
    events: Object.fromEntries(
      Object.entries(stored?.events ?? {}).filter(
        ([, receipt]) => receipt.owner !== claimed.owner,
      ),
    ),
    lastSound:
      stored?.lastSound?.owner === claimed.owner
        ? claimed.previousSound
        : stored?.lastSound,
  }));
};

const preferencesFor = (
  scope: DesktopNotificationScope,
  conversationId: string,
) => {
  const preferences = readChatAlertPreferences(
    scope.accountId,
    scope.workspaceSlug,
  );
  return preferences.mutedConversations.includes(conversationId) ||
    isChatQuietTime(preferences)
    ? null
    : preferences;
};

const browserSound = async (
  events: ChatNotificationEvent[],
  scope: DesktopNotificationScope,
  isCurrent: () => boolean,
) => {
  // Give a focused, unlocked tab first opportunity. An unlocked background tab
  // remains a fallback when the user is working in another application.
  if (!document.hasFocus())
    await new Promise<void>((resolve) => {
      setTimeout(resolve, BACKGROUND_SOUND_DELAY_MS);
    });
  if (
    !isCurrent() ||
    !canPlayChatSound() ||
    !preferencesFor(scope, events[0].conversationId)?.sound
  )
    return;
  const claimed = await claim("sound", events, scope);
  if (!claimed.sound) return;
  if (
    !isCurrent() ||
    !preferencesFor(scope, events[0].conversationId)?.sound ||
    !playChatSound()
  )
    await release(claimed);
};

const present = async (
  events: ChatNotificationEvent[],
  scope: DesktopNotificationScope,
  isCurrent: () => boolean,
) => {
  if (!isCurrent()) return;
  const preferences = preferencesFor(scope, events[0].conversationId);
  if (!preferences) return;
  const unseen = events.filter(
    (event) =>
      !wasChatMessageVisible(
        scope.accountId,
        scope.workspaceSlug,
        event.messageId,
      ),
  );
  const bridge = getDesktopBridge();
  const nativeEnabled =
    bridge?.deliverNotification &&
    (!bridge.getPreferences || (await bridge.getPreferences()).notifications);
  const browserEnabled =
    !bridge &&
    typeof Notification !== "undefined" &&
    Notification.permission === "granted";
  if (
    !unseen.length ||
    !preferences.desktop ||
    (!nativeEnabled && !browserEnabled)
  ) {
    await browserSound(events, scope, isCurrent);
    return;
  }

  const banner = await claim("banner", unseen, scope);
  let soundClaim: Claim | undefined;
  try {
    if (banner.events.length && isCurrent()) {
      const event = [...banner.events].sort(
        (a, b) => Date.parse(a.occurredAt) - Date.parse(b.occurredAt),
      )[banner.events.length - 1];
      const body = event.body;
      if (bridge?.deliverNotification) {
        if (!(await bindDesktopSession(scope)) || !isCurrent()) {
          await release(banner);
          return;
        }
        const currentPreferences = preferencesFor(scope, event.conversationId);
        if (!currentPreferences?.desktop) {
          await release(banner);
          return;
        }
        if (currentPreferences.sound)
          soundClaim = await claim("sound", events, scope);
        if (!isCurrent()) {
          await release(banner);
          if (soundClaim) await release(soundClaim);
          return;
        }
        const notification = {
          ...scope,
          notificationId: event.notificationId,
          createdAt: event.createdAt,
          entityType: "chat_message",
          entityId: event.messageId,
          title: event.title,
          body,
        };
        // Older shells cannot request silence. Do not override the sound choice.
        if (bridge.chatNotificationLinks) {
          await bridge.deliverNotification({
            ...notification,
            conversationId: event.conversationId,
            ...(event.threadRootId ? { threadRootId: event.threadRootId } : {}),
            silent: !soundClaim?.sound,
          });
        } else if (soundClaim?.sound) {
          await bridge.deliverNotification(notification);
        }
      } else if (isCurrent()) {
        const notification = new Notification(event.title, {
          body,
          tag: `chat:${scope.workspaceId}:${event.conversationId}`,
          silent: true,
        });
        notification.onclick = () => {
          notification.close();
          if (!isCurrent()) return;
          window.focus();
          window.location.assign(
            chatNotificationPath(event, scope.workspaceSlug),
          );
        };
      }
    } else if (!isCurrent()) {
      await release(banner);
    }
  } catch (error) {
    await release(banner);
    if (soundClaim) await release(soundClaim);
    throw error;
  }
  // Every browser tab gets a chance at audio, even if another owns the banner.
  if (!bridge) await browserSound(events, scope, isCurrent);
};

export const createChatAlertDelivery = (
  scope: DesktopNotificationScope,
  isCurrent: () => boolean,
  onError: (error: unknown) => void,
) => {
  const pending = new Map<
    string,
    {
      events: Map<string, ChatNotificationEvent>;
      timer: ReturnType<typeof setTimeout>;
    }
  >();
  const retries = new Set<ReturnType<typeof setTimeout>>();
  let disposed = false;
  const active = () => !disposed && isCurrent();
  const deliver = (events: ChatNotificationEvent[], attempt = 0) => {
    void present(events, scope, active).catch((error: unknown) => {
      if (!active()) return;
      onError(error);
      if (attempt >= 2) return;
      const timer = setTimeout(() => {
        retries.delete(timer);
        if (active()) deliver(events, attempt + 1);
      }, RETRY_DELAY_MS);
      retries.add(timer);
    });
  };
  return {
    receive(event: ChatNotificationEvent) {
      if (!active() || !isLiveChatNotification(event, scope)) return;
      const batch = pending.get(event.conversationId);
      if (batch) {
        batch.events.set(event.eventId, event);
        return;
      }
      const events = new Map([[event.eventId, event]]);
      const timer = setTimeout(() => {
        pending.delete(event.conversationId);
        deliver([...events.values()]);
      }, BURST_WINDOW_MS);
      pending.set(event.conversationId, { events, timer });
    },
    dispose() {
      disposed = true;
      for (const batch of pending.values()) clearTimeout(batch.timer);
      for (const timer of retries) clearTimeout(timer);
      pending.clear();
      retries.clear();
    },
  };
};
