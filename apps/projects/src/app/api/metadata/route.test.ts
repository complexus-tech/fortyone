/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { NextRequest } from "next/server";
import * as cheerio from "cheerio";
import { auth } from "@/auth";
import { fetchPublicHtml } from "./safe-fetch";
import { MetadataFetchError } from "./target-policy";
import { GET } from "./route";
import { metadataService } from "./metadata-service";
import { MetadataCapacityError } from "./metadata-cache";
import { allowPreviewRefresh } from "./refresh-limiter";

jest.mock("./refresh-limiter", () => ({
  allowPreviewRefresh: jest.fn().mockResolvedValue(true),
}));
jest.mock("./shared-cache", () => ({
  ...jest.requireActual("./shared-cache"),
  sharedPreviewCache: {
    read: async () => null,
    write: async () => undefined,
    claim: async () => null,
    release: async () => undefined,
  },
}));
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("cheerio", () => ({ load: jest.fn() }));
jest.mock("next/server", () => ({
  NextResponse: {
    json: (body: unknown, init?: { status?: number }) => ({
      json: async () => body,
      status: init?.status ?? 200,
    }),
  },
}));
jest.mock("./safe-fetch", () => ({
  ...jest.requireActual("./safe-fetch"),
  fetchPublicHtml: jest.fn(),
}));

const mockedAuth = jest.mocked(auth);
const mockedFetchPublicHtml = jest.mocked(fetchPublicHtml);
const mockedLoad = jest.mocked(cheerio.load);

const session = {
  user: {
    email: "member@example.com",
    fullName: "Example Member",
    id: "user-1",
    image: null,
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
    name: "Example Member",
    username: "member",
  },
};

const createRequest = (url: string) =>
  ({ signal: new AbortController().signal, url }) as NextRequest;

describe("GET /api/metadata", () => {
  beforeEach(() => {
    metadataService.clear();
    jest.clearAllMocks();
    mockedAuth.mockResolvedValue(session);
    jest.mocked(allowPreviewRefresh).mockResolvedValue(true);
  });

  it("authenticates before validating or fetching the target", async () => {
    mockedAuth.mockResolvedValue(null);

    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=http://127.0.0.1",
      ),
    );

    expect(response.status).toBe(401);
    expect(mockedFetchPublicHtml).not.toHaveBeenCalled();
  });

  it("rejects invalid target URLs", async () => {
    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=ftp://example.com/file",
      ),
    );

    expect(response.status).toBe(400);
    expect(mockedFetchPublicHtml).not.toHaveBeenCalled();
  });

  it("extracts metadata and resolves assets against the final redirect URL", async () => {
    const selectorValues = new Map([
      ['link[rel="icon"]', "assets/favicon.png"],
      ['meta[name="description"]', "Example description"],
      ['meta[property="og:image"]', "/og-image.png"],
      ['meta[property="og:title"]', "Example title"],
      ["title", "Fallback title"],
    ]);
    mockedLoad.mockReturnValue(((selector: string) => ({
      attr: () => selectorValues.get(selector),
      text: () => selectorValues.get(selector) ?? "",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://www.example.com/articles/final"),
    });

    const request = createRequest(
      "https://projects.example.com/api/metadata?url=https://example.com/start",
    );
    const response = await GET(request);

    await expect(response.json()).resolves.toMatchObject({
      description: "Example description",
      image: expect.stringMatching(
        /^\/api\/metadata\/image\?url=https%3A%2F%2Fwww.example.com%2Farticles%2Fassets%2Ffavicon.png&v=[a-f0-9-]+$/,
      ),
      previewImage: expect.stringMatching(
        /^\/api\/metadata\/image\?url=https%3A%2F%2Fwww.example.com%2Fog-image.png&v=[a-f0-9-]+$/,
      ),
      title: "Example title",
    });
    expect(response.status).toBe(200);
    expect(mockedFetchPublicHtml).toHaveBeenCalledWith(
      new URL("https://example.com/start"),
      { signal: expect.any(AbortSignal) },
    );
  });

  it("does not enlarge a favicon into a social image preview", async () => {
    mockedLoad.mockReturnValue(((selector: string) => ({
      attr: () =>
        selector === 'link[rel="icon"]' ? "/favicon.png" : undefined,
      text: () => "Example title",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://example.com"),
    });

    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com",
      ),
    );

    await expect(response.json()).resolves.toMatchObject({
      image: expect.stringMatching(
        /^\/api\/metadata\/image\?url=https%3A%2F%2Fexample.com%2Ffavicon.png&v=[a-f0-9-]+$/,
      ),
      previewImage: undefined,
    });
  });

  it("falls back to a proxied Twitter image when the OG candidate is unsupported", async () => {
    const selectorValues = new Map([
      ['meta[property="og:image"]', "data:image/png;base64,ignored"],
      ['meta[name="twitter:image"]', "../social-card.jpg"],
    ]);
    mockedLoad.mockReturnValue(((selector: string) => ({
      attr: () => selectorValues.get(selector),
      text: () => "Example title",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://example.com/articles/final"),
    });

    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com/start",
      ),
    );

    await expect(response.json()).resolves.toMatchObject({
      previewImage: expect.stringMatching(
        /^\/api\/metadata\/image\?url=https%3A%2F%2Fexample.com%2Fsocial-card.jpg&v=[a-f0-9-]+$/,
      ),
    });
  });

  it("never returns a direct browser path for metadata images", async () => {
    const selectorValues = new Map([
      ['link[rel="icon"]', "http://127.0.0.1:3000/probe"],
      ['meta[property="og:image"]', "https://cdn.example.com/preview.png"],
    ]);
    mockedLoad.mockReturnValue(((selector: string) => ({
      attr: () => selectorValues.get(selector),
      text: () => "",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://www.example.com/articles/final"),
    });

    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com/start",
      ),
    );

    await expect(response.json()).resolves.toMatchObject({
      image: expect.stringMatching(
        /^\/api\/metadata\/image\?url=https%3A%2F%2Fcdn.example.com%2Fpreview.png&v=[a-f0-9-]+$/,
      ),
    });
  });

  it("returns a client error when a redirect resolves to a private address", async () => {
    mockedFetchPublicHtml.mockRejectedValue(
      new MetadataFetchError("unsafe-target", "Unsafe target"),
    );

    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com",
      ),
    );

    expect(response.status).toBe(400);
  });

  it("distinguishes timeouts from other upstream failures", async () => {
    mockedFetchPublicHtml.mockRejectedValue(
      new MetadataFetchError("timeout", "Timed out"),
    );
    const request = createRequest(
      "https://projects.example.com/api/metadata?url=https://example.com",
    );

    await expect(GET(request)).resolves.toMatchObject({ status: 504 });

    mockedFetchPublicHtml.mockRejectedValue(new Error("Connection failed"));
    await expect(
      GET(createRequest(`${request.url}/other`)),
    ).resolves.toMatchObject({ status: 502 });
  });

  it("reuses cached metadata across callers but still authenticates each request", async () => {
    mockedLoad.mockReturnValue((() => ({
      attr: () => undefined,
      text: () => "Shared article",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://example.com/article"),
    });
    const request = createRequest(
      "https://projects.example.com/api/metadata?url=https://example.com/article",
    );
    expect((await GET(request)).status).toBe(200);
    expect((await GET(request)).status).toBe(200);
    expect(mockedFetchPublicHtml).toHaveBeenCalledTimes(1);
    expect(mockedAuth).toHaveBeenCalledTimes(2);
    mockedAuth.mockResolvedValue(null);
    expect((await GET(request)).status).toBe(401);
    expect(mockedFetchPublicHtml).toHaveBeenCalledTimes(1);
  });

  it("returns a retryable service status when distinct requests reach capacity", async () => {
    const get = jest
      .spyOn(metadataService, "get")
      .mockRejectedValueOnce(new MetadataCapacityError());
    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com/article",
      ),
    );
    expect(response.status).toBe(503);
    expect(mockedFetchPublicHtml).not.toHaveBeenCalled();
    get.mockRestore();
  });

  it("bounds cached text and rejects oversized image URLs", async () => {
    const selectors = new Map([
      ['meta[property="og:title"]', "t".repeat(5000)],
      ['meta[name="description"]', "d".repeat(10000)],
      ['meta[property="og:image"]', `https://example.com/${"x".repeat(9000)}`],
    ]);
    mockedLoad.mockReturnValue(((selector: string) => ({
      attr: () => selectors.get(selector),
      text: () => "",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://example.com"),
    });
    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com/article",
      ),
    );
    await expect(response.json()).resolves.toMatchObject({
      title: "t".repeat(512),
      description: "d".repeat(2048),
      image: undefined,
      previewImage: undefined,
    });
  });
});

describe("forced metadata refresh", () => {
  it("rate-limits forced upstream work", async () => {
    jest.mocked(allowPreviewRefresh).mockResolvedValueOnce(false);
    const response = await GET(
      createRequest(
        "https://projects.example.com/api/metadata?url=https://example.com&refresh=1",
      ),
    );
    expect(response.status).toBe(429);
  });

  it("fetches new metadata and image URLs even when the upstream image URL is unchanged", async () => {
    metadataService.clear();
    mockedAuth.mockResolvedValue(session);
    jest.mocked(allowPreviewRefresh).mockResolvedValue(true);
    mockedLoad.mockReturnValue(((selector: string) => ({
      attr: () =>
        selector === 'meta[property="og:image"]' ? "/og.png" : undefined,
      text: () => "Example",
    })) as never);
    mockedFetchPublicHtml.mockResolvedValue({
      body: Buffer.from("<html></html>"),
      finalUrl: new URL("https://example.com"),
    });
    const request = createRequest(
      "https://projects.example.com/api/metadata?url=https://example.com",
    );
    const first = await (await GET(request)).json();
    const next = await (
      await GET(createRequest(`${request.url}&refresh=1`))
    ).json();
    expect(next.version).not.toBe(first.version);
    expect(next.previewImage).not.toBe(first.previewImage);
    expect(next.fetchedAt).toEqual(expect.any(Number));
  });
});
