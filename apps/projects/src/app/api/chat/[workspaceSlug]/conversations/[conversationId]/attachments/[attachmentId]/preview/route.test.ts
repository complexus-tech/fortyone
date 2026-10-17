/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.
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
const params = {
  attachmentId: "8f63b26c-6397-49d7-a5e7-c3cb7f257510",
  conversationId: "c803668d-75c0-41de-a31c-6a9afc0f70bf",
  workspaceSlug: "product-team",
};
const signedURL =
  "https://storage.example.test/private.pdf?signature=temporary";
const request = (init?: RequestInit) =>
  new Request("https://app.example.test/api/chat/preview", init);
const audioRequest = (init?: RequestInit) =>
  new Request("https://app.example.test/api/chat/preview?media=audio", init);
const context = () => ({ params: Promise.resolve(params) });
const authorizationResponse = (location = signedURL) =>
  new Response(null, { status: 307, headers: { Location: location } });
const pdfResponse = (body: BodyInit = "%PDF-1.7", init?: ResponseInit) =>
  new Response(body, {
    ...init,
    headers: { "Content-Type": "application/pdf", ...init?.headers },
  });

describe("private chat attachment preview", () => {
  let fetchMock: jest.SpyInstance<ReturnType<typeof fetch>>;

  beforeEach(() => {
    jest.clearAllMocks();
    fetchMock = jest.spyOn(globalThis, "fetch");
    mockedAuth.mockResolvedValue({ user: { id: "user-1" } } as never);
    mockedHeaders.mockResolvedValue(
      new Headers({ cookie: "fortyone_session=browser-session" }) as never,
    );
    mockedGetApiUrl.mockReturnValue("https://api.example.test");
  });

  afterEach(() => {
    fetchMock.mockRestore();
  });

  it("authenticates before validating identifiers or contacting an upstream", async () => {
    mockedAuth.mockResolvedValue(null);
    const response = await GET(request(), {
      params: Promise.resolve({ ...params, attachmentId: "invalid" }),
    });

    expect(response.status).toBe(401);
    expect(response.headers.get("cache-control")).toBe("private, no-store");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it.each([
    { attachmentId: "../secret" },
    { conversationId: "not-a-uuid" },
    { workspaceSlug: "../other-workspace" },
    { workspaceSlug: "a".repeat(256) },
  ])("rejects malformed route scope %j", async (invalid) => {
    const response = await GET(request(), {
      params: Promise.resolve({ ...params, ...invalid }),
    });

    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it.each([
    "bytes=0-10,20-30",
    "bytes=10-2",
    "bytes=-0",
    "bytes=-",
    "items=0-2",
  ])("rejects invalid or multiple ranges: %s", async (range) => {
    const response = await GET(
      request({ headers: { Range: range } }),
      context(),
    );
    expect(response.status).toBe(416);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("authorizes the exact conversation and streams without leaking cookies or signed URLs", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        pdfResponse("%PDF-1.7", {
          headers: { "Content-Length": "8", "Accept-Ranges": "bytes" },
        }),
      );
    const response = await GET(request(), context());

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      new URL(
        `https://api.example.test/workspaces/${params.workspaceSlug}/chat/conversations/${params.conversationId}/attachments/${params.attachmentId}`,
      ),
      expect.objectContaining({
        cache: "no-store",
        headers: { cookie: "fortyone_session=browser-session" },
        redirect: "manual",
      }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      new URL(signedURL),
      expect.objectContaining({
        cache: "no-store",
        credentials: "omit",
        headers: { "Accept-Encoding": "identity" },
        redirect: "manual",
      }),
    );
    expect(response.status).toBe(200);
    expect(response.headers.get("location")).toBeNull();
    expect(response.headers.get("cache-control")).toBe("private, no-store");
    expect(response.headers.get("content-type")).toBe("application/pdf");
    expect(response.headers.get("referrer-policy")).toBe("no-referrer");
    expect(response.headers.get("x-content-type-options")).toBe("nosniff");
    expect(response.headers.get("accept-ranges")).toBe("bytes");
    expect(await response.text()).toBe("%PDF-1.7");
  });

  it.each([
    ["bytes=2-4", "bytes 2-4/8", "PDF"],
    ["bytes=5-", "bytes 5-7/8", "1.7"],
    ["bytes=-3", "bytes 5-7/8", "1.7"],
  ])(
    "preserves valid partial responses for %s",
    async (range, returnedRange, bytes) => {
      fetchMock
        .mockResolvedValueOnce(authorizationResponse())
        .mockResolvedValueOnce(
          pdfResponse(bytes, {
            status: 206,
            headers: { "Content-Range": returnedRange, "Content-Length": "3" },
          }),
        );
      const response = await GET(
        request({ headers: { Range: range } }),
        context(),
      );

      expect(fetchMock).toHaveBeenNthCalledWith(
        2,
        new URL(signedURL),
        expect.objectContaining({
          headers: { "Accept-Encoding": "identity", Range: range },
        }),
      );
      expect(response.status).toBe(206);
      expect(response.headers.get("content-range")).toBe(returnedRange);
      expect(response.headers.get("content-length")).toBe("3");
      expect(await response.text()).toBe(bytes);
    },
  );

  it("rejects mismatched range data instead of corrupting the viewer's bytes", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        pdfResponse("PDF", {
          status: 206,
          headers: { "Content-Range": "bytes 0-2/8", "Content-Length": "3" },
        }),
      );
    const response = await GET(
      request({ headers: { Range: "bytes=2-4" } }),
      context(),
    );
    expect(response.status).toBe(502);
    expect(await response.text()).toBe("");
  });

  it.each([401, 403, 404])(
    "preserves authorization/deletion status %i without exposing an upstream body",
    async (status) => {
      fetchMock.mockResolvedValueOnce(
        new Response("private upstream detail", { status }),
      );
      const response = await GET(request(), context());

      expect(response.status).toBe(status);
      expect(await response.text()).toBe("");
      expect(fetchMock).toHaveBeenCalledTimes(1);
    },
  );

  it.each([
    "http://remote-storage.example.test/file.pdf",
    "file:///etc/private",
    "https://user:password@storage.example.test/file.pdf",
    "/relative.pdf",
  ])("does not fetch an unsafe storage target %s", async (location) => {
    fetchMock.mockResolvedValueOnce(authorizationResponse(location));
    const response = await GET(request(), context());
    expect(response.status).toBe(502);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("does not follow a further storage redirect", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response(null, {
          status: 307,
          headers: { Location: "https://other.example.test/private" },
        }),
      );
    const response = await GET(request(), context());
    expect(response.status).toBe(502);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it.each<Record<string, string>>([
    { "Content-Type": "text/html" },
    {
      "Content-Type": "application/pdf",
      "Content-Length": String(25 * 1024 * 1024 + 1),
    },
    { "Content-Type": "application/pdf", "Content-Encoding": "gzip" },
  ])("rejects unexpected file metadata %j", async (metadata) => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response("private body", { headers: metadata }),
      );
    const response = await GET(request(), context());
    expect(response.status).toBe(502);
    expect(await response.text()).toBe("");
  });

  it("returns safe unsatisfiable range metadata", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response("private error", {
          status: 416,
          headers: { "Content-Range": "bytes */8" },
        }),
      );
    const response = await GET(
      request({ headers: { Range: "bytes=9-" } }),
      context(),
    );
    expect(response.status).toBe(416);
    expect(response.headers.get("content-range")).toBe("bytes */8");
    expect(await response.text()).toBe("");
  });

  it("enforces the byte cap while streaming a response without Content-Length", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(pdfResponse(new Uint8Array(25 * 1024 * 1024 + 1)));
    const response = await GET(request(), context());
    expect(response.status).toBe(200);
    await expect(response.arrayBuffer()).rejects.toThrow("allowed size");
  });

  it("detects a truncated file instead of completing an invalid PDF response", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        pdfResponse("short", { headers: { "Content-Length": "20" } }),
      );
    const response = await GET(request(), context());
    await expect(response.text()).rejects.toThrow(
      "before the file was complete",
    );
  });

  it("propagates viewer cancellation to the storage stream without buffering it", async () => {
    const cancelled = jest.fn();
    const storageBody = new ReadableStream<Uint8Array>({ cancel: cancelled });
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(pdfResponse(storageBody));
    const response = await GET(request(), context());
    await response.body?.cancel("viewer closed");
    expect(cancelled).toHaveBeenCalledWith("viewer closed");
  });

  it("propagates request aborts during streaming", async () => {
    const aborted = new AbortController();
    const cancelled = jest.fn();
    const storageBody = new ReadableStream<Uint8Array>({ cancel: cancelled });
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(pdfResponse(storageBody));
    const response = await GET(request({ signal: aborted.signal }), context());
    const reading = response.text();
    aborted.abort();
    await expect(reading).rejects.toThrow();
    expect(cancelled).toHaveBeenCalled();
  });

  it("does not contact storage when the viewer was already closed", async () => {
    const aborted = new AbortController();
    aborted.abort();
    const response = await GET(request({ signal: aborted.signal }), context());
    expect(response.status).toBe(499);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("sanitizes network failures", async () => {
    fetchMock.mockRejectedValue(new Error(`failed to fetch ${signedURL}`));
    const response = await GET(request(), context());
    expect(response.status).toBe(502);
    expect(await response.text()).toBe("");
  });

  it.each(["audio/mp4", "audio/webm; codecs=opus"])(
    "authorizes and privately streams explicit audio previews for %s",
    async (mediaType) => {
      fetchMock
        .mockResolvedValueOnce(authorizationResponse())
        .mockResolvedValueOnce(
          new Response("DATA", {
            headers: { "Content-Type": mediaType, "Content-Length": "4" },
          }),
        );
      const response = await GET(audioRequest(), context());
      expect(response.status).toBe(200);
      expect(response.headers.get("content-type")).toBe(
        mediaType.split(";")[0],
      );
      expect(response.headers.get("cache-control")).toBe("private, no-store");
      expect(response.headers.get("cross-origin-resource-policy")).toBe(
        "same-origin",
      );
      expect(response.headers.get("content-disposition")).toBe(
        `inline; filename="recording.${mediaType === "audio/mp4" ? "m4a" : "webm"}"`,
      );
      expect(response.headers.get("location")).toBeNull();
      expect(fetchMock).toHaveBeenNthCalledWith(
        1,
        expect.any(URL),
        expect.objectContaining({
          headers: { cookie: "fortyone_session=browser-session" },
          cache: "no-store",
          redirect: "manual",
        }),
      );
      expect(fetchMock).toHaveBeenNthCalledWith(
        2,
        new URL(signedURL),
        expect.objectContaining({
          credentials: "omit",
          headers: { "Accept-Encoding": "identity" },
          cache: "no-store",
          redirect: "manual",
        }),
      );
      expect(await response.text()).toBe("DATA");
    },
  );

  it("preserves bounded audio range responses", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response("ATA", {
          status: 206,
          headers: {
            "Content-Type": "audio/mp4",
            "Content-Length": "3",
            "Content-Range": "bytes 1-3/4",
          },
        }),
      );
    const response = await GET(
      audioRequest({ headers: { Range: "bytes=1-3" } }),
      context(),
    );
    expect(response.status).toBe(206);
    expect(response.headers.get("content-range")).toBe("bytes 1-3/4");
    expect(response.headers.get("content-length")).toBe("3");
    expect(await response.text()).toBe("ATA");
  });

  it.each(["application/pdf", "video/webm", "audio/mpeg", "text/html"])(
    "rejects a mismatched audio preview type: %s",
    async (mediaType) => {
      fetchMock
        .mockResolvedValueOnce(authorizationResponse())
        .mockResolvedValueOnce(
          new Response("private data", {
            headers: { "Content-Type": mediaType },
          }),
        );
      const response = await GET(audioRequest(), context());
      expect(response.status).toBe(502);
      expect(await response.text()).toBe("");
    },
  );

  it("keeps requests without the explicit audio option PDF-only", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response("DATA", { headers: { "Content-Type": "audio/mp4" } }),
      );
    expect((await GET(request(), context())).status).toBe(502);
  });

  it("rejects unsupported preview modes before contacting the API", async () => {
    const response = await GET(
      new Request("https://app.example.test/api/chat/preview?media=video"),
      context(),
    );
    expect(response.status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it.each([200, 206])(
    "applies the 20 MiB audio limit to %i metadata",
    async (status) => {
      const length = String(20 * 1024 * 1024 + 1);
      fetchMock
        .mockResolvedValueOnce(authorizationResponse())
        .mockResolvedValueOnce(
          new Response("DATA", {
            status,
            headers: {
              "Content-Type": "audio/mp4",
              "Content-Length": status === 206 ? "4" : length,
              ...(status === 206
                ? { "Content-Range": `bytes 0-3/${length}` }
                : {}),
            },
          }),
        );
      const response = await GET(
        audioRequest(
          status === 206 ? { headers: { Range: "bytes=0-3" } } : undefined,
        ),
        context(),
      );
      expect(response.status).toBe(502);
      expect(await response.text()).toBe("");
    },
  );

  it("enforces the audio size cap on a stream without length metadata", async () => {
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response(new Uint8Array(20 * 1024 * 1024 + 1), {
          headers: { "Content-Type": "audio/webm" },
        }),
      );
    const response = await GET(audioRequest(), context());
    expect(response.status).toBe(200);
    await expect(response.arrayBuffer()).rejects.toThrow("allowed size");
  });

  it.each([401, 403, 404])(
    "retains audio authorization failures: %i",
    async (status) => {
      fetchMock.mockResolvedValueOnce(
        new Response("private details", { status }),
      );
      const response = await GET(audioRequest(), context());
      expect(response.status).toBe(status);
      expect(await response.text()).toBe("");
      expect(fetchMock).toHaveBeenCalledTimes(1);
    },
  );

  it("requires an authenticated app session for audio previews", async () => {
    mockedAuth.mockResolvedValue(null);
    expect((await GET(audioRequest(), context())).status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("cancels the upstream audio stream when the waveform request aborts", async () => {
    const controller = new AbortController();
    const cancelled = jest.fn();
    fetchMock
      .mockResolvedValueOnce(authorizationResponse())
      .mockResolvedValueOnce(
        new Response(new ReadableStream<Uint8Array>({ cancel: cancelled }), {
          headers: { "Content-Type": "audio/mp4" },
        }),
      );
    const response = await GET(
      audioRequest({ signal: controller.signal }),
      context(),
    );
    const reading = response.arrayBuffer();
    controller.abort();
    await expect(reading).rejects.toThrow();
    expect(cancelled).toHaveBeenCalled();
  });
});
