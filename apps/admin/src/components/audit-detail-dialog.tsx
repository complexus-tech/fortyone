"use client";

import { useState } from "react";
import { InfoIcon } from "icons";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { formatAuditValue, formatDateTime, humanizeKey } from "@/lib/format";
import type { AuditLog } from "@/lib/types";

const formatMetadata = (value: unknown) => {
  if (!value || typeof value !== "object") {
    return "None";
  }

  return JSON.stringify(value, null, 2);
};

export const AuditDetailDialog = ({ entry }: { entry: AuditLog }) => {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        asIcon
        color="tertiary"
        onClick={() => {
          setOpen(true);
        }}
        size="sm"
        type="button"
        variant="naked"
      >
        <InfoIcon className="h-4" />
      </Button>
      <Dialog onOpenChange={setOpen} open={open}>
        <Dialog.Content size="md">
          <Dialog.Header className="px-6">
            <Dialog.Title>
              <Text as="span" className="text-[1.35rem]" fontWeight="semibold">
                Audit entry
              </Text>
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Box className="space-y-4">
              <DetailRow label="Actor" value={entry.actorName || "Unknown"} />
              <DetailRow label="Action" value={humanizeKey(entry.action)} />
              <DetailRow label="Target" value={entry.targetType} />
              <DetailRow
                label="Workspace"
                value={entry.workspaceName || entry.workspaceId || "None"}
              />
              <DetailRow
                label="Field"
                value={entry.fieldName ? humanizeKey(entry.fieldName) : "None"}
              />
              <DetailRow
                label="From"
                value={formatAuditValue(entry.oldValue)}
              />
              <DetailRow label="To" value={formatAuditValue(entry.newValue)} />
              <DetailRow
                label="Reason"
                value={entry.reason || "No reason provided"}
              />
              <DetailRow label="Time" value={formatDateTime(entry.createdAt)} />
              <Box>
                <Text className="text-[0.95rem]" color="muted">
                  Metadata
                </Text>
                <pre className="border-border bg-surface-muted/65 mt-2 max-h-52 overflow-auto rounded-lg border-[0.5px] p-3 text-[0.95rem] leading-6">
                  {formatMetadata(entry.metadata)}
                </pre>
              </Box>
            </Box>
          </Dialog.Body>
          <Dialog.Footer justify="end">
            <Button
              color="tertiary"
              onClick={() => {
                setOpen(false);
              }}
              type="button"
              variant="naked"
            >
              Close
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

const DetailRow = ({ label, value }: { label: string; value: string }) => (
  <Flex className="gap-4" justify="between">
    <Text className="text-[0.95rem]" color="muted">
      {label}
    </Text>
    <Text className="max-w-[70%] min-w-0 truncate text-right" title={value}>
      {value}
    </Text>
  </Flex>
);
