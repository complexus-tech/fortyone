/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { GoogleDriveFileReference } from "@/shared/google-drive/types";
import { GoogleDriveFileCard } from "./google-drive-file-card";

const mockOpenChatWithGoogleDriveFile = jest.fn();
const mockDeleteFile = jest.fn();
const mockRefreshFile = jest.fn();

jest.mock("@/context/chat-context", () => ({
  useChatContext: () => ({
    openChatWithGoogleDriveFile: mockOpenChatWithGoogleDriveFile,
  }),
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("./hooks", () => ({
  useDeleteGoogleDriveFile: () => ({
    isPending: false,
    mutate: mockDeleteFile,
  }),
  useRefreshGoogleDriveFile: () => ({
    isPending: false,
    mutate: mockRefreshFile,
  }),
}));

const target = { id: "story-1", type: "story" as const };

const createFile = (
  overrides: Partial<GoogleDriveFileReference> = {},
): GoogleDriveFileReference => ({
  availability: "available",
  connectionEmail: "joseph@example.com",
  createdAt: "2026-09-04T08:00:00Z",
  id: "reference-1",
  mimeType: "application/vnd.google-apps.document",
  modifiedTime: "2026-09-04T08:00:00Z",
  name: "Project brief",
  targetId: target.id,
  targetType: target.type,
  updatedAt: "2026-09-04T08:00:00Z",
  webViewLink: "https://docs.google.com/document/d/provider-file/edit",
  ...overrides,
});

const renderCard = (file: GoogleDriveFileReference) =>
  render(
    <GoogleDriveFileCard
      canEdit
      canUseGoogleContent
      file={file}
      onImport={jest.fn()}
      onPreviewError={jest.fn()}
      target={target}
    />,
  );

const openActions = (fileName: string) => {
  fireEvent.pointerDown(
    screen.getByRole("button", { name: `Actions for ${fileName}` }),
    { button: 0, ctrlKey: false },
  );
};

describe("GoogleDriveFileCard", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("uses the native Docs icon without repeating the file type", () => {
    renderCard(createFile());

    const docsMark = document.querySelector(
      'svg[viewBox="0 0 192 192"] path[fill="#3186FF"]',
    );
    const fileIcon = docsMark?.closest("span");
    const cardDetails = fileIcon?.parentElement;

    expect(docsMark).not.toBeNull();
    expect(fileIcon).toHaveClass("size-10");
    expect(cardDetails).toHaveClass("gap-2", "px-3");
    expect(screen.queryByText("Google Doc")).not.toBeInTheDocument();
    const email = screen.getByText("joseph@example.com");
    expect(email.parentElement?.querySelector(".ml-auto")).not.toBeNull();
    const actions = screen.getByRole("button", {
      name: "Actions for Project brief",
    });
    expect(actions.parentElement?.parentElement).toBe(
      screen.getByText("Project brief").parentElement,
    );
    expect(actions).toHaveClass("size-6", "p-0");
    expect(
      actions.querySelector('path[d="M11.9842 18H11.9932"]'),
    ).not.toBeNull();
  });

  it("shows Convert for Docs and keeps removal destructive in both themes", () => {
    renderCard(createFile());

    openActions("Project brief");

    expect(screen.getByRole("menuitem", { name: "Convert" })).toBeEnabled();
    const removeItem = screen.getByRole("menuitem", {
      name: "Remove attachment",
    });
    expect(removeItem).toHaveClass("text-danger", "dark:!text-danger");
    expect(removeItem.querySelector("svg")).toHaveClass(
      "text-danger",
      "dark:!text-danger",
    );
  });

  it("uses the native Sheets icon and hides Convert", () => {
    renderCard(
      createFile({
        mimeType: "application/vnd.google-apps.spreadsheet",
        name: "Launch plan",
        webViewLink:
          "https://docs.google.com/spreadsheets/d/provider-file/edit",
      }),
    );

    expect(
      document.querySelector('svg[viewBox="0 0 192 192"] path[fill="#009954"]'),
    ).not.toBeNull();
    expect(screen.queryByText("Google Sheet")).not.toBeInTheDocument();

    openActions("Launch plan");

    expect(
      screen.queryByRole("menuitem", { name: "Convert" }),
    ).not.toBeInTheDocument();
  });
});
