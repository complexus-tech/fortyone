/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import ky, { HTTPError } from "ky";
import {
  acceptUserInvitation,
  InvitationUnavailableError,
} from "./accept-user-invitation";

jest.mock("ky", () => ({
  __esModule: true,
  default: { post: jest.fn() },
  HTTPError: class extends Error {
    constructor(public response: { status: number }) {
      super("HTTP error");
    }
  },
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.com",
}));
jest.mock("@/lib/fetch-error", () => ({
  requestError: jest
    .fn()
    .mockResolvedValue({ error: { message: "Request failed" } }),
}));

describe("acceptUserInvitation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });
  it("uses the authenticated account endpoint with an invitation ID", async () => {
    jest
      .mocked(ky.post)
      .mockResolvedValue({ json: async () => ({ data: null }) } as Response);
    await expect(acceptUserInvitation("invitation-id")).resolves.toEqual({
      data: null,
    });
    expect(ky.post).toHaveBeenCalledWith(
      "https://api.example.com/users/me/invitations/invitation-id/accept",
      { credentials: "include" },
    );
  });
  it.each([404, 401, 403, 429, 500])(
    "classifies HTTP %s without discarding retryable invitations",
    async (status) => {
      const error = new HTTPError(
        { status } as Response,
        {} as Request,
        {} as never,
      );
      jest.mocked(ky.post).mockRejectedValueOnce(error);
      const result = await acceptUserInvitation("invitation-id").catch(
        (failure: unknown) => failure,
      );
      expect(result).toBeInstanceOf(Error);
      expect(result instanceof InvitationUnavailableError).toBe(status === 404);
    },
  );
});
