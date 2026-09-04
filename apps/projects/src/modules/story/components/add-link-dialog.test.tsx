/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useCreateLinkMutation } from "@/lib/hooks/create-link-mutation";
import { useDeleteLinkMutation } from "@/lib/hooks/delete-link-mutation";
import { useUpdateLinkMutation } from "@/lib/hooks/update-link-mutation";
import { AddLinkDialog } from "./add-link-dialog";

const createLink = jest.fn();
const deleteLink = jest.fn();
const updateLink = jest.fn();
const mockAttachGoogleDriveFiles = jest.fn();

jest.mock("@/lib/hooks/create-link-mutation", () => ({
  useCreateLinkMutation: jest.fn(),
}));
jest.mock("@/lib/hooks/delete-link-mutation", () => ({
  useDeleteLinkMutation: jest.fn(),
}));
jest.mock("@/lib/hooks/update-link-mutation", () => ({
  useUpdateLinkMutation: jest.fn(),
}));
jest.mock("@/lib/hooks/figma", () => ({
  useLinkFigmaStory: () => ({ mutate: jest.fn() }),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "story" }),
}));
jest.mock("@/modules/settings/workspace/integrations/figma/url", () => ({
  isFigmaURL: () => false,
}));
jest.mock("@/modules/google-drive/public/files", () => ({
  GoogleDrivePickerDialog: ({
    onAttached,
    onClose,
  }: {
    onAttached: () => void;
    onClose: () => void;
  }) => (
    <div data-testid="google-drive-picker">
      <button onClick={onClose} type="button">
        Cancel picker
      </button>
      <button onClick={onAttached} type="button">
        Attach file
      </button>
    </div>
  ),
  useAttachGoogleDriveFiles: () => ({
    mutateAsync: mockAttachGoogleDriveFiles,
  }),
}));

const googleSheetURL =
  "https://docs.google.com/spreadsheets/d/google-sheet-1/edit";

const renderDialog = () =>
  render(<AddLinkDialog isOpen setIsOpen={jest.fn()} storyId="story-1" />);

const submitGoogleSheet = () => {
  fireEvent.change(screen.getByPlaceholderText("https://..."), {
    target: { value: googleSheetURL },
  });
  fireEvent.click(screen.getByRole("button", { name: "Add link" }));
};

describe("AddLinkDialog Google Drive promotion", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockAttachGoogleDriveFiles.mockResolvedValue([]);
    jest.mocked(useCreateLinkMutation).mockReturnValue({
      mutate: createLink,
    } as unknown as ReturnType<typeof useCreateLinkMutation>);
    jest.mocked(useDeleteLinkMutation).mockReturnValue({
      mutate: deleteLink,
    } as unknown as ReturnType<typeof useDeleteLinkMutation>);
    jest.mocked(useUpdateLinkMutation).mockReturnValue({
      mutate: updateLink,
    } as unknown as ReturnType<typeof useUpdateLinkMutation>);
    createLink.mockImplementation(
      (
        input: { storyId: string; title?: string; url: string },
        options: {
          onSuccess: (response: { data: { id: string } }) => void;
        },
      ) => {
        options.onSuccess({ data: { id: "fallback-link-1" } });
      },
    );
  });

  it("creates the preview automatically and removes the fallback link", async () => {
    renderDialog();

    submitGoogleSheet();

    expect(createLink).toHaveBeenCalledWith(
      { storyId: "story-1", title: "", url: googleSheetURL },
      expect.objectContaining({
        onError: expect.any(Function),
        onSuccess: expect.any(Function),
      }),
    );
    expect(mockAttachGoogleDriveFiles).toHaveBeenCalledWith([
      {
        id: "google-sheet-1",
        mimeType: "application/vnd.google-apps.spreadsheet",
      },
    ]);
    await waitFor(() => {
      expect(deleteLink).toHaveBeenCalledWith({
        linkId: "fallback-link-1",
        storyId: "story-1",
      });
    });
    expect(screen.queryByTestId("google-drive-picker")).not.toBeInTheDocument();
  });

  it("opens Picker only when Google requires file authorization", async () => {
    mockAttachGoogleDriveFiles.mockRejectedValue(
      Object.assign(new Error("Google denied access to this file"), {
        code: "permission_denied",
      }),
    );
    renderDialog();

    submitGoogleSheet();

    await waitFor(() => {
      expect(screen.getByTestId("google-drive-picker")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText("Cancel picker"));

    expect(screen.queryByTestId("google-drive-picker")).not.toBeInTheDocument();
    expect(deleteLink).not.toHaveBeenCalled();
  });

  it("removes the fallback link after an authorized Picker attachment", async () => {
    mockAttachGoogleDriveFiles.mockRejectedValue(
      Object.assign(new Error("Google denied access to this file"), {
        code: "permission_denied",
      }),
    );
    renderDialog();

    submitGoogleSheet();
    await waitFor(() => {
      expect(screen.getByTestId("google-drive-picker")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("Attach file"));

    expect(deleteLink).toHaveBeenCalledWith({
      linkId: "fallback-link-1",
      storyId: "story-1",
    });
  });
});
