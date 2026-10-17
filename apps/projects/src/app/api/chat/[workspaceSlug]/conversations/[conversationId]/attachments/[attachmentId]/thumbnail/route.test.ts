/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest environment directive.
/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals. */
import { auth } from "@/auth";
import { GET } from "./route";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("next/headers", () => ({
  headers: async () => new Headers({ cookie: "session=private" }),
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));
const params = {
  workspaceSlug: "acme",
  conversationId: "c803668d-75c0-41de-a31c-6a9afc0f70bf",
  attachmentId: "8f63b26c-6397-49d7-a5e7-c3cb7f257510",
};
const context = () => ({ params: Promise.resolve(params) });
const etag = `"${"a".repeat(64)}"`;
const request = (headers?: HeadersInit) =>
  new Request("https://app.example.test/thumbnail", { headers });
const access = (
  location = "https://storage.example.test/thumb.jpg?token=secret",
) =>
  new Response(null, {
    status: 307,
    headers: { Location: location, ETag: etag },
  });
let fetchMock: jest.SpyInstance;
beforeEach(() => {
  jest.mocked(auth).mockResolvedValue({ user: { id: "alice" } } as never);
  fetchMock = jest.spyOn(globalThis, "fetch");
});
afterEach(() => {
  jest.restoreAllMocks();
});

describe("authorized thumbnail proxy", () => {
  it("checks the session before any upstream call", async () => {
    jest.mocked(auth).mockResolvedValue(null);
    expect((await GET(request(), context())).status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it.each([401, 403, 404])(
    "does not reuse pixels after API authorization returns %i",
    async (status) => {
      fetchMock.mockResolvedValueOnce(new Response(null, { status }));
      expect(
        (await GET(request({ "If-None-Match": etag }), context())).status,
      ).toBe(status);
      expect(fetchMock).toHaveBeenCalledTimes(1);
    },
  );
  it("forwards a conditional request only to the authorized API", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(null, { status: 304, headers: { ETag: etag } }),
    );
    const result = await GET(request({ "If-None-Match": etag }), context());
    expect(result.status).toBe(304);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(URL),
      expect.objectContaining({
        headers: { cookie: "session=private", "If-None-Match": etag },
        redirect: "manual",
      }),
    );
  });
  it("returns worker pending state without storage access", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 202 }));
    const result = await GET(request(), context());
    expect(result.status).toBe(202);
    expect(result.headers.get("retry-after")).toBe("2");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("returns bounded private JPEG bytes and never forwards application credentials to storage", async () => {
    fetchMock
      .mockResolvedValueOnce(access())
      .mockResolvedValueOnce(
        new Response(new Uint8Array([255, 216, 255, 217]), {
          headers: { "Content-Type": "image/jpeg" },
        }),
      );
    const result = await GET(request(), context());
    expect(result.status).toBe(200);
    expect(result.headers.get("etag")).toBe(etag);
    expect(result.headers.get("cache-control")).toBe("private, no-store");
    expect(result.headers.has("location")).toBe(false);
    expect((await result.arrayBuffer()).byteLength).toBe(4);
    expect(fetchMock.mock.calls[1][1]).toEqual(
      expect.objectContaining({
        credentials: "omit",
        headers: { "Accept-Encoding": "identity" },
        redirect: "manual",
      }),
    );
  });
  it.each([
    "https://user:password@example.test/image.jpg",
    "http://remote.example.test/image.jpg",
    "file:///private/file",
  ])("rejects unsafe storage target %s", async (target) => {
    fetchMock.mockResolvedValueOnce(access(target));
    expect((await GET(request(), context())).status).toBe(502);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
  it("refuses oversized storage responses before streaming", async () => {
    fetchMock
      .mockResolvedValueOnce(access())
      .mockResolvedValueOnce(
        new Response("x", {
          headers: {
            "Content-Type": "image/jpeg",
            "Content-Length": String(2 * 1024 * 1024 + 1),
          },
        }),
      );
    expect((await GET(request(), context())).status).toBe(502);
  });
});
