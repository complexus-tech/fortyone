/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { NextRequest } from "next/server";
import { auth } from "@/auth";
import { fetchPublicImage } from "../safe-fetch";
import { MetadataFetchError } from "../target-policy";
import { GET } from "./route";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("../safe-fetch", () => ({ fetchPublicImage: jest.fn() }));
jest.mock("next/server", () => {
  class MockNextResponse {
    readonly body: Uint8Array | null;
    readonly headers: Headers;
    readonly status: number;

    constructor(
      body: Uint8Array | null = null,
      init?: { headers?: HeadersInit; status?: number },
    ) {
      this.body = body;
      this.headers = new Headers(init?.headers);
      this.status = init?.status ?? 200;
    }

    static json(value: unknown, init?: { status?: number }) {
      return {
        json: async () => value,
        status: init?.status ?? 200,
      };
    }

    async arrayBuffer() {
      if (!this.body) return new ArrayBuffer(0);
      return this.body.buffer.slice(
        this.body.byteOffset,
        this.body.byteOffset + this.body.byteLength,
      );
    }
  }

  return { NextResponse: MockNextResponse };
});

const mockedAuth = jest.mocked(auth);
const mockedFetchPublicImage = jest.mocked(fetchPublicImage);

const createRequest = (target: string) =>
  ({
    signal: new AbortController().signal,
    url: `https://projects.example.com/api/metadata/image?url=${encodeURIComponent(
      target,
    )}`,
  }) as NextRequest;

describe("GET /api/metadata/image", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockedAuth.mockResolvedValue({ user: { id: "user-1" } } as never);
  });

  it("authenticates before fetching the image", async () => {
    mockedAuth.mockResolvedValue(null);

    const response = await GET(createRequest("http://127.0.0.1/image.png"));

    expect(response.status).toBe(401);
    expect(mockedFetchPublicImage).not.toHaveBeenCalled();
  });

  it("returns a bounded public raster image with defensive headers", async () => {
    const body = Buffer.from("raster-image");
    mockedFetchPublicImage.mockResolvedValue({
      body,
      contentType: "image/png",
      finalUrl: new URL("https://cdn.example.com/final.png"),
    });
    const request = createRequest("https://cdn.example.com/image.png");

    const response = await GET(request);

    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toBe("image/png");
    expect(response.headers.get("x-content-type-options")).toBe("nosniff");
    expect(response.headers.get("content-security-policy")).toContain(
      "sandbox",
    );
    expect(Buffer.from(await response.arrayBuffer())).toEqual(body);
    expect(mockedFetchPublicImage).toHaveBeenCalledWith(
      new URL("https://cdn.example.com/image.png"),
      { signal: request.signal },
    );
  });

  it("rejects private targets without returning image bytes", async () => {
    mockedFetchPublicImage.mockRejectedValue(
      new MetadataFetchError("unsafe-target", "Unsafe target"),
    );

    const response = await GET(createRequest("http://127.0.0.1/image.png"));

    expect(response.status).toBe(400);
  });

  it("distinguishes timeouts from other upstream image failures", async () => {
    mockedFetchPublicImage.mockRejectedValue(
      new MetadataFetchError("timeout", "Timed out"),
    );
    const request = createRequest("https://cdn.example.com/image.png");

    await expect(GET(request)).resolves.toMatchObject({ status: 504 });

    mockedFetchPublicImage.mockRejectedValue(
      new MetadataFetchError("unsupported-response", "Not a raster image"),
    );
    await expect(GET(request)).resolves.toMatchObject({ status: 502 });
  });
});
