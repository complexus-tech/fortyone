import type { Doc } from "yjs";
import { HocuspocusProvider } from "@hocuspocus/provider";

// A sync message alone is insufficient: read-only peers also receive it. Only
// clear the local dirty flag after write authorization and the server's ACK.
export function uploadCollaboration(
  ticket: { url: string; name: string; token: string },
  document: Doc,
  signal: AbortSignal,
): Promise<boolean> {
  return new Promise((resolve) => {
    if (signal.aborted) {
      resolve(false);
      return;
    }
    let finished = false;
    let writable = false;
    let provider: HocuspocusProvider | undefined;
    let timeout: number | undefined;
    const finish = (success: boolean) => {
      if (finished) return;
      finished = true;
      window.clearTimeout(timeout);
      signal.removeEventListener("abort", cancel);
      provider?.destroy();
      resolve(success);
    };
    const cancel = () => finish(false);
    const acknowledged = () => {
      if (writable && provider?.isSynced && !provider.hasUnsyncedChanges)
        finish(true);
    };
    signal.addEventListener("abort", cancel, { once: true });
    provider = new HocuspocusProvider({
      ...ticket,
      document,
      onAuthenticated: ({ scope }) => {
        writable = scope === "read-write";
        if (!writable) finish(false);
        else acknowledged();
      },
      onSynced: acknowledged,
      onUnsyncedChanges: acknowledged,
      onAuthenticationFailed: () => finish(false),
    });
    if (finished) provider.destroy();
    else timeout = window.setTimeout(() => finish(false), 20_000);
  });
}
