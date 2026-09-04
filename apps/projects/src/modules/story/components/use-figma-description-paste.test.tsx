import { Editor } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useRouter } from "next/navigation";
import type { ClipboardEvent } from "react";
import { toast } from "sonner";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import {
  useCreateFigmaInstallSession,
  useFigmaIntegration,
  useLinkFigmaStory,
} from "@/lib/hooks/figma";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";
import { useFigmaDescriptionPaste } from "./use-figma-description-paste";

jest.mock("sonner", () => ({
  toast: {
    dismiss: jest.fn(),
    error: jest.fn(),
    info: jest.fn(() => "figma-prompt"),
    loading: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("@/lib/hooks/figma", () => ({
  useCreateFigmaInstallSession: jest.fn(),
  useFigmaIntegration: jest.fn(),
  useLinkFigmaStory: jest.fn(),
}));

jest.mock("@/hooks/role", () => ({
  useUserRole: jest.fn(),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: jest.fn(),
}));

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

jest.mock("@/modules/settings/workspace/integrations/figma/icon", () => ({
  FigmaIcon: () => null,
}));

const RAW_URL = "https://www.figma.com/design/file-key/file-name?node-id=12-34";

const artifact: FigmaArtifact = {
  canonicalUrl:
    "https://www.figma.com/design/file-key/file-name?node-id=12%3A34",
  fileKey: "file-key",
  fileName: "Product foundations",
  lastModified: null,
  nodeId: "12:34",
  nodeName: "Checkout flow",
  nodeType: "FRAME",
  originalUrl: RAW_URL,
  thumbnailUrl: null,
  version: null,
};

const mockConnectFigma = jest.fn();
const mockLinkFigmaStory = jest.fn();
const mockPush = jest.fn();

const setFigmaIntegration = ({
  configured,
  connected,
}: {
  configured: boolean;
  connected: boolean;
}) => {
  jest.mocked(useFigmaIntegration).mockReturnValue({
    data: {
      configured,
      connection: connected ? { isActive: true } : null,
    },
  } as ReturnType<typeof useFigmaIntegration>);
};

const createEditor = () =>
  new Editor({
    content: "<p></p>",
    extensions: [Document, Paragraph, Text],
  });

const pasteFigmaURL = (
  onPaste: (event: ClipboardEvent<HTMLDivElement>) => void,
) => {
  const preventDefault = jest.fn();
  act(() => {
    onPaste({
      clipboardData: { getData: () => RAW_URL },
      preventDefault,
    } as unknown as ClipboardEvent<HTMLDivElement>);
  });
  return { preventDefault };
};

describe("useFigmaDescriptionPaste", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(useCreateFigmaInstallSession).mockReturnValue({
      mutate: mockConnectFigma,
    } as unknown as ReturnType<typeof useCreateFigmaInstallSession>);
    setFigmaIntegration({ configured: true, connected: true });
    jest.mocked(useLinkFigmaStory).mockReturnValue({
      mutateAsync: mockLinkFigmaStory,
    } as unknown as ReturnType<typeof useLinkFigmaStory>);
    jest.mocked(useRouter).mockReturnValue({
      push: mockPush,
    } as unknown as ReturnType<typeof useRouter>);
    jest.mocked(useUserRole).mockReturnValue({ userRole: "member" });
    jest.mocked(useWorkspacePath).mockReturnValue({
      withWorkspace: (path: string) => `/acme${path}`,
      workspaceSlug: "acme",
    });
  });

  it("keeps the action toast alive while the design is attached", async () => {
    mockLinkFigmaStory.mockResolvedValue({
      kind: "figma",
      link: { artifact },
    });
    const editor = createEditor();
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    pasteFigmaURL(result.current);

    const promptOptions = (toast.info as jest.Mock).mock.calls[0][1];
    const preventDefault = jest.fn();

    expect(promptOptions.action.label).toBe("Attach");
    expect(promptOptions.cancel.label).toBe("Keep link");
    expect(mockConnectFigma).not.toHaveBeenCalled();

    act(() => {
      promptOptions.action.onClick({ preventDefault });
    });

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(toast.loading).toHaveBeenCalledWith("Attaching Figma design...", {
      id: "figma-prompt",
    });
    await waitFor(() => {
      expect(mockLinkFigmaStory).toHaveBeenCalledWith({
        storyId: "story-1",
        url: RAW_URL,
      });
    });
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith(
        "Figma design attached",
        expect.objectContaining({ id: "figma-prompt" }),
      );
    });

    editor.destroy();
  });

  it("confirms when unavailable Figma metadata is saved as a normal link", async () => {
    mockLinkFigmaStory.mockResolvedValue({
      kind: "generic",
      link: { id: "link-1", storyId: "story-1", url: RAW_URL },
    });
    const editor = createEditor();
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    pasteFigmaURL(result.current);
    const promptOptions = (toast.info as jest.Mock).mock.calls[0][1];

    act(() => {
      promptOptions.action.onClick({ preventDefault: jest.fn() });
    });

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Figma link saved", {
        description: "Saved as a normal link without a preview.",
        id: "figma-prompt",
      });
    });
    expect(editor.getText()).toBe("");

    editor.destroy();
  });

  it("continues offering attach prompts for connected workspaces", () => {
    const editor = createEditor();
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    pasteFigmaURL(result.current);
    pasteFigmaURL(result.current);

    expect(toast.info).toHaveBeenCalledTimes(2);
    for (const [title, options] of jest.mocked(toast.info).mock.calls) {
      expect(title).toBe("Figma link detected");
      expect(options).toEqual(
        expect.objectContaining({
          action: expect.objectContaining({ label: "Attach" }),
          cancel: expect.objectContaining({ label: "Keep link" }),
        }),
      );
    }

    editor.destroy();
  });

  it("offers admins one explicit connection action when Figma is disconnected", () => {
    jest.mocked(useUserRole).mockReturnValue({ userRole: "admin" });
    setFigmaIntegration({ configured: true, connected: false });
    const editor = createEditor();
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    const firstPaste = pasteFigmaURL(result.current);
    pasteFigmaURL(result.current);

    expect(firstPaste.preventDefault).not.toHaveBeenCalled();
    expect(toast.info).toHaveBeenCalledTimes(1);
    expect(mockConnectFigma).not.toHaveBeenCalled();
    expect(mockLinkFigmaStory).not.toHaveBeenCalled();
    const [title, options] = jest.mocked(toast.info).mock
      .calls[0] as unknown as [
      string,
      {
        action: {
          label: string;
          onClick: (event: { preventDefault: () => void }) => void;
        };
        cancel: { label: string; onClick: () => void };
      },
    ];
    expect(title).toBe("Connect Figma to preview this design");
    expect(options.action.label).toBe("Connect Figma");
    expect(options.cancel.label).toBe("Keep link");

    const preventDefault = jest.fn();
    act(() => {
      options.action.onClick({ preventDefault });
    });

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(mockConnectFigma).toHaveBeenCalledTimes(1);
    expect(mockPush).not.toHaveBeenCalled();
    expect(toast.dismiss).toHaveBeenCalledWith("figma-prompt");

    editor.destroy();
  });

  it("takes admins to Figma settings when the provider is unconfigured", () => {
    jest.mocked(useUserRole).mockReturnValue({ userRole: "admin" });
    setFigmaIntegration({ configured: false, connected: false });
    const editor = createEditor();
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    pasteFigmaURL(result.current);
    const options = jest.mocked(toast.info).mock.calls[0]?.[1] as unknown as {
      action: {
        label: string;
        onClick: (event: { preventDefault: () => void }) => void;
      };
      cancel: { onClick: () => void };
    };

    expect(options.action.label).toBe("Open settings");

    act(() => {
      options.action.onClick({ preventDefault: jest.fn() });
    });

    expect(mockConnectFigma).not.toHaveBeenCalled();
    expect(mockPush).toHaveBeenCalledWith(
      "/acme/settings/workspace/integrations/figma",
    );

    act(() => {
      options.cancel.onClick();
    });
    expect(toast.dismiss).toHaveBeenCalledWith("figma-prompt");

    editor.destroy();
  });

  it("leaves disconnected members' pasted links alone without an install CTA", () => {
    setFigmaIntegration({ configured: true, connected: false });
    const editor = createEditor();
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    const paste = pasteFigmaURL(result.current);

    expect(paste.preventDefault).not.toHaveBeenCalled();
    expect(toast.info).not.toHaveBeenCalled();
    expect(mockConnectFigma).not.toHaveBeenCalled();
    expect(mockLinkFigmaStory).not.toHaveBeenCalled();

    editor.destroy();
  });
});
