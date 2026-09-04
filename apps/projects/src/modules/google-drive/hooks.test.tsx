/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  attachGoogleDriveFilesAction,
  createGoogleDriveFileAction,
  importGoogleDriveFileAction,
} from "./actions";
import {
  googleDriveKeys,
  useAttachGoogleDriveFiles,
  useCreateGoogleDriveFile,
  useImportGoogleDriveFile,
} from "./hooks";
import type { GoogleDriveFileReference } from "./types";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { token: "session-token" } }),
}));
jest.mock("./actions", () => ({
  attachGoogleDriveFilesAction: jest.fn(),
  createGoogleDriveFileAction: jest.fn(),
  importGoogleDriveFileAction: jest.fn(),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));

const target = { id: "story-1", type: "story" as const };
const existingFile: GoogleDriveFileReference = {
  availability: "available",
  createdAt: "2026-09-03T08:00:00Z",
  id: "reference-1",
  mimeType: "application/vnd.google-apps.document",
  name: "Existing brief",
  targetId: target.id,
  targetType: target.type,
  updatedAt: "2026-09-03T08:00:00Z",
  webViewLink: "https://docs.google.com/document/d/google-file-1/edit",
};
const newFile: GoogleDriveFileReference = {
  ...existingFile,
  id: "reference-2",
  name: "New brief",
  webViewLink: "https://docs.google.com/document/d/google-file-2/edit",
};

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });

const createWrapper = (queryClient: QueryClient) =>
  function TestQueryClientProvider({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };

beforeEach(() => {
  jest.clearAllMocks();
});

describe("Google Drive mutations", () => {
  it("merges newly attached references without hiding existing files", async () => {
    const queryClient = createQueryClient();
    const queryKey = googleDriveKeys.files("acme", target);
    queryClient.setQueryData(queryKey, [existingFile]);
    jest.mocked(attachGoogleDriveFilesAction).mockResolvedValue({
      data: [newFile],
    });
    const { result } = renderHook(() => useAttachGoogleDriveFiles(target), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate([{ id: "google-file-2" }]);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(
      queryClient
        .getQueryData<GoogleDriveFileReference[]>(queryKey)
        ?.map(({ id }) => id),
    ).toEqual(["reference-1", "reference-2"]);
  });

  it("keeps the create operation id stable at the action boundary", async () => {
    const queryClient = createQueryClient();
    jest
      .mocked(createGoogleDriveFileAction)
      .mockResolvedValue({ data: newFile });
    const { result } = renderHook(() => useCreateGoogleDriveFile(target), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate({
        fileType: "document",
        idempotencyKey: "create-operation-1",
        title: "New brief",
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(createGoogleDriveFileAction).toHaveBeenCalledWith(
      "acme",
      target,
      "document",
      "New brief",
      "create-operation-1",
    );
  });

  it("keeps the import operation id stable at the action boundary", async () => {
    const queryClient = createQueryClient();
    jest.mocked(importGoogleDriveFileAction).mockResolvedValue({
      data: {
        documentId: "document-1",
        sourceReferenceId: "reference-1",
      },
    });
    const { result } = renderHook(() => useImportGoogleDriveFile(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate({
        idempotencyKey: "import-operation-1",
        referenceId: "reference-1",
        visibility: "private",
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(importGoogleDriveFileAction).toHaveBeenCalledWith(
      "acme",
      "reference-1",
      "private",
      "import-operation-1",
    );
  });
});
