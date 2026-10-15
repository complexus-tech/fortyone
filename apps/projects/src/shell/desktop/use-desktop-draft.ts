"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";
import { useDesktopWorkspaceScope } from "./workspace-scope";
import {
  removeDesktopDraft,
  retryDesktopDrafts,
  writeDesktopDraft,
} from "./drafts";

export function useDesktopDraft<T extends object>(key: string, enabled = true) {
  const scope = useDesktopWorkspaceScope();
  const supported = Boolean(
    scope && getDesktopBridge()?.draftRead && getDesktopBridge()?.draftWrite,
  );
  const [recovered, setRecovered] = useState<T | null>(null);
  const [ready, setReady] = useState(!supported);
  const [status, setStatus] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const sequence = useRef(0);
  const mounted = useRef(true);
  useEffect(() => {
    mounted.current = true;
    sequence.current++;
    if (!supported || !enabled || !scope) return;
    let active = true;
    setReady(false);
    setStatus("");
    setError(null);
    setRecovered(null);
    const load = async () => {
      await bindDesktopSession(scope);
      const payload = await getDesktopBridge()?.draftRead?.({
        accountId: scope.accountId,
        workspaceId: scope.workspaceId,
        key,
      });
      if (!active) return;
      if (payload && typeof payload === "object" && !Array.isArray(payload))
        setRecovered(payload as T);
      setReady(true);
    };
    void load().catch((reason: unknown) => {
      if (active) setError(String(reason));
    });
    return () => {
      active = false;
      mounted.current = false;
    };
  }, [scope, key, supported, enabled, attempt]);
  const save = useCallback(
    (payload: T) => {
      if (!supported || !enabled || !scope) return Promise.resolve();
      const version = ++sequence.current;
      setStatus("Saving draft…");
      setError(null);
      const write = writeDesktopDraft(scope, key, payload);
      void write.then(
        () => {
          if (mounted.current && sequence.current === version)
            setStatus("Draft saved on this device");
        },
        (reason: unknown) => {
          if (mounted.current && sequence.current === version) {
            setStatus("Draft not saved");
            setError(String(reason));
          }
        },
      );
      return write;
    },
    [enabled, key, scope, supported],
  );
  const discard = useCallback(async () => {
    if (!supported || !scope) return;
    const version = sequence.current;
    await removeDesktopDraft(scope, key);
    if (!mounted.current || sequence.current !== version) return;
    setRecovered(null);
    setStatus("");
  }, [key, scope, supported]);
  const retry = async () => {
    setError(null);
    try {
      if (scope) await retryDesktopDrafts(scope);
      if (!ready) setAttempt((value) => value + 1);
      else setStatus("Draft saved on this device");
    } catch (reason) {
      setError(String(reason));
    }
  };
  return {
    supported,
    ready,
    recovered,
    status,
    error,
    save,
    discard,
    retry,
    acceptRecovered: () => {
      setRecovered(null);
    },
  };
}
