"use client";

import { useCallback, useEffect, useState } from "react";
import type { ChatAlertPreferences } from "./preferences";
import {
  CHAT_ALERT_PREFERENCES_CHANGED,
  DEFAULT_CHAT_ALERT_PREFERENCES,
  chatAlertPreferencesKey,
  readChatAlertPreferences,
  saveChatAlertPreferences,
} from "./preferences";

export const useChatAlertPreferences = (userId: string, workspace: string) => {
  const scope = chatAlertPreferencesKey(userId, workspace);
  const [snapshot, setSnapshot] = useState(() => ({
    scope,
    preferences: DEFAULT_CHAT_ALERT_PREFERENCES,
  }));
  const preferences =
    snapshot.scope === scope
      ? snapshot.preferences
      : DEFAULT_CHAT_ALERT_PREFERENCES;

  useEffect(() => {
    const refresh = () => {
      setSnapshot({
        scope,
        preferences: readChatAlertPreferences(userId, workspace),
      });
    };
    refresh();
    window.addEventListener("storage", refresh);
    window.addEventListener(CHAT_ALERT_PREFERENCES_CHANGED, refresh);
    return () => {
      window.removeEventListener("storage", refresh);
      window.removeEventListener(CHAT_ALERT_PREFERENCES_CHANGED, refresh);
    };
  }, [scope, userId, workspace]);

  const updatePreferences = useCallback(
    (patch: Partial<ChatAlertPreferences>) => {
      if (!userId || !workspace) {
        throw new Error("Sign in to update your chat notification settings.");
      }
      saveChatAlertPreferences(userId, workspace, {
        ...readChatAlertPreferences(userId, workspace),
        ...patch,
      });
    },
    [userId, workspace],
  );

  return { preferences, updatePreferences };
};
