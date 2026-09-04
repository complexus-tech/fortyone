/* eslint-disable @next/next/no-img-element -- authenticated previews are streamed by a no-store application route */
"use client";

import { Badge, Box, Flex, Menu, Text, TimeAgo } from "ui";
import {
  AiIcon,
  DeleteIcon,
  DownloadIcon,
  ExternalLinkIcon,
  MoreVerticalIcon,
  RefreshIcon,
} from "icons";
import { useEffect, useState } from "react";
import { useChatContext } from "@/context/chat-context";
import { useWorkspacePath } from "@/hooks";
import {
  canMayaReadGoogleDriveFile,
  isTrustedGoogleDriveWebViewLink,
} from "./capabilities";
import { useDeleteGoogleDriveFile, useRefreshGoogleDriveFile } from "./hooks";
import { GoogleDrivePickerButton } from "./google-drive-picker-button";
import { parseGoogleDriveURL } from "./google-drive-url";
import {
  getGoogleFileTypeLabel,
  GoogleFileTypeIcon,
  hasNativeGoogleWorkspaceIcon,
  isGoogleDocsMimeType,
} from "./google-file-type-icon";
import type {
  GoogleDriveFileAvailability,
  GoogleDriveFileReference,
  GoogleDriveTarget,
} from "./types";

const availabilityPresentation: Record<
  Exclude<GoogleDriveFileAvailability, "available">,
  { color: "secondary" | "warning"; label: string }
> = {
  access_required: { color: "warning", label: "Access needed" },
  deleted: { color: "secondary", label: "Unavailable" },
  reauthorization_required: { color: "warning", label: "Reconnect" },
};

export const GoogleDriveFileCard = ({
  canEdit,
  canUseGoogleContent,
  file,
  onImport,
  onPreviewError,
  target,
}: {
  canEdit: boolean;
  canUseGoogleContent: boolean;
  file: GoogleDriveFileReference;
  onImport: (file: GoogleDriveFileReference) => void;
  onPreviewError: () => void;
  target: GoogleDriveTarget;
}) => {
  const { openChatWithGoogleDriveFile } = useChatContext();
  const { workspaceSlug } = useWorkspacePath();
  const deleteFile = useDeleteGoogleDriveFile(target);
  const refreshFile = useRefreshGoogleDriveFile(target);
  const [previewFailed, setPreviewFailed] = useState(false);
  const status =
    file.availability === "available"
      ? null
      : availabilityPresentation[file.availability];
  const canOpen =
    file.availability !== "deleted" &&
    isTrustedGoogleDriveWebViewLink(file.webViewLink);
  const isGoogleDoc = isGoogleDocsMimeType(file.mimeType);
  const canConvert =
    canEdit &&
    canUseGoogleContent &&
    file.availability === "available" &&
    isGoogleDoc;
  const canAskMaya =
    canUseGoogleContent &&
    file.availability === "available" &&
    canMayaReadGoogleDriveFile(file.mimeType);
  const parsedFileURL = parseGoogleDriveURL(file.webViewLink);
  const canAuthorizePreview = Boolean(
    canEdit &&
      canUseGoogleContent &&
      file.availability === "access_required" &&
      parsedFileURL,
  );
  const canLoadPreview = Boolean(
    workspaceSlug &&
      canUseGoogleContent &&
      file.availability === "available" &&
      !previewFailed,
  );
  const previewURL = canLoadPreview
    ? `/api/google-drive/${encodeURIComponent(workspaceSlug)}/files/${encodeURIComponent(file.id)}/preview?v=${encodeURIComponent(file.updatedAt)}`
    : null;
  const metadata = [
    hasNativeGoogleWorkspaceIcon(file.mimeType)
      ? null
      : getGoogleFileTypeLabel(file.mimeType),
    file.connectionEmail,
  ].filter((value): value is string => Boolean(value));

  useEffect(() => {
    setPreviewFailed(false);
  }, [file.id, file.updatedAt]);

  return (
    <Box className="border-border bg-surface group min-w-0 overflow-hidden rounded-xl border">
      {previewURL ? (
        <a
          aria-label={`Open ${file.name} in Google`}
          className="bg-surface-muted block aspect-video overflow-hidden"
          href={file.webViewLink}
          rel="noopener noreferrer"
          target="_blank"
        >
          <img
            alt=""
            className="size-full bg-white object-contain object-top transition-opacity group-hover:opacity-95"
            loading="lazy"
            onError={() => {
              setPreviewFailed(true);
              onPreviewError();
            }}
            src={previewURL}
          />
        </a>
      ) : (
        <Flex
          align="center"
          className="bg-surface-muted/40 aspect-video flex-col px-4 text-center"
          gap={2}
          justify="center"
        >
          <GoogleFileTypeIcon className="size-12" mimeType={file.mimeType} />
          {file.availability === "access_required" ? (
            <>
              <Text fontWeight="medium">Preview needs your Google access</Text>
              {canAuthorizePreview && parsedFileURL ? (
                <GoogleDrivePickerButton
                  fileIds={[parsedFileURL.fileId]}
                  size="sm"
                  target={target}
                  variant="outline"
                >
                  Authorize preview
                </GoogleDrivePickerButton>
              ) : null}
            </>
          ) : null}
          {file.availability !== "access_required" && previewFailed ? (
            <Text color="muted">Thumbnail unavailable</Text>
          ) : null}
        </Flex>
      )}
      <Flex align="center" className="gap-2 px-3 py-3">
        <GoogleFileTypeIcon className="size-10" mimeType={file.mimeType} />
        <Box className="min-w-0 flex-1">
          <Flex align="center" className="min-w-0 gap-1">
            {canOpen ? (
              <a
                className="min-w-0 truncate font-medium hover:underline"
                href={file.webViewLink}
                rel="noopener noreferrer"
                target="_blank"
                title={file.name}
              >
                {file.name}
              </a>
            ) : (
              <Text className="min-w-0 truncate font-medium" title={file.name}>
                {file.name}
              </Text>
            )}
            <Menu>
              <Menu.Button>
                <button
                  aria-label={`Actions for ${file.name}`}
                  className="text-text-muted hover:bg-state-hover hover:text-text-primary focus-visible:ring-ring flex size-6 shrink-0 items-center justify-center rounded-md p-0 transition-colors outline-none focus-visible:ring-2"
                  type="button"
                >
                  <MoreVerticalIcon className="size-4 text-current" />
                </button>
              </Menu.Button>
              <Menu.Items align="end" className="min-w-56">
                <Menu.Group>
                  <Menu.Item
                    disabled={!canOpen}
                    onSelect={() => {
                      if (canOpen) {
                        window.open(
                          file.webViewLink,
                          "_blank",
                          "noopener,noreferrer",
                        );
                      }
                    }}
                  >
                    <ExternalLinkIcon /> Open in Google
                  </Menu.Item>
                  {canAskMaya ? (
                    <Menu.Item
                      onSelect={() => {
                        openChatWithGoogleDriveFile(file);
                      }}
                    >
                      <AiIcon /> Ask Maya
                    </Menu.Item>
                  ) : null}
                  <Menu.Item
                    disabled={
                      !canEdit ||
                      !canUseGoogleContent ||
                      file.availability !== "available" ||
                      refreshFile.isPending
                    }
                    onSelect={() => {
                      refreshFile.mutate(file.id, {
                        onSuccess: () => {
                          setPreviewFailed(false);
                        },
                      });
                    }}
                  >
                    <RefreshIcon /> Refresh preview
                  </Menu.Item>
                  {isGoogleDoc ? (
                    <Menu.Item
                      disabled={!canConvert}
                      onSelect={() => {
                        onImport(file);
                      }}
                    >
                      <DownloadIcon /> Convert
                    </Menu.Item>
                  ) : null}
                </Menu.Group>
                {canEdit ? (
                  <>
                    <Menu.Separator />
                    <Menu.Group>
                      <Menu.Item
                        className="text-danger dark:!text-danger"
                        disabled={deleteFile.isPending}
                        onSelect={() => {
                          deleteFile.mutate(file.id);
                        }}
                      >
                        <DeleteIcon className="text-danger dark:!text-danger" />
                        Remove attachment
                      </Menu.Item>
                    </Menu.Group>
                  </>
                ) : null}
              </Menu.Items>
            </Menu>
            {status ? (
              <Badge
                className="ml-auto shrink-0"
                color={status.color}
                variant="outline"
              >
                {status.label}
              </Badge>
            ) : null}
          </Flex>
          <Flex
            align="center"
            className="mt-0.5 min-w-0 gap-3"
            justify="between"
          >
            {metadata.length > 0 ? (
              <Text className="min-w-0 flex-1 truncate" color="muted">
                {metadata.join(" · ")}
              </Text>
            ) : null}
            <Flex align="center" className="ml-auto shrink-0 gap-1.5">
              {metadata.length > 0 ? (
                <span aria-hidden="true" className="text-text-muted">
                  ·
                </span>
              ) : null}
              <Text className="whitespace-nowrap" color="muted">
                <TimeAgo timestamp={file.modifiedTime ?? file.updatedAt} />
              </Text>
            </Flex>
          </Flex>
        </Box>
      </Flex>
    </Box>
  );
};
