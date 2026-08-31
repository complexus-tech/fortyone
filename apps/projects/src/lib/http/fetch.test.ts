/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { RequestOptions } from "api-client";
import {
  ApiContractError,
  get,
  patch,
  post,
  put,
  remove,
  type WorkspaceCtx,
} from "./fetch";
import { installRequestOptionsScopeResolver } from "./request-options-scope";

const clientMethods = {
  delete: jest.fn(),
  get: jest.fn(),
  patch: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
};

jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {},
  createApiClient: jest.fn(() => clientMethods),
}));

jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.com",
}));

const ctx: WorkspaceCtx = { workspaceSlug: "fortyone" };
let uninstallScope: (() => void) | undefined;

const response = () =>
  ({
    status: 200,
    text: async () => JSON.stringify({ data: { id: "result-1" } }),
  }) as Response;

beforeEach(() => {
  jest.clearAllMocks();
  for (const method of Object.values(clientMethods)) {
    method.mockImplementation(async () => response());
  }
});

afterEach(() => {
  uninstallScope?.();
  uninstallScope = undefined;
});

describe("workspace HTTP request scoping", () => {
  it("leaves non-Maya request options unchanged", async () => {
    const signal = new AbortController().signal;
    const options: RequestOptions = {
      headers: { "x-request-id": "request-1" },
      retry: { backoffLimit: 9_000, limit: 7, maxRetryAfter: 12_000 },
      signal,
    };

    await get("stories", ctx, options);

    expect(clientMethods.get).toHaveBeenCalledWith("stories", options);
    expect(clientMethods.get.mock.calls[0]?.[1]).toBe(options);
  });

  it("composes Maya cancellation and clamps retry delays", async () => {
    const mayaController = new AbortController();
    const requestController = new AbortController();
    uninstallScope = installRequestOptionsScopeResolver(() => ({
      signal: mayaController.signal,
    }));

    await get("stories", ctx, {
      headers: { "x-request-id": "request-2" },
      retry: {
        backoffLimit: 7_000,
        limit: 4,
        maxRetryAfter: 8_000,
        statusCodes: [408, 500],
      },
      signal: requestController.signal,
    });

    const options = clientMethods.get.mock.calls[0]?.[1] as RequestOptions;
    expect(options).toEqual(
      expect.objectContaining({
        headers: { "x-request-id": "request-2" },
        retry: {
          backoffLimit: 1_000,
          limit: 4,
          maxRetryAfter: 2_000,
          statusCodes: [408, 500],
        },
      }),
    );
    expect(options.signal).not.toBe(mayaController.signal);
    expect(options.signal).not.toBe(requestController.signal);
    expect(options.signal?.aborted).toBe(false);

    requestController.abort(new Error("caller cancelled"));
    expect(options.signal?.aborted).toBe(true);
    expect(options.signal?.reason).toEqual(new Error("caller cancelled"));
  });

  it("preserves retry zero and uses the Maya signal directly", async () => {
    const mayaSignal = new AbortController().signal;
    uninstallScope = installRequestOptionsScopeResolver(() => ({
      signal: mayaSignal,
    }));

    await get("stories", ctx, { retry: 0 });

    expect(clientMethods.get).toHaveBeenCalledWith("stories", {
      retry: 0,
      signal: mayaSignal,
    });
  });

  it("preserves numeric retry limits while bounding their delays", async () => {
    const mayaSignal = new AbortController().signal;
    uninstallScope = installRequestOptionsScopeResolver(() => ({
      signal: mayaSignal,
    }));

    await get("stories", ctx, { retry: 5 });

    expect(clientMethods.get).toHaveBeenCalledWith("stories", {
      retry: { backoffLimit: 1_000, limit: 5, maxRetryAfter: 2_000 },
      signal: mayaSignal,
    });
  });

  it("applies the scope to every workspace HTTP verb", async () => {
    const mayaSignal = new AbortController().signal;
    uninstallScope = installRequestOptionsScopeResolver(() => ({
      signal: mayaSignal,
    }));

    await post("stories", { title: "Story" }, ctx);
    await put("stories/story-1", { title: "Updated" }, ctx);
    await patch("stories/story-1", { title: "Patched" }, ctx);
    await remove("stories/story-1", ctx);

    const expectedOptions = {
      retry: { backoffLimit: 1_000, maxRetryAfter: 2_000 },
      signal: mayaSignal,
    };
    expect(clientMethods.post).toHaveBeenCalledWith("stories", {
      json: { title: "Story" },
      ...expectedOptions,
    });
    expect(clientMethods.put).toHaveBeenCalledWith("stories/story-1", {
      json: { title: "Updated" },
      ...expectedOptions,
    });
    expect(clientMethods.patch).toHaveBeenCalledWith("stories/story-1", {
      json: { title: "Patched" },
      ...expectedOptions,
    });
    expect(clientMethods.delete).toHaveBeenCalledWith(
      "stories/story-1",
      expectedOptions,
    );
  });
});

describe("workspace HTTP response contracts", () => {
  it("preserves explicit no-content responses", async () => {
    clientMethods.delete.mockResolvedValue({
      status: 204,
      text: jest.fn(),
    } as unknown as Response);

    await expect(remove("stories/story-1", ctx)).resolves.toEqual({
      data: null,
    });
  });

  it("applies runtime decoders to explicit no-content responses", async () => {
    const decodeNoContent = jest.fn((value: unknown) => value);
    clientMethods.delete.mockResolvedValue({
      status: 204,
      text: jest.fn(),
    } as unknown as Response);

    await expect(
      remove("stories/story-1", ctx, undefined, decodeNoContent),
    ).resolves.toEqual({ data: null });
    expect(decodeNoContent).toHaveBeenCalledWith({ data: null });
  });

  it.each([
    ["an empty successful body", ""],
    ["a whitespace-only successful body", "  \n"],
  ])("rejects %s", async (_name, body) => {
    clientMethods.get.mockResolvedValue({
      status: 200,
      text: async () => body,
    } as Response);

    await expect(get("stories", ctx)).rejects.toMatchObject({
      name: "ApiContractError",
      status: 200,
    });
  });

  it("rejects invalid JSON without exposing the response body", async () => {
    clientMethods.get.mockResolvedValue({
      status: 200,
      text: async () => '<html data-secret="provider-token">broken</html>',
    } as Response);

    await expect(get("stories", ctx)).rejects.toEqual(
      expect.objectContaining({
        message: "API returned invalid JSON with status 200",
        name: "ApiContractError",
        status: 200,
      }),
    );

    try {
      await get("stories", ctx);
    } catch (error) {
      expect(String(error)).not.toContain("provider-token");
      expect((error as Error).cause).toBeUndefined();
    }
  });

  it("supports feature-owned runtime decoders", async () => {
    const decodeStory = (value: unknown) => {
      if (
        typeof value !== "object" ||
        value === null ||
        !("data" in value) ||
        typeof value.data !== "object" ||
        value.data === null ||
        !("id" in value.data) ||
        typeof value.data.id !== "string"
      ) {
        throw new Error("Invalid story response");
      }

      return { data: { id: value.data.id } };
    };

    await expect(
      get("stories/story-1", ctx, undefined, decodeStory),
    ).resolves.toEqual({ data: { id: "result-1" } });

    clientMethods.get.mockResolvedValue({
      status: 200,
      text: async () => '{"data":{"id":42}}',
    } as Response);

    await expect(
      get("stories/story-1", ctx, undefined, decodeStory),
    ).rejects.toBeInstanceOf(ApiContractError);
  });

  it("does not retain body values thrown by runtime decoders", async () => {
    clientMethods.get.mockResolvedValue({
      status: 200,
      text: async () => '{"data":{"secret":"provider-token"}}',
    } as unknown as Response);

    try {
      await get("stories/story-1", ctx, undefined, (value) => {
        throw new Error(JSON.stringify(value));
      });
      throw new Error("Expected runtime decoding to fail");
    } catch (error) {
      expect(error).toBeInstanceOf(ApiContractError);
      expect(String(error)).not.toContain("provider-token");
      expect((error as Error).cause).toBeUndefined();
    }
  });
});
