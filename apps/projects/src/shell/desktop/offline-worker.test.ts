/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest environment pragma.
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { runInNewContext } from "node:vm";

const ORIGIN = "http://localhost:3000";
const scope = {
  accountId: "10000000-0000-4000-8000-000000000001",
  workspaceId: "10000000-0000-4000-8000-000000000002",
  workspaceSlug: "acme",
  expiresAt: Date.now() + 60_000,
};
const NONCE = "10000000-0000-4000-8000-000000000003";
const HTML =
  '<html><head><script src="/_next/static/app.js"></script></head><body>Private workspace</body></html>';

it("marks native-restored HTML offline before hydration and removes only its one-use nonce", () => {
  const restoredWindow: { __FORTYONE_OFFLINE_PAGE__?: boolean } = {};
  const replaceState = jest.fn();
  runInNewContext(
    readFileSync(resolve(process.cwd(), "public/desktop-offline.js"), "utf8"),
    {
      window: restoredWindow,
      URL,
      location: {
        href: `${ORIGIN}/acme/my-work?tab=assigned&__fortyone_offline=${NONCE}`,
      },
      history: { state: { next: true }, replaceState },
      document: { documentElement: { hasAttribute: () => false } },
    },
  );
  expect(restoredWindow.__FORTYONE_OFFLINE_PAGE__).toBe(true);
  expect(replaceState).toHaveBeenCalledWith(
    { next: true },
    "",
    `${ORIGIN}/acme/my-work?tab=assigned`,
  );
});

type EventHandler = (event: Record<string, unknown>) => void;
function harness() {
  const handlers = new Map<string, EventHandler>();
  const entries = new Map<string, Map<string, Response>>();
  const normalize = (key: string | Request) =>
    new URL(typeof key === "string" ? key : key.url, ORIGIN).href;
  const caches = {
    keys: async () => [...entries.keys()],
    delete: async (name: string) => entries.delete(name),
    open: async (name: string) => {
      if (!entries.has(name)) entries.set(name, new Map());
      const cache = entries.get(name)!;
      return {
        match: async (key: string | Request) =>
          cache.get(normalize(key))?.clone(),
        matchAll: async () => [...cache.values()].map((value) => value.clone()),
        put: async (key: string | Request, response: Response) => {
          cache.set(normalize(key), response);
        },
      };
    },
  };
  const fetch = jest.fn<Promise<Response>, [string | Request, RequestInit?]>(
    async (input) => {
      if (normalize(input).includes("/_next/static/")) {
        return new Response("window.booted=true", {
          headers: { "content-type": "text/javascript" },
        });
      }
      return new Response(HTML, {
        headers: {
          "content-type": "text/html",
          "content-security-policy": "frame-ancestors 'none'",
        },
      });
    },
  );
  runInNewContext(
    readFileSync(resolve(process.cwd(), "public/desktop-worker.js"), "utf8"),
    {
      self: {
        location: { origin: ORIGIN },
        clients: { claim: async () => undefined },
        skipWaiting: async () => undefined,
        addEventListener: (type: string, handler: EventHandler) =>
          handlers.set(type, handler),
      },
      URL,
      Response,
      Headers,
      Request,
      TextEncoder,
      TextDecoder,
      AbortController,
      setTimeout,
      clearTimeout,
      caches,
      fetch,
    },
  );
  const message = async (data: unknown) => {
    let task: Promise<void> | undefined;
    let reply:
      | { ok: boolean; error?: string; pages?: { url: string; html: string }[] }
      | undefined;
    handlers.get("message")!({
      data,
      source: {
        id: "window-1",
        type: "window",
        frameType: "top-level",
        url: `${ORIGIN}/acme/my-work`,
      },
      ports: [
        {
          postMessage: (value: typeof reply) => {
            reply = value;
          },
        },
      ],
      waitUntil: (promise: Promise<void>) => {
        task = promise;
      },
    });
    await task;
    return reply!;
  };
  const navigate = async (path: string, clientId = "window-1") => {
    let response: Promise<Response> | undefined;
    handlers.get("fetch")!({
      request: {
        method: "GET",
        mode: "navigate",
        url: new URL(path, ORIGIN).href,
      },
      clientId,
      respondWith: (promise: Promise<Response>) => {
        response = promise;
      },
    });
    return response!;
  };
  return { entries, fetch, message, navigate };
}

test("downloads real HTML but persists only public static assets", async () => {
  const app = harness();
  const result = await app.message({
    type: "warm",
    scope,
    urls: ["/acme/my-work"],
  });
  expect(result.ok).toBe(true);
  expect(result.pages?.[0].html).toBe(HTML);
  expect(
    [...app.entries.values()].flatMap((cache) => [...cache.keys()]),
  ).toEqual([`${ORIGIN}/_next/static/app.js`]);
  expect(app.fetch).toHaveBeenCalledTimes(2);
});

test("offline navigation never returns private HTML without native-authorized handoff", async () => {
  const app = harness();
  app.fetch.mockRejectedValue(new Error("offline"));
  const page = await app.navigate("/acme/my-work");
  expect(await page.text()).toContain("data-fortyone-offline-bootstrap");
  expect(
    (
      await app.message({
        type: "restore",
        scope,
        nonce: NONCE,
        url: `${ORIGIN}/acme/my-work`,
      })
    ).ok,
  ).toBe(false);
});

test("an unresponsive navigation aborts after eight seconds and opens the offline bootstrap", async () => {
  jest.useFakeTimers();
  try {
    const app = harness();
    let signal: AbortSignal | null | undefined;
    app.fetch.mockImplementationOnce(
      (_request, options) =>
        new Promise<Response>((_resolve, reject) => {
          signal = options?.signal;
          signal?.addEventListener(
            "abort",
            () => {
              reject(new Error("aborted"));
            },
            { once: true },
          );
        }),
    );
    const settled = jest.fn();
    const navigation = app.navigate("/acme/my-work").then((response) => {
      settled();
      return response;
    });
    await jest.advanceTimersByTimeAsync(7_999);
    expect(settled).not.toHaveBeenCalled();
    expect(signal?.aborted).toBe(false);
    await jest.advanceTimersByTimeAsync(1);
    expect(signal?.aborted).toBe(true);
    expect(await (await navigation).text()).toContain(
      "data-fortyone-offline-bootstrap",
    );
    expect(jest.getTimerCount()).toBe(0);
    expect(app.entries.size).toBe(0);
    expect(Object.keys(app.fetch.mock.calls[0][1]!)).toEqual(["signal"]);
  } finally {
    jest.useRealTimers();
  }
});

test("a successful navigation cancels its deadline without changing the response", async () => {
  jest.useFakeTimers();
  try {
    const app = harness();
    const network = new Response(HTML, {
      headers: { "content-type": "text/html" },
    });
    app.fetch.mockResolvedValueOnce(network);
    expect(await app.navigate("/acme/my-work")).toBe(network);
    const signal = app.fetch.mock.calls[0][1]?.signal;
    expect(jest.getTimerCount()).toBe(0);
    await jest.advanceTimersByTimeAsync(8_000);
    expect(signal?.aborted).toBe(false);
  } finally {
    jest.useRealTimers();
  }
});

test("handoff is one-use, URL bound and preserves CSP", async () => {
  const app = harness();
  await app.message({
    type: "restore",
    scope,
    nonce: NONCE,
    url: `${ORIGIN}/acme/my-work`,
    html: HTML,
    contentSecurityPolicy: "frame-ancestors 'none'",
  });
  const first = await app.navigate(`/acme/my-work?__fortyone_offline=${NONCE}`);
  expect(first.headers.get("content-security-policy")).toBe(
    "frame-ancestors 'none'",
  );
  expect(await first.text()).toContain("Private workspace");
  const repeated = await app.navigate(
    `/acme/my-work?__fortyone_offline=${NONCE}`,
  );
  expect(await repeated.text()).toContain("data-fortyone-offline-bootstrap");
});

test.each([
  "/other/my-work",
  "/acme/auth/callback",
  "/acme/my-work?token=secret",
])("rejects unsafe private page destination %s", async (url) => {
  const app = harness();
  const result = await app.message({ type: "warm", scope, urls: [url] });
  expect(result.ok).toBe(false);
  expect(app.fetch).not.toHaveBeenCalled();
});

test("rejects expired scope and clears in-memory handoffs on logout", async () => {
  const app = harness();
  expect(
    (
      await app.message({
        type: "restore",
        scope: { ...scope, expiresAt: 0 },
        nonce: NONCE,
        url: `${ORIGIN}/acme/my-work`,
        html: HTML,
      })
    ).ok,
  ).toBe(false);
  await app.message({
    type: "restore",
    scope,
    nonce: NONCE,
    url: `${ORIGIN}/acme/my-work`,
    html: HTML,
  });
  await app.message({ type: "clear" });
  expect(
    await (
      await app.navigate(`/acme/my-work?__fortyone_offline=${NONCE}`)
    ).text(),
  ).toContain("data-fortyone-offline-bootstrap");
});

test("logout during download prevents the private result being handed back for persistence", async () => {
  const app = harness();
  let complete: ((response: Response) => void) | undefined;
  app.fetch.mockImplementationOnce(
    () =>
      new Promise<Response>((resolve) => {
        complete = resolve;
      }),
  );
  const pending = app.message({ type: "warm", scope, urls: ["/acme/my-work"] });
  await app.message({ type: "clear" });
  complete!(new Response(HTML, { headers: { "content-type": "text/html" } }));
  expect((await pending).ok).toBe(false);
});
