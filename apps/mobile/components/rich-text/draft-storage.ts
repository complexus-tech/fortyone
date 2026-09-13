export const DRAFT_STORAGE_PREFIX = "fortyone:draft:v1:";

export type DraftStorage = {
  getItem: (key: string) => Promise<string | null>;
  setItem: (key: string, value: string) => Promise<void>;
  removeItem: (key: string) => Promise<void>;
};

export const getDraftKey = (
  userId: string,
  workspace: string,
  documentId: string,
) =>
  `${DRAFT_STORAGE_PREFIX}${[userId, workspace, documentId].map(encodeURIComponent).join(":")}`;

/** One ordered writer per key prevents a slow autosave from resurrecting a saved draft. */
export const createDraftRepository = (storage: DraftStorage) => {
  const pending = new Map<string, Promise<void>>();
  const enqueue = (key: string, action: () => Promise<void>) => {
    const previous = pending.get(key) ?? Promise.resolve();
    const operation = previous.catch(() => undefined).then(action);
    pending.set(key, operation);
    void operation.then(
      () => {
        if (pending.get(key) === operation) pending.delete(key);
      },
      () => {
        if (pending.get(key) === operation) pending.delete(key);
      },
    );
    return operation;
  };
  return {
    read: async <T>(
      key: string,
      isValid: (value: unknown) => value is T,
    ): Promise<T | null> => {
      await pending.get(key);
      const serialized = await storage.getItem(key);
      if (!serialized) return null;
      const value: unknown = JSON.parse(serialized);
      if (!isValid(value))
        throw new Error(
          "This saved draft cannot be opened. Your original content is unchanged.",
        );
      return value;
    },
    write: (key: string, value: unknown) =>
      enqueue(key, () => storage.setItem(key, JSON.stringify(value))),
    remove: (key: string) => enqueue(key, () => storage.removeItem(key)),
    // Cleanup still needs to run when an earlier write failed.
    flush: () => Promise.allSettled([...pending.values()]),
  };
};
