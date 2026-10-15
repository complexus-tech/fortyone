(() => {
  const url = new URL(location.href);
  url.searchParams.delete("__fortyone_offline");
  if (
    !document.documentElement.hasAttribute("data-fortyone-offline-bootstrap")
  ) {
    // Only native-verified saved HTML includes this script in the normal app.
    // The OS may still report online when the application servers are unreachable.
    window.__FORTYONE_OFFLINE_PAGE__ = true;
    history.replaceState(history.state, "", url.href);
    return;
  }
  const start = async () => {
    const status = document.getElementById("status");
    const retry = document.getElementById("retry");
    retry.hidden = true;
    try {
      const bridge = window.__FORTYONE_DESKTOP__;
      if (!bridge?.restoreSession || !bridge.pageRead)
        throw new Error(
          "Open FortyOne while connected to prepare this device for offline use.",
        );
      const host = location.hostname;
      const workspaceSlug =
        host.endsWith(".fortyone.app") && host !== "cloud.fortyone.app"
          ? host.slice(0, -".fortyone.app".length)
          : location.pathname.split("/")[1];
      const scope = await bridge.restoreSession({ workspaceSlug });
      const page = await bridge.pageRead({
        accountId: scope.accountId,
        workspaceId: scope.workspaceId,
        url: url.href,
      });
      if (!page)
        throw new Error(
          "This page has not been downloaded yet. Reconnect to open it.",
        );
      const worker = navigator.serviceWorker.controller;
      if (!worker)
        throw new Error("Reconnect to prepare this workspace for offline use.");
      const nonce = crypto.randomUUID();
      await new Promise((resolve, reject) => {
        const channel = new MessageChannel();
        const timeout = setTimeout(() => {
          channel.port1.close();
          reject(
            new Error("The offline workspace did not respond. Try again."),
          );
        }, 10_000);
        channel.port1.onmessage = ({ data }) => {
          clearTimeout(timeout);
          channel.port1.close();
          if (data?.ok) resolve();
          else reject(new Error(data?.error || "Reconnect to open this page."));
        };
        worker.postMessage(
          {
            type: "restore",
            scope,
            nonce,
            url: url.href,
            html: page.html,
            contentSecurityPolicy: page.contentSecurityPolicy,
          },
          [channel.port2],
        );
      });
      url.searchParams.set("__fortyone_offline", nonce);
      location.replace(url.href);
    } catch (error) {
      status.textContent =
        error instanceof Error
          ? error.message
          : "Reconnect to open this workspace.";
      retry.hidden = false;
    }
  };
  document
    .getElementById("retry")
    .addEventListener("click", () => location.replace(url.href));
  void start();
})();
