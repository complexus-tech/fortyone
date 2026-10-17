import {
  DEFAULT_CHAT_ALERT_PREFERENCES,
  isChatQuietTime,
  readChatAlertPreferences,
  saveChatAlertPreferences,
} from "./preferences";

beforeEach(() => {
  localStorage.clear();
});

it("keeps sound/mute settings scoped to the account and workspace", () => {
  saveChatAlertPreferences("alice", "Engineering", {
    ...DEFAULT_CHAT_ALERT_PREFERENCES,
    sound: false,
    mutedConversations: ["team"],
  });
  expect(readChatAlertPreferences("alice", "engineering").sound).toBe(false);
  expect(readChatAlertPreferences("bob", "engineering").sound).toBe(true);
  expect(
    readChatAlertPreferences("alice", "marketing").mutedConversations,
  ).toEqual([]);
});

it.each([
  [23, true],
  [7, true],
  [8, false],
  [12, false],
  [22, true],
])("handles quiet hours crossing midnight at %s:00", (hour, quiet) => {
  expect(
    isChatQuietTime(
      { ...DEFAULT_CHAT_ALERT_PREFERENCES, quietHours: true },
      new Date(2026, 9, 17, hour),
    ),
  ).toBe(quiet);
});

it("quiet hours never change unread state or the saved sound choice", () => {
  const preferences = { ...DEFAULT_CHAT_ALERT_PREFERENCES, quietHours: true };
  expect(isChatQuietTime(preferences, new Date(2026, 9, 17, 23))).toBe(true);
  expect(preferences.sound).toBe(true);
});
