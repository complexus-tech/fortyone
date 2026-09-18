"use client";

import { useState } from "react";
import { useInfiniteQuery, useMutation, useQuery } from "@tanstack/react-query";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import {
  getDocument,
  getDocumentRevision,
  getDocumentRevisions,
} from "./queries";
import { restoreDocumentRevisionAction } from "./actions";
import { DocumentPreview } from "./document-preview";
import type { WorkspaceDocument } from "./types";

export function DocumentHistory({
  document,
  isOpen,
  onClose,
  canRestore,
}: {
  document: WorkspaceDocument;
  isOpen: boolean;
  onClose: () => void;
  canRestore: boolean;
}) {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const { data: members = [] } = useMembers();
  const [selected, setSelected] = useState<number | null>(null);
  const [confirmRestore, setConfirmRestore] = useState(false);
  const context = { session: session!, workspaceSlug };
  const history = useInfiniteQuery({
    queryKey: ["document-history", workspaceSlug, document.id],
    initialPageParam: 0,
    queryFn: ({ pageParam }) =>
      getDocumentRevisions(document.id, context, pageParam),
    getNextPageParam: (last) =>
      last.length === 50 ? last[last.length - 1].revision : undefined,
    enabled: Boolean(session),
  });
  const revisions = history.data?.pages.flat() ?? [];
  const revision = selected ?? revisions[0]?.revision;
  const preview = useQuery({
    queryKey: ["document-history", workspaceSlug, document.id, revision],
    queryFn: () => getDocumentRevision(document.id, revision, context),
    enabled: Boolean(revision && session),
  });
  const restore = useMutation({
    mutationFn: async () => {
      // Use a fresh revision for the explicit restore operation. Any subsequent
      // concurrent write still causes the API's atomic precondition to fail.
      const latest = await getDocument(document.id, context);
      return restoreDocumentRevisionAction(
        document.id,
        workspaceSlug,
        revision,
        latest.revision,
      );
    },
    onSuccess: () => {
      toast.success("Version restored");
      window.location.reload();
    },
    onError: () =>
      toast.error(
        "The document changed or could not be restored. Review the latest version and try again.",
      ),
  });
  let restoreLabel = confirmRestore
    ? "Confirm restore"
    : "Restore this version";
  if (restore.isPending) restoreLabel = "Restoring…";
  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !restore.isPending) onClose();
      }}
      open={isOpen}
    >
      <Dialog.Content className="max-w-5xl">
        <Dialog.Header>
          <Dialog.Title>Version history</Dialog.Title>
          <Dialog.Description>
            Review previous versions. Restoring creates a new version and keeps
            the history.
          </Dialog.Description>
        </Dialog.Header>
        <Flex className="mt-5 h-[60vh] min-h-0 flex-col sm:flex-row" gap={6}>
          <Box className="max-h-40 w-full shrink-0 space-y-2 overflow-y-auto sm:max-h-none sm:w-64">
            {history.isPending ? <Text>Loading versions…</Text> : null}
            {history.isError ? (
              <Button onClick={() => void history.refetch()}>
                Retry loading history
              </Button>
            ) : null}
            {revisions.map((item) => (
              <button
                aria-pressed={item.revision === revision}
                className={`w-full rounded-xl p-3 text-left ${item.revision === revision ? "bg-state-active" : "hover:bg-state-hover"}`}
                key={item.revision}
                onClick={() => {
                  setSelected(item.revision);
                  setConfirmRestore(false);
                }}
                type="button"
              >
                <Text fontWeight="medium">
                  {new Date(item.createdAt).toLocaleString()}
                </Text>
                <Text color="muted">
                  {members.find((member) => member.id === item.editedBy)
                    ?.fullName || "Teammate"}{" "}
                  · Version {item.revision}
                </Text>
              </button>
            ))}
            {history.hasNextPage ? (
              <Button
                disabled={history.isFetchingNextPage}
                onClick={() => void history.fetchNextPage()}
                size="sm"
              >
                Load older versions
              </Button>
            ) : null}
          </Box>
          <Box className="min-w-0 flex-1 overflow-y-auto px-2">
            {preview.isPending ? <Text>Loading preview…</Text> : null}
            {preview.isError ? <Text>Could not load this version.</Text> : null}
            {preview.isSuccess ? (
              <>
                <Text className="mb-5" fontSize="2xl" fontWeight="semibold">
                  {preview.data.title}
                </Text>
                <DocumentPreview html={preview.data.contentHtml ?? ""} />
              </>
            ) : null}
          </Box>
        </Flex>
        <Flex
          align="center"
          className="border-border mt-5 border-t pt-4"
          gap={3}
          justify="end"
        >
          {confirmRestore ? (
            <Text className="mr-auto">
              Replace the current content with this version for everyone?
            </Text>
          ) : null}
          <Button
            color="tertiary"
            disabled={restore.isPending}
            onClick={() => {
              confirmRestore ? setConfirmRestore(false) : onClose();
            }}
          >
            Cancel
          </Button>
          <Button
            disabled={!canRestore || !preview.data || restore.isPending}
            onClick={() => {
              confirmRestore ? restore.mutate() : setConfirmRestore(true);
            }}
          >
            {restoreLabel}
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
