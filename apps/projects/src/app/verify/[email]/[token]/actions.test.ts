/* global beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { ApiError, post } from "api-client";
import { logIn } from "./actions";

jest.mock("api-client", () => ({
  post: jest.fn(),
  ApiError: class ApiError extends Error {
    constructor(
      message: string,
      public status: number,
      public data: unknown,
    ) {
      super(message);
    }
  },
}));

beforeEach(() => jest.resetAllMocks());

it("shows account recovery guidance when a verified email cannot start a session", async () => {
  jest
    .mocked(post)
    .mockRejectedValue(new ApiError("invalid credentials", 401, null));
  await expect(logIn("returning@example.com", "123456")).resolves.toEqual({
    error: "account_unavailable",
  });
});

it("keeps invalid and expired verification codes as retryable errors", async () => {
  jest.mocked(post).mockRejectedValue(new ApiError("Invalid token", 400, null));
  await expect(logIn("returning@example.com", "123456")).resolves.toEqual({
    error: "Invalid token",
  });
});

it("continues after successful email verification", async () => {
  jest.mocked(post).mockResolvedValue({});
  await expect(logIn("returning@example.com", "123456")).resolves.toEqual({
    error: null,
  });
});
