import { createApiClient } from "api-client";
import { deleteAccount } from "./action";

jest.mock("api-client", () => ({ createApiClient: jest.fn() }));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.fortyone.app",
}));

const request = jest.fn();
const confirmedUserId = "07b8ee3c-31be-4ccd-8529-5c276a8c6096";
beforeEach(() => {
  jest.clearAllMocks();
  (createApiClient as jest.Mock).mockReturnValue({ delete: request });
});

it("deletes the current account without workspace scoping or automatic retries", async () => {
  request.mockResolvedValue({ status: 204 });
  await expect(deleteAccount(confirmedUserId)).resolves.toBe("deleted");
  expect(request).toHaveBeenCalledWith("users/account", {
    retry: 0,
    json: { expectedUserId: confirmedUserId },
  });
});

it("distinguishes accepted provider cleanup from completed cleanup", async () => {
  request.mockResolvedValue({
    status: 202,
    json: async () => ({ data: { status: "cleanup_pending" } }),
  });
  await expect(deleteAccount(confirmedUserId)).resolves.toBe("cleanup_pending");
});

it.each([
  { status: 200 },
  { status: 202, json: async () => ({ data: { status: "unknown" } }) },
])(
  "does not treat unexpected response $status as confirmed deletion",
  async (response) => {
    request.mockResolvedValue(response);
    await expect(deleteAccount(confirmedUserId)).rejects.toThrow(
      "did not confirm",
    );
  },
);

it("preserves actionable server conflicts", async () => {
  const error = new Error("Transfer administrator access for Acme first.");
  request.mockRejectedValue(error);
  await expect(deleteAccount(confirmedUserId)).rejects.toBe(error);
});

it("rejects missing confirmation identity before sending a request", async () => {
  await expect(deleteAccount("")).rejects.toThrow("could not be verified");
  expect(request).not.toHaveBeenCalled();
});
