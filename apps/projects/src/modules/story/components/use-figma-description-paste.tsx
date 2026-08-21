import type { Editor } from "@tiptap/core";
import type { ClipboardEventHandler } from "react";
import { useCallback } from "react";
import { toast } from "sonner";
import { useLinkFigmaStory } from "@/lib/hooks/figma";
import { FigmaIcon } from "@/modules/settings/workspace/integrations/figma/icon";
import {
  getFigmaLinkLabel,
  getStandaloneFigmaURL,
  replacePastedFigmaURLLabel,
} from "./figma-description-link";

export const useFigmaDescriptionPaste = ({
  editor,
  storyId,
}: {
  editor: Editor | null;
  storyId: string;
}) => {
  const { mutateAsync: linkFigmaStory } = useLinkFigmaStory();

  const attachFigmaDesign = useCallback(
    async ({
      approximatePosition,
      rawURL,
      toastId,
    }: {
      approximatePosition: number;
      rawURL: string;
      toastId: number | string;
    }) => {
      toast.loading("Attaching Figma design...", { id: toastId });

      try {
        const link = await linkFigmaStory({ storyId, url: rawURL });
        const updatedLabel = editor
          ? replacePastedFigmaURLLabel({
              approximatePosition,
              artifact: link.artifact,
              editor,
              rawURL,
            })
          : false;

        toast.success("Figma design attached", {
          description: updatedLabel
            ? `Linked as ${getFigmaLinkLabel(link.artifact)}.`
            : "The design is now available in the Design section.",
          id: toastId,
        });
      } catch (error) {
        toast.error("Figma design could not be attached", {
          description:
            error instanceof Error
              ? error.message
              : "The pasted link has been kept in the description.",
          id: toastId,
        });
      }
    },
    [editor, linkFigmaStory, storyId],
  );

  return useCallback<ClipboardEventHandler<HTMLDivElement>>(
    (event) => {
      if (!editor || editor.isActive("code") || editor.isActive("codeBlock")) {
        return;
      }

      const rawURL = getStandaloneFigmaURL(
        event.clipboardData.getData("text/plain"),
      );
      if (!rawURL) return;

      const approximatePosition = editor.state.selection.from;
      const promptId = toast.info("Figma link detected", {
        action: {
          label: "Attach",
          onClick: () => {
            void attachFigmaDesign({
              approximatePosition,
              rawURL,
              toastId: promptId,
            });
          },
        },
        cancel: {
          label: "Keep link",
          onClick: () => {
            toast.dismiss(promptId);
          },
        },
        description: "Attach this design to the story?",
        duration: 10_000,
        icon: <FigmaIcon className="h-4 w-auto" />,
      });
    },
    [attachFigmaDesign, editor],
  );
};
