/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { post, remove } from "@/lib/http";
import {
  attachGoogleDriveFilesAction,
  createGoogleDriveConnectSessionAction,
  createGoogleDriveFileAction,
  deleteGoogleDriveFileAction,
  importGoogleDriveFileAction,
  refreshGoogleDriveFileAction,
} from "./actions";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({ post: jest.fn(), remove: jest.fn() }));
jest.mock("@/utils", () => ({
  getApiError: (error: Error) => ({
    data: null,
    error: { message: error.message },
  }),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);
const removeMock = jest.mocked(remove);
const target = { id: "story-1", type: "story" as const };

beforeEach(() => {
  jest.clearAllMocks();
  authMock.mockResolvedValue({ token: "session-token" } as unknown as Awaited<
    ReturnType<typeof auth>
  >);
});

describe("Google Drive actions", () => {
  it("passes a validated return destination to the connection session", async () => {
    postMock.mockResolvedValue({
      data: { authorizationUrl: "https://accounts.google.test/oauth" },
    });

    await createGoogleDriveConnectSessionAction(
      "acme",
      "https://acme.fortyone.app/work/ENG-1",
    );

    expect(postMock).toHaveBeenCalledWith(
      "integrations/google-drive/connect-session",
      { returnUrl: "https://acme.fortyone.app/work/ENG-1" },
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("submits selected file identities for server-side validation", async () => {
    postMock.mockResolvedValue({ data: [] });
    const files = [
      {
        id: "drive-file-1",
        mimeType: "application/vnd.google-apps.document",
        name: "Launch brief",
        resourceKey: "resource-key",
      },
    ];

    await attachGoogleDriveFilesAction("acme", target, files);

    expect(postMock).toHaveBeenCalledWith(
      "google-drive/files",
      { files, targetId: "story-1", targetType: "story" },
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("preserves the caller's idempotency key when creating a file", async () => {
    postMock.mockResolvedValue({ data: { id: "reference-1" } });

    await createGoogleDriveFileAction(
      "acme",
      target,
      "document",
      "Launch brief",
      "create-operation-1",
    );

    expect(postMock).toHaveBeenCalledWith(
      "google-drive/files/create",
      {
        fileType: "document",
        targetId: "story-1",
        targetType: "story",
        title: "Launch brief",
      },
      expect.objectContaining({ workspaceSlug: "acme" }),
      { headers: { "Idempotency-Key": "create-operation-1" } },
    );
  });

  it("uses reference-scoped delete and import routes", async () => {
    removeMock.mockResolvedValue({ data: null });
    postMock.mockResolvedValue({
      data: { documentId: "document-1", sourceReferenceId: "reference-1" },
    });

    await deleteGoogleDriveFileAction("acme", "reference-1");
    await refreshGoogleDriveFileAction("acme", "reference-1");
    await importGoogleDriveFileAction(
      "acme",
      "reference-1",
      "private",
      "import-operation-1",
    );

    expect(removeMock).toHaveBeenCalledWith(
      "google-drive/files/reference-1",
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
    expect(postMock).toHaveBeenCalledWith(
      "google-drive/files/reference-1/refresh",
      {},
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
    expect(postMock).toHaveBeenCalledWith(
      "google-drive/files/reference-1/imports",
      { visibility: "private" },
      expect.objectContaining({ workspaceSlug: "acme" }),
      { headers: { "Idempotency-Key": "import-operation-1" } },
    );
  });
});
