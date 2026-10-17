import { useCallback, useRef } from "react";
import { useDebouncedCallback } from "@/hooks/debounce";
import type { DocumentUpdate } from "./types";
import { useUpdateDocument } from "./hooks";

/** Serializes legacy HTTP saves so each request uses the last committed revision. */
export function useDocumentAutosave(documentId: string, enabled: boolean) {
  const updateDocument = useUpdateDocument(documentId);
  const editingRevision = useRef(0);
  const saveQueue = useRef<Promise<void> | null>(null);
  const saveFailed = useRef(false);
  const initializeRevision = useCallback((revision: number) => {
    editingRevision.current = revision;
  }, []);
  const persist = (payload: Omit<DocumentUpdate, "expectedRevision">) => {
    if (!enabled) return Promise.resolve();
    const pending = (saveQueue.current ?? Promise.resolve()).then(async () => {
      if (saveFailed.current) throw new Error("Reload before saving again");
      const response = await updateDocument.mutateAsync({
        ...payload,
        expectedRevision: editingRevision.current,
      });
      if (!response.data) throw new Error("Document was not saved");
      editingRevision.current = response.data.revision;
    });
    saveQueue.current = pending.catch(() => {
      saveFailed.current = true;
    });
    return pending;
  };
  const { callback: saveTitle, flush: flushTitle } = useDebouncedCallback(
    (nextTitle: string) => {
      void persist({ title: nextTitle }).catch(() => {});
    },
    700,
    { flushOnUnmount: true },
  );
  const { callback: saveContent, flush: flushContent } = useDebouncedCallback(
    (content: Pick<DocumentUpdate, "contentHtml" | "contentText">) => {
      void persist(content).catch(() => {});
    },
    700,
    { flushOnUnmount: true },
  );
  return {
    updateDocument,
    persist,
    saveTitle,
    flushTitle,
    saveContent,
    flushContent,
    initializeRevision,
  };
}
