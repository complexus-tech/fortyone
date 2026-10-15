"use client";

import { useId, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Box, Button, Checkbox, Dialog, Flex, Text } from "ui";
import type { DesktopOperation } from "./outbox-types";
import type { DesktopCacheScope } from "./cache-types";
import { outboxFieldValue } from "./outbox-field-labels";
import {
  refreshDesktopOutbox,
  resolveDesktopOperation,
  useDesktopOutbox,
} from "./desktop-outbox-coordinator";

const fieldLabels: Record<string, string> = {
  title: "Title",
  description: "Description",
  content: "Comment",
  comment: "Comment",
  statusId: "Status",
  priority: "Priority",
  assigneeId: "Assignee",
  teamId: "Team",
  sprintId: "Sprint",
  startDate: "Start date",
  endDate: "Due date",
  estimate: "Estimate",
  parentId: "Parent story",
  labelIds: "Labels",
  labels: "Labels",
  collaboratorIds: "Collaborators",
  objectiveId: "Objective",
  keyResultId: "Key result",
  estimateValue: "Estimate",
};
const operationName = (operation: DesktopOperation) => {
  if (operation.kind === "attachment.upload")
    return typeof operation.payload.fileName === "string"
      ? operation.payload.fileName
      : "File upload";
  if (operation.kind === "comment.create") return "New comment";
  if (typeof operation.payload.title === "string")
    return operation.payload.title;
  return operation.kind === "story.create" ? "New story" : "Story update";
};

export const DesktopOutboxStatus = () => {
  const outbox = useDesktopOutbox();
  const [open, setOpen] = useState(false);
  if (!outbox || (!outbox.snapshot.operations.length && !outbox.error))
    return null;
  const needsReview = outbox.snapshot.operations.some(
    (operation) => operation.status !== "pending",
  );
  return (
    <>
      <Flex align="center" gap={4} justify="between">
        <Box>
          <Text className="font-medium">Saved changes</Text>
          <Text color="muted">
            {needsReview || outbox.error
              ? "Some changes need your attention. Review them before choosing what to keep."
              : "Review changes that are waiting to reach your workspace."}
          </Text>
        </Box>
        <Button
          className="shrink-0"
          onClick={() => {
            setOpen(true);
          }}
          size="sm"
          variant="outline"
        >
          Review saved changes
        </Button>
      </Flex>
      <Dialog onOpenChange={setOpen} open={open}>
        <Dialog.Content className="max-h-[85vh] overflow-y-auto" size="lg">
          <Dialog.Header className="px-6 pr-12">
            <Dialog.Title>Saved changes</Dialog.Title>
            <Dialog.Description className="mt-2">
              Changes stay on this computer until FortyOne confirms them. Review
              conflicts before applying your version.
            </Dialog.Description>
          </Dialog.Header>
          <Box className="space-y-5 px-6 pb-6">
            {outbox.error ? <Text role="alert">{outbox.error}</Text> : null}
            {!outbox.online ? (
              <Text color="muted">
                Connect to synchronize these changes. Your saved work remains on
                this computer.
              </Text>
            ) : null}
            {outbox.snapshot.operations.map((operation) => (
              <OperationReview
                busy={outbox.busy}
                key={`${operation.id}:${operation.serverVersion ?? "pending"}`}
                online={outbox.online}
                operation={operation}
                scope={outbox.scope}
              />
            ))}
            <Button
              disabled={outbox.busy || !outbox.online}
              onClick={() => {
                void refreshDesktopOutbox(outbox.scope);
              }}
              variant="outline"
            >
              Retry sync
            </Button>
          </Box>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

const OperationReview = ({
  operation,
  scope,
  busy,
  online,
}: {
  operation: DesktopOperation;
  scope: DesktopCacheScope;
  busy: boolean;
  online: boolean;
}) => {
  const id = useId();
  const client = useQueryClient();
  const [reviewed, setReviewed] = useState(false);
  const [confirmDiscard, setConfirmDiscard] = useState(false);
  const [error, setError] = useState<string>();
  const conflict = operation.status === "conflict";
  const uncertain = operation.status === "pending" && operation.attempted;
  const cannotDiscard =
    uncertain || (operation.kind === "story.update" && !operation.serverValue);
  const hasServerVersion = Boolean(
    operation.serverVersion && operation.serverValue,
  );
  const resolve = async (resolution: "retry" | "discard") => {
    setError(undefined);
    try {
      await resolveDesktopOperation(scope, {
        operationId: operation.id,
        resolution,
        ...(resolution === "retry" && conflict
          ? {
              payload: operation.payload,
              baseVersion: operation.serverVersion!,
            }
          : {}),
      });
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
    }
  };
  return (
    <Box className="border-border rounded-xl border p-4">
      <Text className="font-medium">{operationName(operation)}</Text>
      <Text className="mt-1" color="muted">
        {operation.error ??
          (uncertain
            ? "Waiting for confirmation. Retrying uses the same operation so it cannot create a duplicate."
            : "Saved on this computer and waiting to sync.")}
      </Text>
      {conflict && hasServerVersion ? (
        <>
          <div className="my-4 overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr>
                  <th className="pr-4 pb-2 font-medium">Field</th>
                  <th className="pr-4 pb-2 font-medium">Server version</th>
                  <th className="pb-2 font-medium">Your saved change</th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(operation.payload).map(([field, value]) => (
                  <tr className="border-border border-t align-top" key={field}>
                    <th className="py-2 pr-4 font-normal">
                      {fieldLabels[field] ?? field}
                    </th>
                    <td className="max-w-80 py-2 pr-4 break-words whitespace-pre-wrap">
                      {outboxFieldValue(
                        client,
                        scope.workspaceSlug,
                        field,
                        operation.serverValue?.[field],
                      )}
                    </td>
                    <td className="max-w-80 py-2 break-words whitespace-pre-wrap">
                      {outboxFieldValue(
                        client,
                        scope.workspaceSlug,
                        field,
                        value,
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Flex align="center" gap={3}>
            <Checkbox
              checked={reviewed}
              disabled={busy}
              id={id}
              onCheckedChange={(checked) => {
                setReviewed(checked === true);
              }}
            />
            <label htmlFor={id}>
              I reviewed the server version and want to apply my saved changes.
            </label>
          </Flex>
        </>
      ) : null}
      {conflict && !hasServerVersion ? (
        <Text className="mt-3" color="muted">
          The server version is unavailable. Reconnect and refresh before
          choosing which changes to keep.
        </Text>
      ) : null}
      {!conflict && cannotDiscard && !uncertain ? (
        <Text className="mt-3" color="muted">
          Reconnect to load the server version before discarding this change.
        </Text>
      ) : null}
      {error ? (
        <Text className="mt-3" role="alert">
          {error}
        </Text>
      ) : null}
      <Flex className="mt-4" gap={3} wrap>
        {operation.status !== "pending" ? (
          <Button
            disabled={
              busy || !online || (conflict && (!reviewed || !hasServerVersion))
            }
            onClick={() => {
              void resolve("retry");
            }}
            size="sm"
            variant="outline"
          >
            {conflict ? "Apply reviewed changes" : "Retry change"}
          </Button>
        ) : null}
        {!confirmDiscard ? (
          <Button
            disabled={busy || cannotDiscard}
            onClick={() => {
              setConfirmDiscard(true);
            }}
            size="sm"
            variant="naked"
          >
            Discard local change
          </Button>
        ) : (
          <>
            <Text>Remove this saved change and keep the server version?</Text>
            <Button
              color="danger"
              disabled={busy || cannotDiscard}
              onClick={() => {
                void resolve("discard");
              }}
              size="sm"
            >
              Confirm discard
            </Button>
            <Button
              disabled={busy}
              onClick={() => {
                setConfirmDiscard(false);
              }}
              size="sm"
              variant="naked"
            >
              Cancel
            </Button>
          </>
        )}
      </Flex>
    </Box>
  );
};
