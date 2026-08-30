/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { NextRequest } from "next/server";
import * as cheerio from "cheerio";
import { auth } from "@/auth";
import { fetchPublicHtml } from "./safe-fetch";
import { MetadataFetchError } from "./target-policy";
import { GET } from "./route";

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
    jest.clearAllMocks();
    mockedAuth.mockResolvedValue(session);
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

    await expect(response.json()).resolves.toEqual({
      description: "Example description",
      image:
        "/api/metadata/image?url=https%3A%2F%2Fwww.example.com%2Farticles%2Fassets%2Ffavicon.png",
      title: "Example title",
    });
    expect(response.status).toBe(200);
    expect(mockedFetchPublicHtml).toHaveBeenCalledWith(
      new URL("https://example.com/start"),
      { signal: request.signal },
    );
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
      image:
        "/api/metadata/image?url=https%3A%2F%2Fcdn.example.com%2Fpreview.png",
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
    await expect(GET(request)).resolves.toMatchObject({ status: 502 });
  });
});
