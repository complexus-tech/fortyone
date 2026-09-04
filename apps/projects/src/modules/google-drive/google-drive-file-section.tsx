"use client";

import { useState } from "react";
import { Box, Button, Flex, Menu, Skeleton, Text } from "ui";
import { PlusIcon, ReloadIcon } from "icons";
import { cn } from "lib";
import {
  useCreateGoogleDriveConnectSession,
  useGoogleDriveFiles,
  useGoogleDriveIntegration,
} from "./hooks";
import { CreateGoogleFileDialog } from "./create-google-file-dialog";
import { GoogleDriveFileCard } from "./google-drive-file-card";
import { GoogleDriveIcon } from "./google-drive-icon";
import { GoogleDrivePickerButton } from "./google-drive-picker-button";
import {
  GoogleDocsIcon,
  GoogleSheetsIcon,
} from "./google-workspace-file-icons";
import { ImportGoogleFileDialog } from "./import-google-file-dialog";
import type {
  GoogleDriveFileReference,
  GoogleDriveFileType,
  GoogleDriveTarget,
} from "./types";

const GoogleDriveCreateMenu = ({
  disabled,
  onSelect,
}: {
  disabled: boolean;
  onSelect: (fileType: GoogleDriveFileType) => void;
}) => (
  <Menu>
    <Menu.Button>
      <Button
        color="tertiary"
        disabled={disabled}
        leftIcon={<PlusIcon />}
        size="sm"
        variant="naked"
      >
        Create Google file
      </Button>
    </Menu.Button>
    <Menu.Items align="end" className="min-w-48">
      <Menu.Group>
        <Menu.Item
          onSelect={() => {
            onSelect("document");
          }}
        >
          <GoogleDocsIcon />
          Google Doc
        </Menu.Item>
        <Menu.Item
          onSelect={() => {
            onSelect("spreadsheet");
          }}
        >
          <GoogleSheetsIcon />
          Google Sheet
        </Menu.Item>
      </Menu.Group>
    </Menu.Items>
  </Menu>
);

export const GoogleDriveFileSection = ({
  canEdit,
  className,
  suggestedTitle,
  target,
}: {
  canEdit: boolean;
  className?: string;
  suggestedTitle?: string;
  target: GoogleDriveTarget;
}) => {
  const integration = useGoogleDriveIntegration();
  const filesQuery = useGoogleDriveFiles(target);
  const connect = useCreateGoogleDriveConnectSession();
  const [createType, setCreateType] = useState<GoogleDriveFileType | null>(
    null,
  );
  const [importFile, setImportFile] = useState<GoogleDriveFileReference | null>(
    null,
  );
  const files = filesQuery.data ?? [];
  const isConnected = Boolean(integration.data?.connected);
  const needsReauthorization = Boolean(
    integration.data?.requiresReauthorization,
  );
  const canUseGoogleContent = Boolean(
    isConnected && !needsReauthorization && integration.data?.configured,
  );

  if (files.length === 0) return null;

  const connectLabel =
    needsReauthorization || isConnected ? "Reconnect" : "Connect Drive";
  const startConnection = () => {
    connect.mutate(window.location.href);
  };
  const connectionUnavailable =
    integration.isPending || integration.data?.configured === false;
  const openCreateDialog = (fileType: GoogleDriveFileType) => {
    setCreateType(fileType);
  };

  return (
    <Box className={cn("mt-4", className)}>
      <Flex
        align="center"
        className="border-border min-h-10 border-b-[0.5px] pb-2"
        justify="between"
      >
        <Flex align="center" gap={2}>
          <GoogleDriveIcon className="size-5" />
          <Text fontWeight="semibold">Context from Google</Text>
          <Text color="muted">{files.length}</Text>
        </Flex>
        {canEdit ? (
          <Flex align="center" gap={1}>
            {canUseGoogleContent ? (
              <>
                <GoogleDrivePickerButton
                  color="tertiary"
                  disabled={connectionUnavailable}
                  size="sm"
                  target={target}
                  variant="naked"
                >
                  Attach from Drive
                </GoogleDrivePickerButton>
                <GoogleDriveCreateMenu
                  disabled={connectionUnavailable}
                  onSelect={openCreateDialog}
                />
              </>
            ) : (
              <Button
                color="tertiary"
                disabled={connectionUnavailable}
                loading={connect.isPending}
                onClick={startConnection}
                size="sm"
                variant="naked"
              >
                {connectLabel}
              </Button>
            )}
          </Flex>
        ) : null}
      </Flex>

      {filesQuery.isPending ? (
        <Box
          aria-label="Loading Google Drive files"
          className="py-2"
          role="status"
        >
          <Skeleton className="h-14 w-full rounded-lg" />
        </Box>
      ) : null}
      {filesQuery.isError ? (
        <Flex align="center" className="py-4" justify="between">
          <Box>
            <Text fontWeight="semibold">Couldn&apos;t load Google files</Text>
            <Text color="muted">Try again before changing attachments.</Text>
          </Box>
          <Button
            asIcon
            color="tertiary"
            loading={filesQuery.isFetching}
            onClick={() => void filesQuery.refetch()}
            size="sm"
            variant="naked"
          >
            <ReloadIcon />
            <span className="sr-only">Reload Google Drive files</span>
          </Button>
        </Flex>
      ) : null}
      {!filesQuery.isPending && !filesQuery.isError ? (
        <Box className="mt-3 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
          {files.map((file) => (
            <GoogleDriveFileCard
              canEdit={canEdit}
              canUseGoogleContent={canUseGoogleContent}
              file={file}
              key={file.id}
              onImport={setImportFile}
              onPreviewError={() => {
                void filesQuery.refetch();
              }}
              target={target}
            />
          ))}
        </Box>
      ) : null}
      {createType ? (
        <CreateGoogleFileDialog
          initialFileType={createType}
          isOpen
          onOpenChange={(isOpen) => {
            if (!isOpen) setCreateType(null);
          }}
          suggestedTitle={suggestedTitle}
          target={target}
        />
      ) : null}
      <ImportGoogleFileDialog
        file={importFile}
        onOpenChange={(isOpen) => {
          if (!isOpen) setImportFile(null);
        }}
      />
    </Box>
  );
};
