const VISIBLE_FOR_MS = 10_000;
const visibleMessages = new Map<string, number>();
const flushes = new Map<string, ReturnType<typeof setTimeout>>();
const tabId = Math.random().toString(36).slice(2);

const key = (accountId: string, workspaceSlug: string, messageId: string) =>
  `${accountId}:${workspaceSlug.toLowerCase()}:${messageId}`;

/** Called by the timeline only for messages actually visible in a focused tab. */
export const recordChatMessageVisible = (
  accountId: string | undefined,
  workspaceSlug: string,
  message: { id: string },
) => {
  if (!accountId) return;
  const now = Date.now();
  for (const [id, seenAt] of visibleMessages) {
    if (now - seenAt > VISIBLE_FOR_MS) visibleMessages.delete(id);
  }
  visibleMessages.set(key(accountId, workspaceSlug, message.id), now);
  const scope = `${accountId}:${workspaceSlug.toLowerCase()}:`;
  if (!flushes.has(scope)) {
    flushes.set(
      scope,
      setTimeout(() => {
        flushes.delete(scope);
        try {
          const recent = [...visibleMessages]
            .filter(
              ([id, seenAt]) =>
                id.startsWith(scope) && Date.now() - seenAt < VISIBLE_FOR_MS,
            )
            .slice(-200);
          localStorage.setItem(
            `fortyone:chat-visible:${scope}${tabId}`,
            JSON.stringify(Object.fromEntries(recent)),
          );
          const prefix = `fortyone:chat-visible:${scope}`;
          // Each tab owns a key, so concurrent flushes cannot erase each other.
          for (const storedKey of Object.keys(localStorage)) {
            if (!storedKey.startsWith(prefix)) continue;
            const stored: unknown = JSON.parse(
              localStorage.getItem(storedKey) ?? "null",
            );
            if (
              !stored ||
              typeof stored !== "object" ||
              !Object.values(stored).some(
                (at: unknown) =>
                  typeof at === "number" && Date.now() - at < VISIBLE_FOR_MS,
              )
            ) {
              localStorage.removeItem(storedKey);
            }
          }
        } catch {
          // A tab can still suppress its own visible messages when storage is blocked.
        }
      }, 100),
    );
  }
};

export const wasChatMessageVisible = (
  accountId: string,
  workspaceSlug: string,
  messageId: string,
) => {
  const seenAt = visibleMessages.get(key(accountId, workspaceSlug, messageId));
  if (seenAt !== undefined && Date.now() - seenAt < VISIBLE_FOR_MS) return true;
  try {
    const prefix = `fortyone:chat-visible:${accountId}:${workspaceSlug.toLowerCase()}:`;
    for (const storedKey of Object.keys(localStorage)) {
      if (!storedKey.startsWith(prefix)) continue;
      const stored: unknown = JSON.parse(
        localStorage.getItem(storedKey) ?? "null",
      );
      if (!stored || typeof stored !== "object") continue;
      const at = (stored as Record<string, unknown>)[
        key(accountId, workspaceSlug, messageId)
      ];
      if (typeof at === "number" && Date.now() - at < VISIBLE_FOR_MS)
        return true;
    }
    return false;
  } catch {
    return false;
  }
};
