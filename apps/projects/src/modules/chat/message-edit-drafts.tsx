"use client";

import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useState,
  useSyncExternalStore,
} from "react";
import type { PreviewDraft } from "./chat-rich-text";

const DRAFT_SCOPES = ["conversation", "thread", "search"] as const;
export type MessageDraftScope = (typeof DRAFT_SCOPES)[number];

const draftKey = (messageId: string, scope: MessageDraftScope) =>
  JSON.stringify([messageId, scope]);

const createDraftStore = () => {
  const drafts = new Map<string, PreviewDraft>();
  const listeners = new Map<string, Set<() => void>>();
  const set = (key: string, draft: PreviewDraft | null) => {
    if ((drafts.get(key) ?? null) === draft) return;
    if (draft) drafts.set(key, draft);
    else drafts.delete(key);
    listeners.get(key)?.forEach((listener) => {
      listener();
    });
  };

  return {
    get: (key: string) => drafts.get(key) ?? null,
    set,
    subscribe: (key: string, listener: () => void) => {
      const subscribers = listeners.get(key) ?? new Set<() => void>();
      subscribers.add(listener);
      listeners.set(key, subscribers);
      return () => {
        subscribers.delete(listener);
        if (!subscribers.size) listeners.delete(key);
      };
    },
    clearMessage: (messageId: string) => {
      DRAFT_SCOPES.forEach((scope) => {
        set(draftKey(messageId, scope), null);
      });
    },
  };
};

const EditDraftContext = createContext<ReturnType<
  typeof createDraftStore
> | null>(null);
const getServerSnapshot = () => null;

/** Drafts survive virtual row unmounts, but never outlive this chat shell. */
export const MessageEditDraftProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  // eslint-disable-next-line react/hook-use-state -- Store identity is fixed; keyed subscribers own updates.
  const [store] = useState(createDraftStore);
  return (
    <EditDraftContext.Provider value={store}>
      {children}
    </EditDraftContext.Provider>
  );
};

export const useMessageEditDraft = (
  messageId: string,
  scope: MessageDraftScope = "conversation",
) => {
  const store = useContext(EditDraftContext);
  if (!store) {
    throw new Error("Message editing requires MessageEditDraftProvider.");
  }
  const key = draftKey(messageId, scope);
  const subscribe = useCallback(
    (listener: () => void) => store.subscribe(key, listener),
    [key, store],
  );
  const getSnapshot = useCallback(() => store.get(key), [key, store]);
  const draft = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const setDraft = useCallback(
    (value: PreviewDraft | null) => {
      store.set(key, value);
    },
    [key, store],
  );
  const clearMessageDrafts = useCallback(() => {
    store.clearMessage(messageId);
  }, [messageId, store]);

  return { draft, setDraft, clearMessageDrafts };
};
