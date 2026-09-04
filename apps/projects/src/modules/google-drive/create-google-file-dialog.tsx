"use client";

import type { FormEvent } from "react";
import { useRef, useState } from "react";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";
import { cn } from "lib";
import {
  GoogleDocsIcon,
  GoogleSheetsIcon,
} from "@/shared/google-drive/google-workspace-file-icons";
import type {
  GoogleDriveFileType,
  GoogleDriveTarget,
} from "@/shared/google-drive/types";
import { useCreateGoogleDriveFile } from "./hooks";

const fileTypes: {
  description: string;
  label: string;
  value: GoogleDriveFileType;
}[] = [
  {
    description: "Draft briefs, specs, notes, and decisions.",
    label: "Google Doc",
    value: "document",
  },
  {
    description: "Track structured work, plans, and exports.",
    label: "Google Sheet",
    value: "spreadsheet",
  },
];

export const CreateGoogleFileDialog = ({
  initialFileType,
  isOpen,
  onOpenChange,
  suggestedTitle,
  target,
}: {
  initialFileType: GoogleDriveFileType;
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  suggestedTitle?: string;
  target: GoogleDriveTarget;
}) => {
  const [fileType, setFileType] = useState(initialFileType);
  const [title, setTitle] = useState(suggestedTitle ?? "");
  const idempotencyKeyRef = useRef<string | null>(null);
  const createFile = useCreateGoogleDriveFile(target);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedTitle = title.trim();
    if (!trimmedTitle) return;
    const idempotencyKey = idempotencyKeyRef.current ?? crypto.randomUUID();
    idempotencyKeyRef.current = idempotencyKey;
    createFile.mutate(
      {
        fileType,
        idempotencyKey,
        title: trimmedTitle,
      },
      {
        onSuccess: () => {
          onOpenChange(false);
        },
      },
    );
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content size="md">
        <form onSubmit={handleSubmit}>
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-lg">Create Google file</Dialog.Title>
          </Dialog.Header>
          <Dialog.Description>
            Create a file in your connected Google account and attach it here.
          </Dialog.Description>
          <Dialog.Body className="space-y-5 pt-3">
            <Input
              aria-label="Google file name"
              autoFocus
              label="File name"
              maxLength={200}
              onChange={(event) => {
                idempotencyKeyRef.current = null;
                setTitle(event.target.value);
              }}
              placeholder={
                fileType === "document" ? "Project brief" : "Project plan"
              }
              value={title}
            />
            <fieldset>
              <legend className="text-text-muted mb-2">File type</legend>
              <Box className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                {fileTypes.map((item) => {
                  const isSelected = fileType === item.value;
                  return (
                    <button
                      aria-pressed={isSelected}
                      className={cn(
                        "border-border hover:border-border-strong rounded-xl border p-3 text-left transition-colors",
                        isSelected &&
                          "border-primary bg-primary/5 ring-primary/20 ring-2",
                      )}
                      key={item.value}
                      onClick={() => {
                        if (item.value !== fileType) {
                          idempotencyKeyRef.current = null;
                        }
                        setFileType(item.value);
                      }}
                      type="button"
                    >
                      <Flex align="center" gap={3}>
                        {item.value === "document" ? (
                          <GoogleDocsIcon className="size-9" />
                        ) : (
                          <GoogleSheetsIcon className="size-9" />
                        )}
                        <Text className="leading-snug" color="muted">
                          <span className="sr-only">{item.label}. </span>
                          {item.description}
                        </Text>
                      </Flex>
                    </button>
                  );
                })}
              </Box>
            </fieldset>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              disabled={createFile.isPending}
              onClick={() => {
                onOpenChange(false);
              }}
              type="button"
            >
              Cancel
            </Button>
            <Button
              color="invert"
              disabled={!title.trim()}
              loading={createFile.isPending}
              type="submit"
            >
              Create {fileType === "document" ? "Doc" : "Sheet"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
