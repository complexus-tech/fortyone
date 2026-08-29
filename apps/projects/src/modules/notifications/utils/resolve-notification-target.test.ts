/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import { resolveNotificationTarget } from "./resolve-notification-target";

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

describe("resolveNotificationTarget", () => {
  it("returns a successfully loaded target", async () => {
    const target = { id: "objective-1" };

    await expect(
      resolveNotificationTarget(async () => target),
    ).resolves.toEqual({ status: "found", value: target });
  });

  it("treats an empty successful response as a terminal missing target", async () => {
    await expect(resolveNotificationTarget(async () => null)).resolves.toEqual({
      reason: "not_found",
      status: "terminal",
    });
  });

  it.each([
    [403, "forbidden"],
    [404, "not_found"],
  ] as const)(
    "treats an explicit %i API response as a terminal target",
    async (status, reason) => {
      await expect(
        resolveNotificationTarget(async () => {
          throw new ApiError("Target unavailable", status, null);
        }),
      ).resolves.toEqual({ reason, status: "terminal" });
    },
  );

  it.each([
    new ApiError("Server unavailable", 500, null),
    new TypeError("fetch failed"),
  ])("preserves a transient or unexpected failure", async (error) => {
    await expect(
      resolveNotificationTarget(async () => {
        throw error;
      }),
    ).rejects.toBe(error);
  });
});
