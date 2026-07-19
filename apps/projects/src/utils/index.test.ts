/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import { getApiError } from ".";

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
