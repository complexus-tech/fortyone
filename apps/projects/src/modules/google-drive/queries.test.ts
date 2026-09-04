/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { get } from "@/lib/http";
import { getGoogleDriveFiles, getGoogleDriveIntegration } from "./queries";

jest.mock("@/lib/http", () => ({ get: jest.fn() }));

const getMock = jest.mocked(get);
const context = { workspaceSlug: "acme" };

beforeEach(() => {
  jest.clearAllMocks();
});

describe("Google Drive queries", () => {
  it("loads the personal workspace connection", async () => {
    getMock.mockResolvedValue({
      data: {
        configured: true,
        connected: false,
        requiresReauthorization: false,
        status: "disconnected",
      },
    });

    await getGoogleDriveIntegration(context);

    expect(getMock).toHaveBeenCalledWith("integrations/google-drive", context);
  });

  it("encodes a typed target when listing references", async () => {
    getMock.mockResolvedValue({ data: [] });

    await getGoogleDriveFiles(context, {
      id: "objective/with spaces",
      type: "objective",
    });

    expect(getMock).toHaveBeenCalledWith(
      "google-drive/files?targetId=objective%2Fwith+spaces&targetType=objective",
      context,
    );
  });
});
