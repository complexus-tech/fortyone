"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Box, Button, Flex, Switch, Text } from "ui";
import { ArrowUpRightIcon, CopyIcon, LinkIcon } from "icons";
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
      void queryClient.invalidateQueries({
        queryKey: documentKeys.lists(workspaceSlug),
      });
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
  let accessLabel = path
    ? "Anyone with the link can view"
    : "Off · Enable view-only access";
  if (update.isPending) accessLabel = "Updating public access…";
  return (
    <Box className="border-border border-t px-5 py-5 text-base">
      <Flex align="center" gap={3}>
        <LinkIcon className="size-5 shrink-0" />
        <Box className="min-w-0 flex-1">
          <label
            className="block font-medium"
            htmlFor={`public-link-${document.id}`}
          >
            Public link
          </label>
          <Text aria-live="polite" className="mt-0.5" color="muted">
            {accessLabel}
          </Text>
        </Box>
        <Switch
          aria-label="Enable public link"
          checked={Boolean(path)}
          disabled={update.isPending}
          id={`public-link-${document.id}`}
          onCheckedChange={(enabled) => {
            update.mutate(enabled);
          }}
        />
      </Flex>
      {path ? (
        <Flex className="mt-4 flex-wrap" gap={2}>
          <Button
            align="center"
            className="flex-1 text-base whitespace-nowrap"
            disabled={update.isPending}
            leftIcon={<CopyIcon className="size-4" />}
            onClick={() => void copy()}
          >
            Copy public link
          </Button>
          <Button
            align="center"
            className="flex-1 text-base whitespace-nowrap"
            color="tertiary"
            href={path}
            leftIcon={<ArrowUpRightIcon className="size-4" />}
            target="_blank"
            variant="outline"
          >
            Open preview
            <span className="sr-only"> in a new tab</span>
          </Button>
        </Flex>
      ) : null}
    </Box>
  );
}
