import { toast } from "sonner";
import { getDesktopBridge } from "./bridge";
import { getDesktopOnline } from "./connectivity";

let navigation: Promise<void> | null = null;

export function isDesktopOfflineNavigation(href: string): boolean {
  const bridge = getDesktopBridge();
  if (getDesktopOnline() || !bridge?.restoreSession || !bridge.pageRead)
    return false;
  try {
    const url = new URL(href, location.href);
    return url.origin === location.origin && url.pathname !== location.pathname;
  } catch {
    return false;
  }
}

/** All active editors participate in the same durable boundary as native close. */
async function flushEditors() {
  const writes: Promise<unknown>[] = [];
  window.dispatchEvent(
    new CustomEvent("fortyone:desktop-before-close", {
      detail: {
        waitUntil: (write: Promise<unknown>) => {
          writes.push(Promise.resolve(write));
        },
      },
    }),
  );
  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    await Promise.race([
      Promise.all(writes),
      new Promise<never>((_, reject) => {
        timer = setTimeout(() => {
          reject(
            new Error(
              "Saving took too long. Your current page stayed open; retry saving before opening another page.",
            ),
          );
        }, 12_000);
      }),
    ]);
  } finally {
    clearTimeout(timer);
  }
}

function openSavedPage(href: string) {
  return new Promise<void>((resolve, reject) => {
    const done = () => {
      clearTimeout(timeout);
      window.removeEventListener("pagehide", done);
      resolve();
    };
    const timeout = setTimeout(() => {
      window.removeEventListener("pagehide", done);
      window.stop();
      reject(
        new Error(
          "The page did not open. Your current page stayed open; try again.",
        ),
      );
    }, 10_000);
    window.addEventListener("pagehide", done);
    try {
      location.assign(href);
    } catch (error) {
      clearTimeout(timeout);
      window.removeEventListener("pagehide", done);
      reject(error instanceof Error ? error : new Error(String(error)));
    }
  });
}

/** Normal web/online navigation keeps its router behavior. Offline desktop
 * navigation uses a native-verified saved page after local editor writes settle. */
export function navigateWithDesktopOffline(
  href: string,
  fallback: () => void,
): Promise<void> {
  if (!isDesktopOfflineNavigation(href)) {
    fallback();
    return Promise.resolve();
  }
  if (navigation) return navigation;
  const bridge = getDesktopBridge()!;
  const source = location.href;
  const destination = new URL(href, source);
  // Fragments do not identify a stored page and must not enter the native lookup.
  destination.hash = "";
  const workspaceSlug =
    location.hostname.endsWith(".fortyone.app") &&
    location.hostname !== "cloud.fortyone.app"
      ? location.hostname.slice(0, -".fortyone.app".length)
      : location.pathname.split("/")[1];
  const body = document.body;
  const previouslyInert = body.inert;
  body.inert = true;
  navigation = (async () => {
    const scope = await bridge.restoreSession!({ workspaceSlug });
    const page = await bridge.pageRead!({
      accountId: scope.accountId,
      workspaceId: scope.workspaceId,
      url: destination.href,
    });
    if (!page)
      throw new Error(
        "This page has not been downloaded yet. Reconnect to open it; your current page stayed open.",
      );
    await flushEditors();
    const current = await bridge.restoreSession!({ workspaceSlug });
    if (
      source !== location.href ||
      current.accountId !== scope.accountId ||
      current.workspaceId !== scope.workspaceId
    )
      throw new Error(
        "Your account or workspace changed. Open the page again from the current workspace.",
      );
    await openSavedPage(destination.href);
  })()
    .catch((reason: unknown) => {
      toast.error("Could not open the downloaded page", {
        description: reason instanceof Error ? reason.message : String(reason),
      });
    })
    .finally(() => {
      body.inert = previouslyInert;
      navigation = null;
    });
  return navigation;
}
