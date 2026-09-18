"use client";

import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Avatar, Box, Button, Dialog, Flex, Skeleton, Text } from "ui";
import { cn } from "lib";
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
  const history = useQuery({
    queryKey: ["document-history", workspaceSlug, document.id],
    queryFn: () => getDocumentRevisions(document.id, context),
    enabled: Boolean(session),
  });
  const revisions = history.data ?? [];
  const revision = selected ?? revisions[0]?.revision;
  const isLatest = revision === revisions[0]?.revision;
  const preview = useQuery({
    queryKey: [
      "document-history-preview",
      workspaceSlug,
      document.id,
      revision,
    ],
    queryFn: () => getDocumentRevision(document.id, revision, context),
    enabled: Boolean(revision && session),
    retry: false,
  });
  const restore = useMutation({
    mutationFn: async () => {
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
        "This version changed or is no longer available. Refresh the history and try again.",
      ),
  });
  const refresh = () => {
    setSelected(null);
    setConfirmRestore(false);
    void history.refetch();
    void preview.refetch();
  };
  let restoreLabel = confirmRestore ? "Confirm restore" : "Restore version";
  if (restore.isPending) restoreLabel = "Restoring…";
  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !restore.isPending) onClose();
      }}
      open={isOpen}
    >
      <Dialog.Content
        aria-describedby={undefined}
        className="mt-0 flex max-h-[90dvh] flex-col md:mt-0"
        hideClose={restore.isPending}
        overlayClassName="items-center py-5"
        size="xl"
      >
        <Dialog.Header className="shrink-0 px-6 py-5">
          <Dialog.Title className="pr-10 text-2xl">
            Version history
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="border-border flex h-[min(72dvh,48rem)] min-h-0 flex-col overflow-hidden border-t px-0 pt-0 pb-0 sm:flex-row-reverse">
          <aside
            aria-label="Saved versions"
            className="border-border max-h-44 shrink-0 overflow-y-auto border-b p-3 sm:max-h-none sm:w-80 sm:border-b-0 sm:border-l"
          >
            {history.isPending ? (
              <Box className="space-y-3 p-2">
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
              </Box>
            ) : null}
            {history.isError ? (
              <Box className="p-3">
                <Text color="muted">Could not load history.</Text>
                <Button
                  className="mt-3"
                  color="tertiary"
                  onClick={refresh}
                  size="sm"
                >
                  Try again
                </Button>
              </Box>
            ) : null}
            {revisions.map((item, index) => {
              const member = members.find(
                (person) => person.id === item.editedBy,
              );
              const name =
                member?.username || member?.fullName || "Former member";
              const date = new Date(item.createdAt);
              return (
                <button
                  aria-pressed={item.revision === revision}
                  className={cn(
                    "hover:bg-state-hover focus-visible:ring-ring mb-1 flex w-full gap-3 rounded-xl p-3 text-left outline-none focus-visible:ring-2",
                    item.revision === revision && "bg-state-selected",
                  )}
                  disabled={restore.isPending}
                  key={item.revision}
                  onClick={() => {
                    setSelected(item.revision);
                    setConfirmRestore(false);
                  }}
                  type="button"
                >
                  <span className="min-w-0 flex-1">
                    <Flex align="center" gap={2} justify="between">
                      <Text className="truncate" fontWeight="medium">
                        {index === 0
                          ? "Latest saved"
                          : date.toLocaleDateString(undefined, {
                              month: "short",
                              day: "numeric",
                              ...(date.getFullYear() !==
                              new Date().getFullYear()
                                ? { year: "numeric" as const }
                                : {}),
                            })}
                      </Text>
                      <Text className="shrink-0" color="muted">
                        {date.toLocaleTimeString(undefined, {
                          hour: "numeric",
                          minute: "2-digit",
                        })}
                      </Text>
                    </Flex>
                    <Flex align="center" className="mt-1.5" gap={2}>
                      <Avatar name={name} size="xs" src={member?.avatarUrl} />
                      <Text className="truncate" color="muted">
                        {name}
                      </Text>
                    </Flex>
                  </span>
                </button>
              );
            })}
            {history.isSuccess && revisions.length === 0 ? (
              <Text className="p-3" color="muted">
                History will appear after your first save.
              </Text>
            ) : null}
          </aside>
          <section
            aria-busy={preview.isFetching}
            aria-label="Version preview"
            className="min-h-0 min-w-0 flex-1 overflow-y-auto px-6 py-6 sm:px-8"
          >
            {preview.isPending && revisions.length > 0 ? (
              <Box className="space-y-4">
                <Skeleton className="h-9 w-2/3" />
                <Skeleton className="h-5 w-full" />
                <Skeleton className="h-5 w-4/5" />
              </Box>
            ) : null}
            {preview.isError ? (
              <Box className="py-8">
                <Text fontWeight="medium">
                  This version is no longer available
                </Text>
                <Text className="mt-2" color="muted">
                  Recent saves may have updated it, or it may have expired.
                </Text>
                <Button className="mt-4" color="tertiary" onClick={refresh}>
                  Refresh history
                </Button>
              </Box>
            ) : null}
            {preview.isSuccess ? (
              <>
                <Flex
                  align="center"
                  className="mb-6 flex-col items-start sm:flex-row sm:items-center"
                  gap={3}
                  justify="between"
                >
                  <Text
                    as="h2"
                    className="min-w-0 flex-1 break-words"
                    fontSize="3xl"
                    fontWeight="semibold"
                  >
                    {preview.data.title}
                  </Text>
                  {!isLatest && canRestore ? (
                    <Flex className="shrink-0" gap={2}>
                      {confirmRestore ? (
                        <Button
                          color="tertiary"
                          disabled={restore.isPending}
                          onClick={() => {
                            setConfirmRestore(false);
                          }}
                          variant="naked"
                        >
                          Cancel
                        </Button>
                      ) : null}
                      <Button
                        disabled={restore.isPending}
                        onClick={() => {
                          if (confirmRestore) restore.mutate();
                          else setConfirmRestore(true);
                        }}
                      >
                        {restoreLabel}
                      </Button>
                    </Flex>
                  ) : null}
                </Flex>
                <DocumentPreview html={preview.data.contentHtml ?? ""} />
              </>
            ) : null}
          </section>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
}
