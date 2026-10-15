import { useAccountNative } from "./use-account-native";
import { useCallback, useEffect, useRef, useState } from "react";
import type { DraftKind, JsonRecord } from "../types";
import { errorMessage } from "../native";
import { persistDraft } from "../drafts";

export function useDraft<T extends JsonRecord>(
  workspaceId: string,
  resourceId: string,
  kind: DraftKind,
  initial: T,
) {
  const client = useAccountNative();
  const [value, setValue] = useState(initial);
  const [ready, setReady] = useState(false);
  const [readAttempt, setReadAttempt] = useState(0);
  const [status, setStatus] = useState("Loading saved draft…");
  const [error, setError] = useState<string | null>(null);
  const sequence = useRef(0);
  const initialRef = useRef(initial);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    setError(null);
    let active = true;
    void client
      .readDraft(workspaceId, resourceId, kind)
      .then((draft) => {
        if (!active) return;
        setValue(
          draft
            ? ({ ...initialRef.current, ...draft.payload } as T)
            : initialRef.current,
        );
        setStatus(draft ? "Draft saved on this device" : "");
        setReady(true);
      })
      .catch((reason: unknown) => {
        if (!active) return;
        setError(`Could not recover your draft: ${errorMessage(reason)}`);
      });
    return () => {
      active = false;
      mounted.current = false;
    };
  }, [client, workspaceId, resourceId, kind, readAttempt]);

  useEffect(() => {
    const saved = (event: Event) => {
      const detail = (
        event as CustomEvent<
          | {
              accountId: string;
              workspaceId: string;
              resourceId: string;
              kind: DraftKind;
            }
          | undefined
        >
      ).detail;
      if (
        detail?.accountId === client.accountId &&
        detail.workspaceId === workspaceId &&
        detail.resourceId === resourceId &&
        detail.kind === kind
      ) {
        setStatus("Draft saved on this device");
        setError(null);
      }
    };
    window.addEventListener("desktop-draft-saved", saved);
    return () => window.removeEventListener("desktop-draft-saved", saved);
  }, [client, workspaceId, resourceId, kind]);

  const update = useCallback(
    (next: T) => {
      const version = ++sequence.current;
      setValue(next);
      setError(null);
      setStatus("Saving on this device…");
      const write = persistDraft(
        workspaceId,
        resourceId,
        kind,
        next,
        client.accountId,
      );
      void write.then(
        () => {
          if (mounted.current && sequence.current === version)
            setStatus("Draft saved on this device");
        },
        (reason: unknown) => {
          if (mounted.current && sequence.current === version) {
            setStatus("Not saved");
            setError(errorMessage(reason));
          }
        },
      );
    },
    [client, workspaceId, resourceId, kind],
  );

  return {
    value,
    update,
    ready,
    status,
    error,
    retryRead: () => setReadAttempt((attempt) => attempt + 1),
  };
}
