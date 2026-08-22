/* eslint-disable no-undef -- Jest globals are provided by the Projects test environment. */

import { Editor } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ClipboardEvent } from "react";
import { toast } from "sonner";
import { useLinkFigmaStory } from "@/lib/hooks/figma";
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
  useLinkFigmaStory: jest.fn(),
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

describe("useFigmaDescriptionPaste", () => {
  it("keeps the action toast alive while the design is attached", async () => {
    const linkFigmaStory = jest.fn().mockResolvedValue({ artifact });
    (useLinkFigmaStory as jest.Mock).mockReturnValue({
      mutateAsync: linkFigmaStory,
    });
    const editor = new Editor({
      content: "<p></p>",
      extensions: [Document, Paragraph, Text],
    });
    const { result } = renderHook(() =>
      useFigmaDescriptionPaste({ editor, storyId: "story-1" }),
    );

    act(() => {
      result.current({
        clipboardData: {
          getData: () => RAW_URL,
        },
      } as unknown as ClipboardEvent<HTMLDivElement>);
    });

    const promptOptions = (toast.info as jest.Mock).mock.calls[0][1];
    const preventDefault = jest.fn();

    act(() => {
      promptOptions.action.onClick({ preventDefault });
    });

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(toast.loading).toHaveBeenCalledWith("Attaching Figma design...", {
      id: "figma-prompt",
    });
    await waitFor(() => {
      expect(linkFigmaStory).toHaveBeenCalledWith({
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
});
