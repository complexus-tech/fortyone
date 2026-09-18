"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Box, Button, Flex, Text } from "ui";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { documentKeys } from "@/shared/documents/keys";
import { setDocumentPublicLinkAction } from "./actions";
import type { WorkspaceDocument } from "./types";

export function DocumentPublicLink({
  document,
}: {
  document: WorkspaceDocument;
}) {
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const update = useMutation({
    mutationFn: (enabled: boolean) =>
      setDocumentPublicLinkAction(document.id, workspaceSlug, enabled),
    onSuccess: (saved) => {
      queryClient.setQueryData(
        documentKeys.detail(workspaceSlug, document.id),
        saved,
      );
      toast.success(
        saved.publicToken ? "Public link enabled" : "Public link revoked",
      );
    },
    onError: () => toast.error("Could not update public sharing"),
  });
  const path = document.publicToken
    ? `/shared/docs/${document.publicToken}`
    : null;
  const copy = async () => {
    if (!path) return;
    try {
      await navigator.clipboard.writeText(
        new URL(path, window.location.origin).href,
      );
      toast.success("Public link copied");
    } catch {
      toast.error("Could not copy the link");
    }
  };
  return (
    <Box className="border-border border-t px-4 py-4">
      <Text fontWeight="semibold">Public link</Text>
      <Text className="mt-1" color="muted">
        {path
          ? "Anyone with this link can view this document without signing in."
          : "Allow anyone with the link to view. Your workspace sharing settings stay the same."}
      </Text>
      {path ? (
        <>
          <Flex className="mt-3" gap={2}>
            <Button
              disabled={update.isPending}
              onClick={() => void copy()}
              size="sm"
            >
              Copy public link
            </Button>
            <a
              className="hover:bg-state-hover rounded-lg px-3 py-2 font-medium"
              href={path}
              rel="noreferrer"
              target="_blank"
            >
              Open
            </a>
          </Flex>
          <Button
            className="mt-3"
            color="tertiary"
            disabled={update.isPending}
            onClick={() => {
              update.mutate(false);
            }}
            size="sm"
          >
            {update.isPending ? "Revoking…" : "Revoke public link"}
          </Button>
        </>
      ) : (
        <Button
          className="mt-3"
          disabled={update.isPending}
          onClick={() => {
            update.mutate(true);
          }}
          size="sm"
        >
          {update.isPending ? "Enabling…" : "Enable public link"}
        </Button>
      )}
    </Box>
  );
}
