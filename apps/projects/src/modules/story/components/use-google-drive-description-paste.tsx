import type { Editor } from "@tiptap/core";
import type { ClipboardEventHandler } from "react";
import { useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import {
  GoogleDriveIcon,
  GoogleDrivePickerDialog,
  getStandaloneGoogleDriveURL,
  parseGoogleDriveURL,
  useAttachGoogleDriveFiles,
  useCreateGoogleDriveConnectSession,
  useGoogleDriveIntegration,
} from "@/modules/google-drive";
import type {
  GoogleDriveFileReference,
  GoogleDriveURL,
} from "@/modules/google-drive";
import { replacePastedGoogleDriveURLLabel } from "@/modules/google-drive/google-drive-description-link";

type PendingGoogleDrivePaste = {
  approximatePosition: number;
  parsedURL: GoogleDriveURL;
  rawURL: string;
};

const getErrorDetail = (error: unknown, key: "code" | "message") => {
  if (!error || typeof error !== "object" || !(key in error)) return null;
  const value = (error as Record<string, unknown>)[key];
  return typeof value === "string" ? value : null;
};

export const useGoogleDriveDescriptionPaste = ({
  editor,
  storyId,
}: {
  editor: Editor | null;
  storyId: string;
}) => {
  const target = { id: storyId, type: "story" as const };
  const integration = useGoogleDriveIntegration();
  const connect = useCreateGoogleDriveConnectSession();
  const attachFiles = useAttachGoogleDriveFiles(target, {
    notifyOnError: false,
    notifyOnSuccess: false,
  });
  const [pendingPaste, setPendingPaste] =
    useState<PendingGoogleDrivePaste | null>(null);
  const hasShownConnectionPrompt = useRef(false);
  const hasShownConfigurationError = useRef(false);

  const showConnectionPrompt = useCallback(
    ({
      reconnect,
      toastId,
    }: {
      reconnect: boolean;
      toastId?: number | string;
    }) => {
      if (hasShownConnectionPrompt.current) {
        if (toastId !== undefined) toast.dismiss(toastId);
        return;
      }
      hasShownConnectionPrompt.current = true;

      const promptId = toast.info(
        reconnect
          ? "Reconnect Google Drive to preview this file"
          : "Connect Google Drive to preview this file",
        {
          action: {
            label: reconnect ? "Reconnect" : "Connect Drive",
            onClick: (event) => {
              event.preventDefault();
              connect.mutate(window.location.href);
            },
          },
          cancel: {
            label: "Keep link",
            onClick: () => {
              toast.dismiss(promptId);
            },
          },
          description: "The pasted link will stay in the description.",
          duration: 10_000,
          icon: <GoogleDriveIcon className="size-4" />,
          ...(toastId !== undefined ? { id: toastId } : {}),
        },
      );
    },
    [connect],
  );

  const showConfigurationError = useCallback((toastId?: number | string) => {
    if (hasShownConfigurationError.current) {
      if (toastId !== undefined) toast.dismiss(toastId);
      return;
    }
    hasShownConfigurationError.current = true;
    toast.error("Google Drive isn’t available yet", {
      description:
        "The pasted link will stay in the description. Contact your administrator if Google Drive should be enabled.",
      ...(toastId !== undefined ? { id: toastId } : {}),
    });
  }, []);

  const finishAttachment = useCallback(
    (pending: PendingGoogleDrivePaste, files: GoogleDriveFileReference[]) => {
      const file = files.find(
        (candidate) =>
          parseGoogleDriveURL(candidate.webViewLink)?.fileId ===
          pending.parsedURL.fileId,
      );
      if (!file) return;

      const updatedLabel = editor
        ? replacePastedGoogleDriveURLLabel({
            approximatePosition: pending.approximatePosition,
            editor,
            file,
            rawURL: pending.rawURL,
          })
        : false;
      toast.success("Google file preview ready", {
        description: updatedLabel
          ? `Linked as ${file.name}.`
          : "The file is now available in Google Drive previews.",
      });
    },
    [editor],
  );

  const attachDetectedLink = useCallback(
    async (pending: PendingGoogleDrivePaste) => {
      if (integration.data?.configured === false) {
        showConfigurationError();
        return;
      }
      if (
        integration.data &&
        (!integration.data.connected ||
          integration.data.requiresReauthorization)
      ) {
        showConnectionPrompt({
          reconnect:
            integration.data.connected ||
            integration.data.requiresReauthorization,
        });
        return;
      }

      const toastId = toast.loading("Creating Google Drive preview...");
      try {
        const files = await attachFiles.mutateAsync([
          {
            id: pending.parsedURL.fileId,
            ...(pending.parsedURL.mimeType
              ? { mimeType: pending.parsedURL.mimeType }
              : {}),
            ...(pending.parsedURL.resourceKey
              ? { resourceKey: pending.parsedURL.resourceKey }
              : {}),
          },
        ]);
        finishAttachment(pending, files);
        toast.dismiss(toastId);
      } catch (error) {
        const errorCode = getErrorDetail(error, "code");
        const errorMessage = getErrorDetail(error, "message");
        if (errorCode === "permission_denied") {
          toast.info("Authorize this Google file", {
            description:
              "Choose the pasted file once in Google Drive to create its private preview.",
            duration: 8_000,
            icon: <GoogleDriveIcon className="size-4" />,
            id: toastId,
          });
          setPendingPaste(pending);
          return;
        }

        if (
          errorCode === "conflict" &&
          errorMessage &&
          /connect Google Drive|reconnected/i.test(errorMessage)
        ) {
          showConnectionPrompt({
            reconnect: /reconnected/i.test(errorMessage),
            toastId,
          });
          return;
        }

        if (errorCode === "service_unavailable") {
          toast.error("Google Drive preview unavailable", {
            description:
              "The pasted link will stay in the description. Try again later.",
            id: toastId,
          });
          return;
        }

        toast.error("Couldn’t create Google Drive preview", {
          description:
            errorMessage ?? "The pasted link was kept. Try again in a moment.",
          id: toastId,
        });
      }
    },
    [
      attachFiles,
      finishAttachment,
      integration.data,
      showConfigurationError,
      showConnectionPrompt,
    ],
  );

  const onPaste = useCallback<ClipboardEventHandler<HTMLDivElement>>(
    (event) => {
      if (!editor || editor.isActive("code") || editor.isActive("codeBlock")) {
        return;
      }

      const rawURL = getStandaloneGoogleDriveURL(
        event.clipboardData.getData("text/plain"),
      );
      if (!rawURL) return;
      const parsedURL = parseGoogleDriveURL(rawURL);
      if (!parsedURL) return;

      void attachDetectedLink({
        approximatePosition: editor.state.selection.from,
        parsedURL,
        rawURL,
      });
    },
    [attachDetectedLink, editor],
  );

  const picker = pendingPaste ? (
    <GoogleDrivePickerDialog
      fileIds={[pendingPaste.parsedURL.fileId]}
      onAttached={(files) => {
        finishAttachment(pendingPaste, files);
      }}
      onClose={() => {
        setPendingPaste(null);
      }}
      target={target}
    />
  ) : null;

  return { onPaste, picker };
};
