import { act } from "react";
import { render, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { GoogleDrivePickerDialog } from "./google-drive-picker-dialog";
import {
  useAttachGoogleDriveFiles,
  useCreateGoogleDriveConnectSession,
  useCreateGoogleDrivePickerSession,
  useGoogleDriveIntegration,
} from "./hooks";
import type { GoogleDriveFileReference } from "./types";

jest.mock("./hooks", () => ({
  useAttachGoogleDriveFiles: jest.fn(),
  useCreateGoogleDriveConnectSession: jest.fn(),
  useCreateGoogleDrivePickerSession: jest.fn(),
  useGoogleDriveIntegration: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: { error: jest.fn(), info: jest.fn() },
}));

const file: GoogleDriveFileReference = {
  availability: "available",
  createdAt: "2026-09-04T08:00:00Z",
  id: "reference-1",
  mimeType: "application/vnd.google-apps.spreadsheet",
  name: "Launch plan",
  targetId: "story-1",
  targetType: "story",
  updatedAt: "2026-09-04T08:00:00Z",
  webViewLink: "https://docs.google.com/spreadsheets/d/google-sheet-1/edit",
};

describe("GoogleDrivePickerDialog", () => {
  const attachFiles = jest.fn();
  const connect = jest.fn();
  const createSession = jest.fn();
  const onAttached = jest.fn();
  const onClose = jest.fn();
  const view = {
    setEnableDrives: jest.fn(),
    setFileIds: jest.fn(),
    setIncludeFolders: jest.fn(),
    setMimeTypes: jest.fn(),
    setMode: jest.fn(),
    setSelectFolderEnabled: jest.fn(),
  };
  const builder = {
    addView: jest.fn(),
    build: jest.fn(),
    enableFeature: jest.fn(),
    setAppId: jest.fn(),
    setCallback: jest.fn(),
    setDeveloperKey: jest.fn(),
    setMaxItems: jest.fn(),
    setOAuthToken: jest.fn(),
    setOrigin: jest.fn(),
  };
  const setVisible = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    createSession.mockResolvedValue({
      accessToken: "access-token",
      apiKey: "api-key",
      appId: "app-id",
      origin: "https://app.example.com",
    });
    attachFiles.mockResolvedValue([file]);
    jest.mocked(useCreateGoogleDrivePickerSession).mockReturnValue({
      mutateAsync: createSession,
    } as unknown as ReturnType<typeof useCreateGoogleDrivePickerSession>);
    jest.mocked(useAttachGoogleDriveFiles).mockReturnValue({
      mutateAsync: attachFiles,
    } as unknown as ReturnType<typeof useAttachGoogleDriveFiles>);
    jest.mocked(useCreateGoogleDriveConnectSession).mockReturnValue({
      mutate: connect,
    } as unknown as ReturnType<typeof useCreateGoogleDriveConnectSession>);
    jest.mocked(useGoogleDriveIntegration).mockReturnValue({
      data: {
        configured: true,
        connected: true,
        requiresReauthorization: false,
        status: "connected",
      },
      isPending: false,
    } as ReturnType<typeof useGoogleDriveIntegration>);
    builder.build.mockReturnValue({ setVisible });

    Object.defineProperty(window, "google", {
      configurable: true,
      value: {
        picker: {
          Action: { CANCEL: "cancel", PICKED: "picked" },
          DocsView: jest.fn(() => view),
          DocsViewMode: { LIST: "list" },
          Feature: { MULTISELECT_ENABLED: "multiselect" },
          PickerBuilder: jest.fn(() => builder),
          ViewId: { DOCS: "docs" },
        },
      } as unknown as NonNullable<Window["google"]>,
      writable: true,
    });
  });

  it("limits a file-specific Picker to the requested Drive file", async () => {
    render(
      <GoogleDrivePickerDialog
        fileIds={["google-sheet-1"]}
        onAttached={onAttached}
        onClose={onClose}
        target={{ id: "story-1", type: "story" }}
      />,
    );

    await waitFor(() => {
      expect(setVisible).toHaveBeenCalledWith(true);
    });
    expect(view.setFileIds).toHaveBeenCalledWith("google-sheet-1");
    expect(view.setIncludeFolders).toHaveBeenCalledWith(false);
    expect(view.setEnableDrives).not.toHaveBeenCalled();
    expect(builder.enableFeature).not.toHaveBeenCalled();
    expect(builder.setMaxItems).toHaveBeenCalledWith(1);

    const pickerCallback = builder.setCallback.mock.calls[0]?.[0] as
      | ((data: {
          action: string;
          docs: { id: string; mimeType: string; resourceKey: string }[];
        }) => void)
      | undefined;
    expect(pickerCallback).toBeDefined();
    await act(async () => {
      pickerCallback?.({
        action: "picked",
        docs: [
          {
            id: "google-sheet-1",
            mimeType: "application/vnd.google-apps.spreadsheet",
            resourceKey: "resource-key-1",
          },
        ],
      });
    });

    expect(attachFiles).toHaveBeenCalledWith([
      {
        id: "google-sheet-1",
        mimeType: "application/vnd.google-apps.spreadsheet",
        resourceKey: "resource-key-1",
      },
    ]);
    expect(onAttached).toHaveBeenCalledWith([file]);
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("hides the Picker and ignores its callback after unmount", async () => {
    const { unmount } = render(
      <GoogleDrivePickerDialog
        fileIds={["google-sheet-1"]}
        onAttached={onAttached}
        onClose={onClose}
        target={{ id: "story-1", type: "story" }}
      />,
    );

    await waitFor(() => {
      expect(setVisible).toHaveBeenCalledWith(true);
    });
    const pickerCallback = builder.setCallback.mock.calls[0]?.[0] as
      | ((data: {
          action: string;
          docs: { id: string; mimeType: string }[];
        }) => void)
      | undefined;

    unmount();
    expect(setVisible).toHaveBeenLastCalledWith(false);

    act(() => {
      pickerCallback?.({
        action: "picked",
        docs: [
          {
            id: "google-sheet-1",
            mimeType: "application/vnd.google-apps.spreadsheet",
          },
        ],
      });
    });

    expect(attachFiles).not.toHaveBeenCalled();
    expect(onAttached).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("offers the current user a connection before creating a Picker session", async () => {
    jest.mocked(useGoogleDriveIntegration).mockReturnValue({
      data: {
        configured: true,
        connected: false,
        requiresReauthorization: false,
        status: "disconnected",
      },
      isPending: false,
    } as ReturnType<typeof useGoogleDriveIntegration>);

    render(
      <GoogleDrivePickerDialog
        onClose={onClose}
        target={{ id: "story-1", type: "story" }}
      />,
    );

    await waitFor(() => {
      expect(onClose).toHaveBeenCalledTimes(1);
    });
    expect(createSession).not.toHaveBeenCalled();
    expect(toast.info).toHaveBeenCalledWith(
      "Connect Google Drive to continue",
      expect.objectContaining({
        action: expect.objectContaining({ label: "Connect" }),
      }),
    );

    const options = jest.mocked(toast.info).mock.calls[0]?.[1] as unknown as {
      action: { onClick: (event: { preventDefault: () => void }) => void };
    };
    const preventDefault = jest.fn();
    act(() => {
      options.action.onClick({ preventDefault });
    });
    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(connect).toHaveBeenCalledWith(window.location.href);
  });
});
