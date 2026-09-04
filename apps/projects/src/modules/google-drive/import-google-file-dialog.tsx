"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Select, Text } from "ui";
import { LockKeyholeIcon, UserMultiple02Icon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { useImportGoogleDriveFile } from "./hooks";
import { GoogleFileTypeIcon } from "./google-file-type-icon";
import type {
  GoogleDriveFileReference,
  GoogleDriveImportVisibility,
} from "./types";

export const ImportGoogleFileDialog = ({
  file,
  onOpenChange,
}: {
  file: GoogleDriveFileReference | null;
  onOpenChange: (isOpen: boolean) => void;
}) => {
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const [visibility, setVisibility] =
    useState<GoogleDriveImportVisibility>("private");
  const idempotencyKeyRef = useRef<string | null>(null);
  const importFile = useImportGoogleDriveFile();
  const referenceId = file?.id;

  useEffect(() => {
    if (referenceId) {
      idempotencyKeyRef.current = null;
      setVisibility("private");
    }
  }, [referenceId]);

  const handleImport = () => {
    if (!file) return;
    const idempotencyKey = idempotencyKeyRef.current ?? crypto.randomUUID();
    idempotencyKeyRef.current = idempotencyKey;
    importFile.mutate(
      { idempotencyKey, referenceId: file.id, visibility },
      {
        onSuccess: ({ documentId }) => {
          idempotencyKeyRef.current = null;
          onOpenChange(false);
          toast.success("Google Doc imported", {
            description:
              "This is a one-time snapshot. Future Google changes will not sync.",
            action: {
              label: "Open",
              onClick: () => {
                router.push(withWorkspace(`/docs/${documentId}`));
              },
            },
          });
        },
      },
    );
  };

  return (
    <Dialog
      onOpenChange={(isOpen) => {
        if (!importFile.isPending) onOpenChange(isOpen);
      }}
      open={Boolean(file)}
    >
      <Dialog.Content size="sm">
        <Dialog.Header className="px-6 pt-6 pb-2">
          <Dialog.Title className="text-lg">
            Import as a FortyOne document
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Description>
          Create an editable snapshot of this Google Doc in FortyOne.
        </Dialog.Description>
        <Dialog.Body className="space-y-5 pt-3">
          {file ? (
            <Flex
              align="center"
              className="border-border rounded-xl border p-3"
              gap={3}
            >
              <GoogleFileTypeIcon mimeType={file.mimeType} />
              <Box className="min-w-0">
                <Text className="truncate" fontWeight="semibold">
                  {file.name}
                </Text>
                {file.connectionEmail ? (
                  <Text color="muted">{file.connectionEmail}</Text>
                ) : null}
              </Box>
            </Flex>
          ) : null}
          <Box>
            <Text className="mb-1.5" color="muted">
              Who can access the imported document?
            </Text>
            <Select
              onValueChange={(value) => {
                idempotencyKeyRef.current = null;
                setVisibility(value as GoogleDriveImportVisibility);
              }}
              value={visibility}
            >
              <Select.Trigger aria-label="Imported document visibility">
                <Select.Input />
              </Select.Trigger>
              <Select.Content>
                <Select.Option className="text-base" value="private">
                  <LockKeyholeIcon className="size-4" /> Private
                </Select.Option>
                <Select.Option className="text-base" value="workspace">
                  <UserMultiple02Icon className="size-4" /> Workspace
                </Select.Option>
              </Select.Content>
            </Select>
          </Box>
          <Box className="bg-surface-muted rounded-xl p-3">
            <Text fontWeight="semibold">One-time snapshot</Text>
            <Text className="mt-1 leading-relaxed" color="muted">
              Edits made in Google after import will not update the FortyOne
              document. Your source file stays unchanged.
            </Text>
          </Box>
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-2">
          <Button
            color="tertiary"
            disabled={importFile.isPending}
            onClick={() => {
              onOpenChange(false);
            }}
          >
            Cancel
          </Button>
          <Button
            color="invert"
            loading={importFile.isPending}
            onClick={handleImport}
          >
            Import snapshot
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
