/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { GoogleDriveFileSection } from "./google-drive-file-section";
import type { GoogleDriveFileReference } from "./types";

const mockUseGoogleDriveFiles = jest.fn();
const mockUseGoogleDriveIntegration = jest.fn();
const mockConnect = jest.fn();

jest.mock("./hooks", () => ({
  useCreateGoogleDriveConnectSession: () => ({
    isPending: false,
    mutate: mockConnect,
  }),
  useGoogleDriveFiles: () => mockUseGoogleDriveFiles(),
  useGoogleDriveIntegration: () => mockUseGoogleDriveIntegration(),
}));
jest.mock("./google-drive-picker-button", () => ({
  GoogleDrivePickerButton: ({
    children = "Attach from Drive",
    variant,
  }: {
    children?: ReactNode;
    variant?: string;
  }) => (
    <button data-variant={variant} type="button">
      {children}
    </button>
  ),
}));
jest.mock("./google-drive-file-card", () => ({
  GoogleDriveFileCard: ({
    file,
  }: {
    file: GoogleDriveFileReference;
    onPreviewError: () => void;
  }) => <div>{file.name}</div>,
}));
jest.mock("./create-google-file-dialog", () => ({
  CreateGoogleFileDialog: ({
    initialFileType,
  }: {
    initialFileType: string;
  }) => <div role="dialog">Create {initialFileType}</div>,
}));
jest.mock("./import-google-file-dialog", () => ({
  ImportGoogleFileDialog: () => null,
}));

const target = { id: "story-1", type: "story" as const };
const file: GoogleDriveFileReference = {
  availability: "available",
  createdAt: "2026-09-04T08:00:00Z",
  id: "reference-1",
  mimeType: "application/vnd.google-apps.spreadsheet",
  name: "Launch plan",
  targetId: target.id,
  targetType: target.type,
  updatedAt: "2026-09-04T08:00:00Z",
  webViewLink: "https://docs.google.com/spreadsheets/d/provider-file/edit",
};

beforeEach(() => {
  jest.clearAllMocks();
  mockUseGoogleDriveIntegration.mockReturnValue({
    data: {
      configured: true,
      connected: true,
      requiresReauthorization: false,
    },
    isPending: false,
  });
  mockUseGoogleDriveFiles.mockReturnValue({
    data: [],
    isError: false,
    isPending: false,
  });
});

describe("GoogleDriveFileSection", () => {
  it("uses one compact Drive card for the connected empty state", () => {
    render(<GoogleDriveFileSection canEdit target={target} />);

    expect(screen.queryByText("Google Drive")).not.toBeInTheDocument();
    expect(screen.getByText("Bring in work from Google")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Attach from Drive" }),
    ).toHaveAttribute("data-variant", "naked");
    expect(
      document.querySelector('svg[viewBox="0 0 192 192"]'),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create Google file" }),
    ).toHaveTextContent("Create Google file");

    fireEvent.pointerDown(
      screen.getByRole("button", { name: "Create Google file" }),
      { button: 0, ctrlKey: false },
    );

    const docItem = screen.getByRole("menuitem", { name: "Google Doc" });
    const sheetItem = screen.getByRole("menuitem", { name: "Google Sheet" });
    expect(
      docItem.querySelector('svg[viewBox="0 0 192 192"] path[fill="#3186FF"]'),
    ).not.toBeNull();
    expect(
      sheetItem.querySelector(
        'svg[viewBox="0 0 192 192"] path[fill="#009954"]',
      ),
    ).not.toBeNull();

    fireEvent.click(docItem);
    expect(screen.getByRole("dialog")).toHaveTextContent("Create document");
  });

  it("retains the section header and actions when files exist", () => {
    mockUseGoogleDriveFiles.mockReturnValue({
      data: [file],
      isError: false,
      isPending: false,
    });

    render(<GoogleDriveFileSection canEdit target={target} />);

    expect(screen.getByText("Google Drive")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Attach from Drive" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create Google file" }),
    ).toHaveTextContent("Create Google file");
    expect(
      screen.queryByText("Bring in work from Google"),
    ).not.toBeInTheDocument();
    expect(screen.getByText("Launch plan")).toBeInTheDocument();
  });
});
