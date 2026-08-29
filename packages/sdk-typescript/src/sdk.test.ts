import assert from "node:assert/strict";
import { test } from "node:test";

import {
  apiErrorFromResponse,
  createIdempotencyKey,
  createFortyOneClient,
  createRetryingFetch,
  FortyOneApiError,
  paginateStories,
  paginateStoryPages,
  verifyWebhook,
  validateIdempotencyKey,
  WebhookVerificationError,
} from "./index.js";

test("idempotency helpers generate valid independent operation keys", () => {
  let seed = 0;
  const cryptoProvider = {
    getRandomValues<T extends ArrayBufferView | null>(array: T): T {
      if (!(array instanceof Uint8Array)) {
        throw new Error("unexpected random array");
      }
      for (let index = 0; index < array.length; index += 1) {
        array[index] = (seed + index) % 256;
      }
      seed += 1;
      return array as T;
    },
  };
  const first = createIdempotencyKey(cryptoProvider);
  const second = createIdempotencyKey(cryptoProvider);
  assert.equal(first.length, 64);
  assert.notEqual(first, second);
  assert.doesNotThrow(() => validateIdempotencyKey(first));
  assert.throws(() => validateIdempotencyKey("too short"), /16 to 255/u);
});

const WORKSPACE_ID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const STORY_ID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
const WEBHOOK_ID = "11111111-1111-4111-8111-111111111111";
const WEBHOOK_TIMESTAMP = "1787920000";
const WEBHOOK_SECRET =
  "whsec_UlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlI=";
const WEBHOOK_SIGNATURE =
  "v1,fmCSZfMbiTh50bHta4Wg4YSYK94JW+6U9d0vvQHQQbk=";
const WEBHOOK_BODY = new TextEncoder().encode(
  '{"value":"line one\\r\\nline two"}',
);

test("client applies bearer authentication without changing the contract", async () => {
  let request: Request | undefined;
  const client = createFortyOneClient({
    token: "pat_test_token",
    baseUrl: "http://127.0.0.1:8080",
    allowInsecureLoopback: true,
    retry: false,
    fetch: async (input) => {
      request = input instanceof Request ? input : new Request(input);
      return Response.json({ data: story(STORY_ID) });
    },
  });

  const result = await client.GET(
    "/api/v1/workspaces/{workspaceId}/stories/{storyId}",
    { params: { path: { workspaceId: WORKSPACE_ID, storyId: STORY_ID } } },
  );

  assert.equal(result.data?.data.id, STORY_ID);
  assert.equal(request?.headers.get("Authorization"), "Bearer pat_test_token");
  assert.equal(request?.headers.get("Accept"), "application/json");
  assert.equal(
    request?.url,
    `http://127.0.0.1:8080/api/v1/workspaces/${WORKSPACE_ID}/stories/${STORY_ID}`,
  );
});

test("generated client exposes typed collaboration reads", async () => {
  const teamId = "dddddddd-dddd-4ddd-8ddd-dddddddddddd";
  let request: Request | undefined;
  const client = createFortyOneClient({
    token: "pat_test_token",
    baseUrl: "http://127.0.0.1:8080",
    allowInsecureLoopback: true,
    retry: false,
    fetch: async (input) => {
      request = input instanceof Request ? input : new Request(input);
      return Response.json({
        data: [
          {
            id: "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee",
            workspaceId: WORKSPACE_ID,
            teamId,
            name: "API",
            color: "#123456",
            createdAt: "2026-08-28T12:00:00Z",
            updatedAt: "2026-08-28T12:00:00Z",
          },
        ],
        meta: { hasMore: false },
      });
    },
  });

  const result = await client.GET("/api/v1/workspaces/{workspaceId}/labels", {
    params: {
      path: { workspaceId: WORKSPACE_ID },
      query: { teamId, limit: 25 },
    },
  });

  assert.equal(result.data?.data[0]?.name, "API");
  const url = new URL(request?.url ?? "");
  assert.equal(url.pathname, `/api/v1/workspaces/${WORKSPACE_ID}/labels`);
  assert.equal(url.searchParams.get("teamId"), teamId);
  assert.equal(url.searchParams.get("limit"), "25");
});

test("client rejects unsafe base URLs and malformed credentials", () => {
  assert.throws(
    () => createFortyOneClient({ token: "pat_test", baseUrl: "http://api.example.com" }),
    /must use HTTPS/u,
  );
  assert.throws(
    () => createFortyOneClient({ token: "pat test" }),
    /missing or malformed/u,
  );
});

test("retry policy retries only safe reads for explicit transient statuses", async () => {
  let calls = 0;
  const retryingFetch = createRetryingFetch(
    async () => {
      calls += 1;
      return calls === 1
        ? new Response(null, { status: 503 })
        : new Response(null, { status: 204 });
    },
    { maxAttempts: 2, baseDelayMs: 1, maxDelayMs: 1 },
  );

  const response = await retryingFetch("https://api.fortyone.app/health");
  assert.equal(response.status, 204);
  assert.equal(calls, 2);

  calls = 0;
  await retryingFetch("https://api.fortyone.app/webhook", { method: "POST" });
  assert.equal(calls, 1);
});

test("API errors expose safe structured fields and request metadata", () => {
  const error = apiErrorFromResponse(
    new Response(null, {
      status: 429,
      headers: { "Retry-After": "8", "X-Request-ID": "req_header" },
    }),
    {
      error: {
        code: "rate_limited",
        message: "Try again later",
        requestId: "req_body",
        fields: [{ field: "limit", code: "exhausted", message: "Too many" }],
      },
    },
  );

  assert.ok(error instanceof FortyOneApiError);
  assert.equal(error.status, 429);
  assert.equal(error.code, "rate_limited");
  assert.equal(error.requestId, "req_body");
  assert.equal(error.retryAfterSeconds, 8);
  assert.equal(error.fields[0]?.field, "limit");
});

test("story pagination follows opaque cursors and detects repeated cursors", async () => {
  const requestedCursors: Array<string | null> = [];
  let call = 0;
  const client = createFortyOneClient({
    token: "pat_test_token",
    baseUrl: "http://localhost:8080",
    allowInsecureLoopback: true,
    retry: false,
    fetch: async (input) => {
      const request = input instanceof Request ? input : new Request(input);
      requestedCursors.push(new URL(request.url).searchParams.get("cursor"));
      call += 1;
      return Response.json({
        data: [story(call === 1 ? STORY_ID : "cccccccc-cccc-4ccc-8ccc-cccccccccccc")],
        meta:
          call === 1
            ? { hasMore: true, nextCursor: "opaque-next" }
            : { hasMore: false },
      });
    },
  });

  const ids: string[] = [];
  for await (const item of paginateStories(client, WORKSPACE_ID, { limit: 1 })) {
    ids.push(item.id);
  }

  assert.deepEqual(ids, [STORY_ID, "cccccccc-cccc-4ccc-8ccc-cccccccccccc"]);
  assert.deepEqual(requestedCursors, [null, "opaque-next"]);
});

test("story pagination fails closed when the server repeats a cursor", async () => {
  const client = createFortyOneClient({
    token: "pat_test_token",
    baseUrl: "http://localhost:8080",
    allowInsecureLoopback: true,
    retry: false,
    fetch: async () =>
      Response.json({
        data: [],
        meta: { hasMore: true, nextCursor: "repeated" },
      }),
  });

  await assert.rejects(async () => {
    for await (const _page of paginateStoryPages(client, WORKSPACE_ID)) {
      // Pagination must request the second page before identifying the cycle.
    }
  }, /invalid or repeated pagination cursor/u);
});

test("webhook verification accepts the shared fixture and rotation signatures", async () => {
  const verified = await verifyWebhook({
    secret: WEBHOOK_SECRET,
    body: WEBHOOK_BODY,
    webhookId: WEBHOOK_ID,
    webhookTimestamp: WEBHOOK_TIMESTAMP,
    webhookSignature: `v1,${"A".repeat(43)}= ${WEBHOOK_SIGNATURE}`,
    now: new Date(Number(WEBHOOK_TIMESTAMP) * 1_000),
  });

  assert.equal(verified.id, WEBHOOK_ID);
  assert.equal(verified.timestamp.toISOString(), "2026-08-28T12:26:40.000Z");
});

test("webhook verification rejects body tampering and replayed deliveries", async () => {
  await assert.rejects(
    verifyWebhook({
      secret: WEBHOOK_SECRET,
      body: new TextEncoder().encode('{"value":"changed"}'),
      webhookId: WEBHOOK_ID,
      webhookTimestamp: WEBHOOK_TIMESTAMP,
      webhookSignature: WEBHOOK_SIGNATURE,
      now: new Date(Number(WEBHOOK_TIMESTAMP) * 1_000),
    }),
    (error: unknown) =>
      error instanceof WebhookVerificationError &&
      error.code === "invalid_signature",
  );
  await assert.rejects(
    verifyWebhook({
      secret: WEBHOOK_SECRET,
      body: WEBHOOK_BODY,
      webhookId: WEBHOOK_ID,
      webhookTimestamp: WEBHOOK_TIMESTAMP,
      webhookSignature: WEBHOOK_SIGNATURE,
      now: new Date((Number(WEBHOOK_TIMESTAMP) + 301) * 1_000),
    }),
    (error: unknown) =>
      error instanceof WebhookVerificationError &&
      error.code === "stale_timestamp",
  );
});

const story = (id: string) => ({
  autoSchedulingEnabled: false,
  autoSchedulingLocked: false,
  autoSchedulingStatus: "disabled",
  createdAt: "2026-08-28T12:00:00Z",
  id,
  labels: [],
  priority: "none",
  reference: "ENG-1",
  sequenceId: 1,
  teamId: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
  title: "Typed SDK",
  updatedAt: "2026-08-28T12:00:00Z",
  workspaceId: WORKSPACE_ID,
});
