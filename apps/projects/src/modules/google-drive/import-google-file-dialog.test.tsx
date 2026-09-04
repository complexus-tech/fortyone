/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { ImportGoogleFileDialog } from "./import-google-file-dialog";
import type { GoogleDriveFileReference } from "./types";

const mockMutate = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: jest.fn() }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("./hooks", () => ({
  useImportGoogleDriveFile: () => ({
    isPending: false,
    mutate: mockMutate,
  }),
}));

const file: GoogleDriveFileReference = {
  availability: "available",
  createdAt: "2026-09-04T08:00:00Z",
  id: "reference-1",
  mimeType: "application/vnd.google-apps.document",
  name: "Launch brief",
  targetId: "story-1",
  targetType: "story",
  updatedAt: "2026-09-04T08:00:00Z",
  webViewLink: "https://docs.google.com/document/d/provider-file/edit",
};

beforeEach(() => {
  jest.clearAllMocks();
  Object.defineProperty(globalThis.crypto, "randomUUID", {
    configurable: true,
    value: jest.fn(() => "00000000-0000-4000-8000-000000000001"),
  });
});

describe("ImportGoogleFileDialog", () => {
  it("defaults persisted Google content to private and reuses the operation key on retry", () => {
    const { rerender } = render(
      <ImportGoogleFileDialog file={file} onOpenChange={jest.fn()} />,
    );

    const importButton = screen.getByRole("button", {
      name: "Import snapshot",
    });
    fireEvent.click(importButton);
    rerender(
      <ImportGoogleFileDialog file={{ ...file }} onOpenChange={jest.fn()} />,
    );
    fireEvent.click(importButton);

    expect(mockMutate).toHaveBeenCalledTimes(2);
    const firstInput = mockMutate.mock.calls[0]?.[0];
    const secondInput = mockMutate.mock.calls[1]?.[0];
    expect(firstInput).toEqual({
      idempotencyKey: "00000000-0000-4000-8000-000000000001",
      referenceId: "reference-1",
      visibility: "private",
    });
    expect(secondInput.idempotencyKey).toBe(firstInput.idempotencyKey);
  });
});
