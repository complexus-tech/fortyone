/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { headers } from "next/headers";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { GET } from "./route";

jest.mock("next/headers", () => ({ headers: jest.fn() }));
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/api-url", () => ({ getApiUrl: jest.fn() }));

const mockedAuth = jest.mocked(auth);
const mockedHeaders = jest.mocked(headers);
const mockedGetApiUrl = jest.mocked(getApiUrl);
const originalFetch = globalThis.fetch;
const originalResponse = globalThis.Response;

class MockResponse {
  readonly body: BodyInit | null;
  readonly headers: Headers;
  readonly status: number;

  constructor(body: BodyInit | null = null, init?: ResponseInit) {
    this.body = body;
    this.headers = new Headers(init?.headers);
    this.status = init?.status ?? 200;
  }

  async arrayBuffer() {
    if (this.body instanceof Uint8Array) {
      return this.body.buffer.slice(
        this.body.byteOffset,
        this.body.byteOffset + this.body.byteLength,
      );
    }
    return Uint8Array.from(
      Buffer.from(typeof this.body === "string" ? this.body : ""),
    ).buffer;
  }

  async text() {
    if (this.body instanceof Uint8Array)
      return Buffer.from(this.body).toString();
    return typeof this.body === "string" ? this.body : "";
  }
}

const params = {
  referenceId: "8f63b26c-6397-49d7-a5e7-c3cb7f257510",
  workspaceSlug: "product-team",
};

describe("GET /api/google-drive/[workspaceSlug]/files/[referenceId]/preview", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    globalThis.Response = MockResponse as never;
    mockedAuth.mockResolvedValue({ user: { id: "user-1" } } as never);
    mockedHeaders.mockResolvedValue(
      new Headers({ cookie: "session=browser-session" }) as never,
    );
    mockedGetApiUrl.mockReturnValue("https://api.example.com");
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    globalThis.Response = originalResponse;
  });

  it("authenticates before contacting the API", async () => {
    mockedAuth.mockResolvedValue(null);
    const upstreamFetch = jest.fn();
    globalThis.fetch = upstreamFetch as unknown as typeof fetch;

    const response = await GET({} as Request, {
      params: Promise.resolve(params),
    });

    expect(response.status).toBe(401);
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it("rejects overlong route identifiers before contacting the API", async () => {
    const upstreamFetch = jest.fn();
    globalThis.fetch = upstreamFetch as unknown as typeof fetch;

    const response = await GET({} as Request, {
      params: Promise.resolve({
        ...params,
        workspaceSlug: "a".repeat(256),
      }),
    });

    expect(response.status).toBe(400);
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it("rejects path syntax in a workspace slug", async () => {
    const upstreamFetch = jest.fn();
    globalThis.fetch = upstreamFetch as unknown as typeof fetch;

    const response = await GET({} as Request, {
      params: Promise.resolve({ ...params, workspaceSlug: "../admin" }),
    });

    expect(response.status).toBe(400);
    expect(upstreamFetch).not.toHaveBeenCalled();
  });

  it("encodes legacy workspace slugs instead of applying a narrower contract", async () => {
    const upstreamFetch = jest.fn(async () => ({
      body: Uint8Array.from([0x89, 0x50, 0x4e, 0x47]),
      headers: new Headers({ "Content-Type": "image/png" }),
      ok: true,
      status: 200,
    }));
    globalThis.fetch = upstreamFetch as unknown as typeof fetch;

    const response = await GET({} as Request, {
      params: Promise.resolve({ ...params, workspaceSlug: "team--one" }),
    });

    expect(response.status).toBe(200);
    expect(upstreamFetch).toHaveBeenCalledWith(
      new URL(
        "https://api.example.com/workspaces/team--one/google-drive/files/8f63b26c-6397-49d7-a5e7-c3cb7f257510/preview",
      ),
      expect.anything(),
    );
  });

  it("streams only authenticated image responses with private headers", async () => {
    const bytes = Uint8Array.from([0x89, 0x50, 0x4e, 0x47]);
    const upstreamFetch = jest.fn(async () => ({
      body: bytes,
      headers: new Headers({ "Content-Type": "image/png" }),
      ok: true,
      status: 200,
    }));
    globalThis.fetch = upstreamFetch as unknown as typeof fetch;

    const response = await GET({} as Request, {
      params: Promise.resolve(params),
    });

    expect(upstreamFetch).toHaveBeenCalledWith(
      new URL(
        "https://api.example.com/workspaces/product-team/google-drive/files/8f63b26c-6397-49d7-a5e7-c3cb7f257510/preview",
      ),
      {
        cache: "no-store",
        credentials: "include",
        headers: { cookie: "session=browser-session" },
      },
    );
    expect(response.status).toBe(200);
    expect(response.headers.get("cache-control")).toBe("private, no-store");
    expect(response.headers.get("content-type")).toBe("image/png");
    expect(response.headers.get("content-security-policy")).toContain(
      "sandbox",
    );
    expect(response.headers.get("cross-origin-resource-policy")).toBe(
      "same-origin",
    );
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(bytes);
  });

  it("does not forward non-image upstream bodies", async () => {
    globalThis.fetch = jest.fn(async () => ({
      body: "provider error",
      headers: new Headers({ "Content-Type": "text/html" }),
      ok: true,
      status: 200,
    })) as unknown as typeof fetch;

    const response = await GET({} as Request, {
      params: Promise.resolve(params),
    });

    expect(response.status).toBe(502);
    expect(await response.text()).toBe("");
  });
});
