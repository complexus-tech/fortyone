import { useRef, useState } from "react";
import { retryFailedDrafts } from "../drafts";
import { errorMessage, getBoundAccountId, native } from "../native";

export function useBrowserSignIn(accountId: string | null) {
  const [busy, setBusy] = useState(false);
  const [opened, setOpened] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inFlight = useRef(false);
  const begin = async () => {
    if (inFlight.current) return;
    inFlight.current = true;
    setBusy(true);
    setError(null);
    setOpened(false);
    try {
      if (accountId) {
        if (getBoundAccountId() !== accountId)
          throw new Error(
            "Your account changed. Reopen Settings before reconnecting.",
          );
        await retryFailedDrafts(accountId);
        if (getBoundAccountId() !== accountId)
          throw new Error(
            "Your account changed. Reopen Settings before reconnecting.",
          );
      }
      await native.signIn();
      setOpened(true);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      inFlight.current = false;
      setBusy(false);
    }
  };
  return { busy, opened, error, begin };
}

export type BrowserSignIn = ReturnType<typeof useBrowserSignIn>;
