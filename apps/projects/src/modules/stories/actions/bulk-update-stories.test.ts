import { ApiError } from "api-client";
import { auth } from "@/auth";
import { put } from "@/lib/http";
import { bulkUpdateAction } from "./bulk-update-stories";

jest.mock("api-client", () => ({
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

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({ put: jest.fn() }));

const payload = {
  storyIds: ["story-1", "story-2"],
  updates: { statusId: "backlog" },
};

describe("bulkUpdateAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(auth).mockResolvedValue(null);
  });

  it("allows a bounded bulk request without automatically resending the batch", async () => {
    const response = {
      data: {
        totalCount: 2,
        succeededCount: 2,
        failedCount: 0,
        partial: false,
        items: payload.storyIds.map((storyId) => ({ storyId, success: true })),
      },
    };
    jest.mocked(put).mockResolvedValue(response);

    await expect(bulkUpdateAction(payload, "fortyone")).resolves.toBe(response);

    expect(put).toHaveBeenCalledTimes(1);
    expect(put).toHaveBeenCalledWith(
      "stories",
      payload,
      { session: null, workspaceSlug: "fortyone" },
      { timeout: 60_000, retry: 0 },
    );
  });

  it("explains a timeout without claiming the server saved nothing", async () => {
    const timeout = new Error("Request timed out");
    timeout.name = "TimeoutError";
    jest.mocked(put).mockRejectedValue(timeout);

    await expect(bulkUpdateAction(payload, "fortyone")).resolves.toEqual({
      data: null,
      error: {
        message:
          "The bulk update timed out. Refresh the stories to check which changes were saved before retrying.",
      },
    });
    expect(put).toHaveBeenCalledTimes(1);
  });

  it("preserves the API's specific rejection message", async () => {
    const response = {
      data: null,
      error: { message: "The selected status does not belong to this team." },
    };
    jest
      .mocked(put)
      .mockRejectedValue(new ApiError("Invalid status", 400, response));

    await expect(bulkUpdateAction(payload, "fortyone")).resolves.toEqual(
      response,
    );
  });
});
