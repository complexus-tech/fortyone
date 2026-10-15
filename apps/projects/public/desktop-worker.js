/* Registered by the desktop shell only. Private HTML is stored in native SQLite,
 * never CacheStorage. One-use navigation permits contain native-authorized HTML
 * in memory just long enough to hand it to the existing Next interface. */
const PREFIX = "fortyone-desktop-shell-v2-";
const ASSETS = `${PREFIX}assets`;
const BOOTSTRAP = "/desktop-offline.js?v=2";
const MAX_AGE = 7 * 24 * 60 * 60 * 1000;
const MAX_PAGE_BYTES = 4 * 1024 * 1024;
const MAX_ASSET_BYTES = 8 * 1024 * 1024;
const MAX_TOTAL_ASSET_BYTES = 128 * 1024 * 1024;
const NAVIGATION_TIMEOUT = 8_000;
const permits = new Map();
const downloads = new Map();
let generation = 0;
let assetWrites = Promise.resolve();
const sections = new Set([
  "maya",
  "my-work",
  "summary",
  "work",
  "story",
  "docs",
  "teams",
  "sprints",
  "notifications",
  "settings",
  "profile",
  "views",
  "search",
  "calendar",
  "strategy",
  "roadmap",
  "analytics",
  "key-results",
]);
const UUID = /^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i;

function validScope(value) {
  return (
    value &&
    UUID.test(value.accountId) &&
    UUID.test(value.workspaceId) &&
    /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(value.workspaceSlug) &&
    Number.isFinite(value.expiresAt) &&
    value.expiresAt > Date.now() &&
    value.expiresAt <= Date.now() + MAX_AGE + 60_000
  );
}

function workspaceSlug(url) {
  return url.hostname.endsWith(".fortyone.app") &&
    url.hostname !== "cloud.fortyone.app"
    ? url.hostname.slice(0, -".fortyone.app".length)
    : url.pathname.split("/")[1];
}

function pageURL(input, scope) {
  const url = new URL(input, self.location.origin);
  if (
    url.origin !== self.location.origin ||
    url.username ||
    url.password ||
    workspaceSlug(url) !== scope.workspaceSlug
  )
    return null;
  const subdomain =
    url.hostname.endsWith(".fortyone.app") &&
    url.hostname !== "cloud.fortyone.app";
  const parts = url.pathname.split("/").slice(subdomain ? 1 : 2);
  if (!sections.has(parts[0])) return null;
  if (
    parts[0] === "settings" &&
    ![
      "",
      "account",
      "account/preferences",
      "account/notifications",
      "workspace/members",
    ].includes(parts.slice(1).join("/").replace(/\/$/, ""))
  )
    return null;
  for (const key of url.searchParams.keys()) {
    if (/token|secret|password|code|callback|authorization/i.test(key))
      return null;
  }
  url.searchParams.delete("__fortyone_offline");
  url.searchParams.delete("_rsc");
  url.hash = "";
  return url;
}

function assetURL(input, base = self.location.origin) {
  const url = new URL(input, base);
  return url.origin === self.location.origin &&
    url.pathname.startsWith("/_next/static/")
    ? url
    : null;
}

async function storeAsset(url, response) {
  if (!response.ok)
    throw new Error("A required app file could not be downloaded.");
  const body = await response.arrayBuffer();
  if (body.byteLength > MAX_ASSET_BYTES)
    throw new Error("An app file exceeds the offline limit.");
  const headers = new Headers(response.headers);
  headers.delete("content-encoding");
  headers.set("content-length", String(body.byteLength));
  headers.set("X-FortyOne-Bytes", String(body.byteLength));
  const write = assetWrites
    .catch(() => undefined)
    .then(async () => {
      const cache = await caches.open(ASSETS);
      const responses = await cache.matchAll();
      const total = responses.reduce(
        (sum, item) => sum + Number(item.headers.get("X-FortyOne-Bytes") || 0),
        0,
      );
      const previous = await cache.match(url.href);
      const previousSize = Number(
        previous?.headers.get("X-FortyOne-Bytes") || 0,
      );
      // Never evict a chunk still needed by a downloaded page. Report capacity
      // instead of silently breaking a workspace that was previously available.
      if (total - previousSize + body.byteLength > MAX_TOTAL_ASSET_BYTES) {
        throw new Error(
          "Offline app storage is full. Previously downloaded pages remain available.",
        );
      }
      await cache.put(url.href, new Response(body, { headers }));
    });
  assetWrites = write;
  await write;
  if (headers.get("content-type")?.includes("text/css")) {
    const css = new TextDecoder().decode(body);
    const dependencies = [
      ...css.matchAll(/url\(\s*["']?([^\s"')]+)["']?\s*\)/g),
    ]
      .map((match) => assetURL(match[1], url.href))
      .filter(Boolean);
    for (const dependency of dependencies) await saveAsset(dependency.href);
  }
}

async function saveAsset(input) {
  const url = assetURL(input);
  if (!url) return;
  const cache = await caches.open(ASSETS);
  if (!url.pathname.includes("/development/") && (await cache.match(url.href)))
    return;
  if (downloads.has(url.href)) return downloads.get(url.href);
  const download = (async () => {
    await storeAsset(
      url,
      await fetch(url.href, { credentials: "same-origin" }),
    );
  })();
  downloads.set(url.href, download);
  try {
    await download;
  } finally {
    downloads.delete(url.href);
  }
}

async function downloadPage(input, scope, startedGeneration) {
  const url = pageURL(input, scope);
  if (!url) throw new Error("This page cannot be downloaded for offline use.");
  const response = await fetch(url.href, {
    credentials: "same-origin",
    headers: { Accept: "text/html" },
  });
  if (
    !response.ok ||
    !pageURL(response.url || url.href, scope) ||
    !response.headers.get("content-type")?.includes("text/html")
  ) {
    throw new Error("Connect to download this page.");
  }
  const html = await response.text();
  if (
    new TextEncoder().encode(html).byteLength > MAX_PAGE_BYTES ||
    !html.includes("/_next/") ||
    html.includes("NEXT_REDIRECT") ||
    html.includes("NEXT_HTTP_ERROR_FALLBACK")
  ) {
    throw new Error("This page is not ready for offline use.");
  }
  const assets = [
    ...html.matchAll(/(?:src|href)=["']([^"']*\/_next\/static\/[^"']+)["']/g),
  ].map((match) => match[1].replaceAll("&amp;", "&"));
  const unique = [...new Set(assets)];
  for (let offset = 0; offset < unique.length; offset += 4) {
    await Promise.all(unique.slice(offset, offset + 4).map(saveAsset));
  }
  if (startedGeneration !== generation || !validScope(scope)) {
    throw new Error("The account changed before the download completed.");
  }
  // Native pageWrite rechecks the current HttpOnly cookie after this download.
  return {
    url: url.href,
    html,
    contentSecurityPolicy:
      response.headers.get("Content-Security-Policy") || undefined,
  };
}

self.addEventListener("install", (event) => {
  event.waitUntil(
    (async () => {
      await storeAsset(
        new URL(BOOTSTRAP, self.location.origin),
        await fetch(BOOTSTRAP),
      );
      // Updates wait for the previous worker's clients to close. Forcing
      // activation can interrupt Next streams and discard one-use restore permits.
    })(),
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      const names = await caches.keys();
      await Promise.all(
        names
          .filter(
            (name) =>
              name.startsWith("fortyone-desktop-shell-") &&
              !name.startsWith(PREFIX),
          )
          .map((name) => caches.delete(name)),
      );
      await self.clients.claim();
    })(),
  );
});

self.addEventListener("message", (event) => {
  const reply = (value) => event.ports[0]?.postMessage(value);
  event.waitUntil(
    (async () => {
      try {
        const client = event.source;
        if (
          !client ||
          client.type !== "window" ||
          new URL(client.url).origin !== self.location.origin ||
          (client.frameType && client.frameType !== "top-level")
        )
          throw new Error("Invalid desktop window.");
        const message = event.data;
        if (message?.type === "clear") {
          generation += 1;
          permits.clear();
          reply({ ok: true });
          return;
        }
        if (!validScope(message?.scope))
          throw new Error("Reconnect to renew offline access.");
        if (message.type === "warm") {
          const startedGeneration = generation;
          if (!Array.isArray(message.urls) || message.urls.length > 4)
            throw new Error("Too many pages requested.");
          const pages = [];
          for (const url of message.urls)
            pages.push(
              await downloadPage(url, message.scope, startedGeneration),
            );
          reply({ ok: true, pages });
        } else if (message.type === "restore") {
          const url = pageURL(message.url, message.scope);
          if (
            !url ||
            !UUID.test(message.nonce) ||
            typeof message.html !== "string" ||
            new TextEncoder().encode(message.html).byteLength >
              MAX_PAGE_BYTES ||
            !message.html.includes("<head")
          ) {
            throw new Error("Invalid offline navigation.");
          }
          permits.clear();
          permits.set(message.nonce, {
            scope: message.scope,
            url: url.href,
            clientId: client.id,
            html: message.html,
            contentSecurityPolicy:
              typeof message.contentSecurityPolicy === "string"
                ? message.contentSecurityPolicy
                : undefined,
            until: Date.now() + 15_000,
          });
          reply({ ok: true });
        } else throw new Error("Unsupported offline page request.");
      } catch (error) {
        reply({
          ok: false,
          error:
            error instanceof Error
              ? error.message
              : "Could not prepare offline pages.",
        });
      }
    })(),
  );
});

const offlinePage = () =>
  new Response(
    `<!doctype html>
<html lang="en" data-fortyone-offline-bootstrap><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>FortyOne</title>
<style>body{margin:0;background:Canvas;color:CanvasText;font:16px -apple-system,BlinkMacSystemFont,sans-serif;color-scheme:light dark;display:grid;place-items:center;min-height:100vh}main{max-width:440px;padding:32px}h1{font-size:24px}p{line-height:1.5;opacity:.75}button{font:inherit;padding:10px 18px;border:1px solid GrayText;border-radius:8px;background:ButtonFace;color:ButtonText;cursor:pointer}</style>
<script src="${BOOTSTRAP}" defer></script></head><body><main><h1>Opening your downloaded workspace</h1>
<p id="status" role="status">Checking this device’s saved session…</p><button id="retry" hidden>Try again</button></main></body></html>`,
    {
      headers: {
        "Content-Type": "text/html; charset=utf-8",
        "Cache-Control": "no-store",
      },
    },
  );

async function fetchNavigation(request) {
  const controller = new AbortController();
  const abort = () => controller.abort();
  if (request.signal?.aborted) abort();
  else request.signal?.addEventListener("abort", abort, { once: true });
  const timeout = setTimeout(abort, NAVIGATION_TIMEOUT);
  try {
    // Retain the navigation's credentials, headers and cache policy. Only its
    // deadline changes so an unreachable server cannot strand the offline boot.
    return await fetch(request, { signal: controller.signal });
  } finally {
    clearTimeout(timeout);
    request.signal?.removeEventListener("abort", abort);
  }
}

self.addEventListener("fetch", (event) => {
  const request = event.request;
  const url = new URL(request.url);
  if (request.method !== "GET" || url.origin !== self.location.origin) return;
  if (assetURL(url.href) || url.pathname === "/desktop-offline.js") {
    event.respondWith(
      (async () => {
        const cache = await caches.open(ASSETS);
        const cached = await cache.match(request);
        if (cached && !url.pathname.includes("/development/")) return cached;
        try {
          const response = await fetch(request);
          if (response.ok)
            await storeAsset(url, response.clone()).catch(() => undefined);
          return response;
        } catch (error) {
          if (cached) return cached;
          throw error;
        }
      })(),
    );
    return;
  }
  if (request.mode !== "navigate") return;
  event.respondWith(
    (async () => {
      const nonce = url.searchParams.get("__fortyone_offline");
      if (nonce) {
        const permit = permits.get(nonce);
        permits.delete(nonce);
        const canonical = permit && pageURL(url, permit.scope);
        if (
          permit &&
          permit.until >= Date.now() &&
          validScope(permit.scope) &&
          canonical?.href === permit.url &&
          (!event.clientId || event.clientId === permit.clientId)
        ) {
          const html = permit.html.replace(
            /<head([^>]*)>/i,
            `<head$1><script src="${BOOTSTRAP}"></script>`,
          );
          const headers = new Headers({
            "Content-Type": "text/html; charset=utf-8",
            "Cache-Control": "no-store",
          });
          if (permit.contentSecurityPolicy)
            headers.set(
              "Content-Security-Policy",
              permit.contentSecurityPolicy,
            );
          return new Response(html, { headers });
        }
        return offlinePage();
      }
      try {
        const response = await fetchNavigation(request);
        if (response.status < 500) return response;
      } catch {
        /* The boot page asks Rust to verify this device before returning private HTML. */
      }
      return offlinePage();
    })(),
  );
});
