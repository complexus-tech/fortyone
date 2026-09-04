import type { Editor } from "@tiptap/core";
import type { ClipboardEventHandler } from "react";
import { useRouter } from "next/navigation";
import { useCallback, useRef } from "react";
import { toast } from "sonner";
import { useUserRole, useWorkspacePath } from "@/hooks";
import {
  useCreateFigmaInstallSession,
  useFigmaIntegration,
  useLinkFigmaStory,
} from "@/lib/hooks/figma";
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
  const router = useRouter();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const { data: integration } = useFigmaIntegration();
  const { mutate: connectFigma } = useCreateFigmaInstallSession();
  const { mutateAsync: linkFigmaStory } = useLinkFigmaStory();
  const connectionNudgeShown = useRef(false);
  const figmaSettingsHref = withWorkspace(
    "/settings/workspace/integrations/figma",
  );

  const showConnectionNudge = useCallback(() => {
    if (userRole !== "admin" || connectionNudgeShown.current) return;

    connectionNudgeShown.current = true;
    const isConfigured = integration?.configured === true;
    const promptId = toast.info("Connect Figma to preview this design", {
      action: {
        label: isConfigured ? "Connect Figma" : "Open settings",
        onClick: (actionEvent) => {
          actionEvent.preventDefault();
          toast.dismiss(promptId);
          if (isConfigured) {
            connectFigma();
            return;
          }
          router.push(figmaSettingsHref);
        },
      },
      cancel: {
        label: "Keep link",
        onClick: () => {
          toast.dismiss(promptId);
        },
      },
      description: isConfigured
        ? "The pasted link was kept. Connect the workspace to create its preview."
        : "The pasted link was kept. Open Figma settings to finish the workspace setup.",
      duration: 10_000,
      icon: <FigmaIcon className="h-4 w-auto" />,
    });
  }, [
    connectFigma,
    figmaSettingsHref,
    integration?.configured,
    router,
    userRole,
  ]);

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
        const result = await linkFigmaStory({ storyId, url: rawURL });
        if (result.kind === "generic") {
          toast.success("Figma link saved", {
            description: "Saved as a normal link without a preview.",
            id: toastId,
          });
          return;
        }

        const link = result.link;
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

      if (!integration) return;
      if (!integration.connection?.isActive) {
        showConnectionNudge();
        return;
      }

      const approximatePosition = editor.state.selection.from;
      const promptId = toast.info("Figma link detected", {
        action: {
          label: "Attach",
          onClick: (actionEvent) => {
            actionEvent.preventDefault();
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
    [attachFigmaDesign, editor, integration, showConnectionNudge],
  );
};
