/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { RequestOptions } from "api-client";
import { get, patch, post, put, remove, type WorkspaceCtx } from "./fetch";
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
