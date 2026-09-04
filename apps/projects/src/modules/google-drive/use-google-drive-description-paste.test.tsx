import type { Editor } from "@tiptap/core";
import { act, isValidElement } from "react";
import type { ClipboardEvent } from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import type { GoogleDriveFileReference } from "@/shared/google-drive/types";
import {
  useAttachGoogleDriveFiles,
  useCreateGoogleDriveConnectSession,
  useGoogleDriveIntegration,
} from "./hooks";
import { replacePastedGoogleDriveURLLabel } from "./google-drive-description-link";
import type { GoogleDrivePickerDialogProps } from "./google-drive-picker-dialog";
import { useGoogleDriveDescriptionPaste } from "./use-google-drive-description-paste";

jest.mock("sonner", () => ({
  toast: {
    dismiss: jest.fn(),
    error: jest.fn(),
    info: jest.fn(() => "drive-connect-prompt"),
    loading: jest.fn(() => "drive-preview-loading"),
    success: jest.fn(),
  },
}));

jest.mock("./hooks", () => ({
  useAttachGoogleDriveFiles: jest.fn(),
  useCreateGoogleDriveConnectSession: jest.fn(),
  useGoogleDriveIntegration: jest.fn(),
}));
jest.mock("@/shared/google-drive/google-drive-icon", () => ({
  GoogleDriveIcon: () => null,
}));
jest.mock("./google-drive-picker-dialog", () => ({
  GoogleDrivePickerDialog: () => null,
}));

jest.mock("./google-drive-description-link", () => ({
  replacePastedGoogleDriveURLLabel: jest.fn(() => true),
}));

const RAW_URL =
  "https://docs.google.com/spreadsheets/d/google-sheet-1/edit?resourcekey=resource-key-1";
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

const createEditor = ({ code = false } = {}) =>
  ({
    isActive: jest.fn((name: string) => code && name === "code"),
    state: { selection: { from: 17 } },
  }) as unknown as Editor;

const paste = (
  onPaste: (event: ClipboardEvent<HTMLDivElement>) => void,
  value: string,
) => {
  act(() => {
    onPaste({
      clipboardData: { getData: () => value },
    } as unknown as ClipboardEvent<HTMLDivElement>);
  });
};

describe("useGoogleDriveDescriptionPaste", () => {
  const attachFiles = jest.fn();
  const connect = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
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
    } as ReturnType<typeof useGoogleDriveIntegration>);
  });

  it("supports Google context on document editors", () => {
    renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "document-1", type: "document" },
      }),
    );

    expect(useAttachGoogleDriveFiles).toHaveBeenCalledWith(
      { id: "document-1", type: "document" },
      { notifyOnError: false, notifyOnSuccess: false },
    );
  });

  it("attaches an authorized standalone URL and upgrades its pasted label", async () => {
    attachFiles.mockResolvedValue([file]);
    const editor = createEditor();
    const { result } = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor,
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(result.current.onPaste, RAW_URL);

    await waitFor(() => {
      expect(attachFiles).toHaveBeenCalledWith([
        {
          id: "google-sheet-1",
          mimeType: "application/vnd.google-apps.spreadsheet",
          resourceKey: "resource-key-1",
        },
      ]);
    });
    expect(replacePastedGoogleDriveURLLabel).toHaveBeenCalledWith({
      approximatePosition: 17,
      editor,
      file,
      rawURL: RAW_URL,
    });
    expect(toast.success).toHaveBeenCalledWith("Google file preview ready", {
      description: "Linked as Launch plan.",
    });
    expect(toast.dismiss).toHaveBeenCalledWith("drive-preview-loading");
    expect(result.current.picker).toBeNull();
  });

  it("falls back to a focused Picker when Drive has not authorized the file", async () => {
    attachFiles.mockRejectedValue(
      Object.assign(new Error("Google denied access to this file"), {
        code: "permission_denied",
      }),
    );
    const { result } = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(result.current.onPaste, RAW_URL);

    await waitFor(() => {
      expect(result.current.picker).not.toBeNull();
    });
    expect(toast.info).toHaveBeenCalledWith("Authorize this Google file", {
      description:
        "Choose the pasted file once in Google Drive to create its private preview.",
      duration: 8_000,
      icon: expect.anything(),
      id: "drive-preview-loading",
    });

    const picker = result.current.picker;
    expect(isValidElement<GoogleDrivePickerDialogProps>(picker)).toBe(true);
    if (!isValidElement<GoogleDrivePickerDialogProps>(picker)) return;
    expect(picker.props.fileIds).toEqual(["google-sheet-1"]);
    expect(picker.props.target).toEqual({ id: "story-1", type: "story" });

    act(() => {
      picker.props.onClose();
    });
    expect(result.current.picker).toBeNull();
  });

  it("reports non-permission failures without opening an authorization Picker", async () => {
    attachFiles.mockRejectedValue(
      Object.assign(new Error("Google rate limit reached; try again later"), {
        code: "rate_limited",
      }),
    );
    const { result } = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(result.current.onPaste, RAW_URL);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "Couldn’t create Google Drive preview",
        {
          description: "Google rate limit reached; try again later",
          id: "drive-preview-loading",
        },
      );
    });
    expect(result.current.picker).toBeNull();
  });

  it("keeps the URL and offers connection when Drive is disconnected", () => {
    jest.mocked(useGoogleDriveIntegration).mockReturnValue({
      data: {
        configured: true,
        connected: false,
        requiresReauthorization: false,
        status: "disconnected",
      },
    } as ReturnType<typeof useGoogleDriveIntegration>);
    const { result } = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(result.current.onPaste, RAW_URL);

    expect(attachFiles).not.toHaveBeenCalled();
    expect(toast.info).toHaveBeenCalledWith(
      "Get more context from this Google file",
      expect.objectContaining({
        description:
          "Connect Google Drive to add a private preview. The pasted link will stay in the description.",
        duration: 10_000,
      }),
    );
    const options = jest.mocked(toast.info).mock.calls[0]?.[1] as unknown as {
      action: { onClick: (event: { preventDefault: () => void }) => void };
      cancel: { onClick: () => void };
    };
    const preventDefault = jest.fn();
    act(() => {
      options.action.onClick({ preventDefault });
    });
    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(connect).toHaveBeenCalledWith(window.location.href);

    act(() => {
      options.cancel.onClick();
    });
    expect(toast.dismiss).toHaveBeenCalledWith("drive-connect-prompt");
  });

  it("shows the connection nudge only once per editor session", () => {
    jest.mocked(useGoogleDriveIntegration).mockReturnValue({
      data: {
        configured: true,
        connected: false,
        requiresReauthorization: false,
        status: "disconnected",
      },
    } as ReturnType<typeof useGoogleDriveIntegration>);
    const { result } = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(result.current.onPaste, RAW_URL);
    paste(result.current.onPaste, RAW_URL);

    expect(toast.info).toHaveBeenCalledTimes(1);
    expect(attachFiles).not.toHaveBeenCalled();
  });

  it("does not offer OAuth when Drive is not configured", () => {
    jest.mocked(useGoogleDriveIntegration).mockReturnValue({
      data: {
        configured: false,
        connected: false,
        requiresReauthorization: false,
        status: "disconnected",
      },
    } as ReturnType<typeof useGoogleDriveIntegration>);
    const { result } = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(result.current.onPaste, RAW_URL);

    expect(toast.error).toHaveBeenCalledWith(
      "Google Drive isn’t available yet",
      expect.objectContaining({
        description: expect.stringContaining(
          "Contact your administrator if Google Drive should be enabled",
        ),
      }),
    );
    expect(toast.info).not.toHaveBeenCalled();
    expect(connect).not.toHaveBeenCalled();
    expect(attachFiles).not.toHaveBeenCalled();
  });

  it("ignores Drive URLs pasted inside code and non-standalone text", () => {
    const codeHook = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor({ code: true }),
        target: { id: "story-1", type: "story" },
      }),
    );
    const textHook = renderHook(() =>
      useGoogleDriveDescriptionPaste({
        editor: createEditor(),
        target: { id: "story-1", type: "story" },
      }),
    );

    paste(codeHook.result.current.onPaste, RAW_URL);
    paste(textHook.result.current.onPaste, `Review ${RAW_URL}`);

    expect(attachFiles).not.toHaveBeenCalled();
    expect(toast.loading).not.toHaveBeenCalled();
  });
});
