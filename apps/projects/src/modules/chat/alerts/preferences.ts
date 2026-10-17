export type ChatAlertPreferences = {
  desktop: boolean;
  sound: boolean;
  mutedConversations: string[];
  quietHours: boolean;
  quietStart: string;
  quietEnd: string;
};

export const DEFAULT_CHAT_ALERT_PREFERENCES: ChatAlertPreferences = {
  desktop: true,
  sound: true,
  mutedConversations: [],
  quietHours: false,
  quietStart: "22:00",
  quietEnd: "08:00",
};
export const CHAT_ALERT_PREFERENCES_CHANGED = "fortyone:chat-alert-preferences";
export const chatAlertPreferencesKey = (accountId: string, workspace: string) =>
  `fortyone:chat-alerts:v1:${accountId}:${workspace.toLowerCase()}`;

const isTime = (value: unknown): value is string =>
  typeof value === "string" && /^(?:[01]\d|2[0-3]):[0-5]\d$/.test(value);

export const readChatAlertPreferences = (
  accountId: string,
  workspace: string,
): ChatAlertPreferences => {
  if (typeof window === "undefined") return DEFAULT_CHAT_ALERT_PREFERENCES;
  try {
    const raw = localStorage.getItem(
      chatAlertPreferencesKey(accountId, workspace),
    );
    const stored: unknown = raw ? JSON.parse(raw) : null;
    if (!stored || typeof stored !== "object" || Array.isArray(stored))
      return DEFAULT_CHAT_ALERT_PREFERENCES;
    const value = stored as Record<string, unknown>;
    return {
      desktop: typeof value.desktop === "boolean" ? value.desktop : true,
      sound: typeof value.sound === "boolean" ? value.sound : true,
      mutedConversations: Array.isArray(value.mutedConversations)
        ? value.mutedConversations.filter(
            (id): id is string => typeof id === "string",
          )
        : [],
      quietHours: value.quietHours === true,
      quietStart: isTime(value.quietStart) ? value.quietStart : "22:00",
      quietEnd: isTime(value.quietEnd) ? value.quietEnd : "08:00",
    };
  } catch {
    // Storage may be unavailable in private browsing. Keep defaults usable.
    return DEFAULT_CHAT_ALERT_PREFERENCES;
  }
};

export const saveChatAlertPreferences = (
  accountId: string,
  workspace: string,
  preferences: ChatAlertPreferences,
) => {
  localStorage.setItem(
    chatAlertPreferencesKey(accountId, workspace),
    JSON.stringify(preferences),
  );
  window.dispatchEvent(new Event(CHAT_ALERT_PREFERENCES_CHANGED));
};

export const isChatQuietTime = (
  preferences: ChatAlertPreferences,
  date = new Date(),
) => {
  if (!preferences.quietHours) return false;
  const minutes = (time: string) => {
    const [hours, minute] = time.split(":").map(Number);
    return hours * 60 + minute;
  };
  const start = minutes(preferences.quietStart);
  const end = minutes(preferences.quietEnd);
  const now = date.getHours() * 60 + date.getMinutes();
  return (
    start === end ||
    (start < end ? now >= start && now < end : now >= start || now < end)
  );
};
