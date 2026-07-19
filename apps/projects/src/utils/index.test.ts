/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import type { Workspace } from "@/types";
import { getApiError, getRedirectUrl } from ".";

jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {
    data: unknown;
    status: number;

    constructor(message: string, status: number, data: unknown) {
      super(message);
      this.data = data;
      this.status = status;
    }
  },
}));

describe("getApiError", () => {
  it("preserves errors returned by the shared API client", () => {
    const response = {
      data: null,
      error: { message: "A feedback item with this title already exists" },
    };

    expect(getApiError(new ApiError("Request failed", 400, response))).toEqual(
      response,
    );
  });
});

describe("getRedirectUrl", () => {
  it("returns an explicit safe callback before workspace routing", () => {
    const workspace = {
      avatarUrl: null,
      color: "#000000",
      createdAt: "2026-01-01T00:00:00.000Z",
      deletedAt: null,
      id: "workspace-1",
      isActive: true,
      name: "Product",
      slug: "product",
      trialEndsOn: null,
      updatedAt: "2026-01-01T00:00:00.000Z",
      userRole: "member",
    } satisfies Workspace;

    expect(
      getRedirectUrl(
        [workspace],
        [],
        "workspace-1",
        "/portal/city-roads/feedback",
      ),
    ).toBe("/portal/city-roads/feedback");
  });

  it("sends accounts without workspaces to global account settings", () => {
    expect(getRedirectUrl([])).toBe("/account");
  });

  it("does not redirect to an external callback URL", () => {
    expect(
      getRedirectUrl([], [], undefined, "https://malicious.example.com"),
    ).toBe("/account");
  });
});
