/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { put, remove } from "@/lib/http";
import { deleteMemoryAction } from "./delete-memory";
import { updateMemoryAction } from "./update-memory";

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http", () => ({
  put: jest.fn(),
  remove: jest.fn(),
}));

jest.mock("@/utils", () => ({
  getApiError: jest.fn(),
}));

const authMock = jest.mocked(auth);
const putMock = jest.mocked(put);
const removeMock = jest.mocked(remove);

const session = {
  user: {
    id: "user-1",
    name: "Joseph Mukorivo",
    email: "joseph@example.com",
    image: null,
    username: "joseph",
    fullName: "Joseph Mukorivo",
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
  },
};

describe("memory mutation actions", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    putMock.mockResolvedValue({ data: null });
    removeMock.mockResolvedValue({ data: null });
  });

  it("keeps update and delete requests in the authenticated workspace context", async () => {
    await updateMemoryAction(
      "memory-1",
      { content: "Prefer concise summaries." },
      "complexus",
    );
    await deleteMemoryAction("memory-1", "complexus");

    expect(putMock).toHaveBeenCalledWith(
      "users/memory/memory-1",
      { content: "Prefer concise summaries." },
      { session, workspaceSlug: "complexus" },
    );
    expect(removeMock).toHaveBeenCalledWith("users/memory/memory-1", {
      session,
      workspaceSlug: "complexus",
    });
  });

  it("rejects unauthenticated mutations before the API call", async () => {
    authMock.mockResolvedValue(null);

    await expect(
      updateMemoryAction(
        "memory-1",
        { content: "Do not save this." },
        "complexus",
      ),
    ).resolves.toEqual({
      data: null,
      error: { message: "Authentication required" },
    });
    await expect(deleteMemoryAction("memory-1", "complexus")).resolves.toEqual({
      data: null,
      error: { message: "Authentication required" },
    });

    expect(putMock).not.toHaveBeenCalled();
    expect(removeMock).not.toHaveBeenCalled();
  });
});
