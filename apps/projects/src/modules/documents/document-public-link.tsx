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
    <Box className="border-border border-t px-5 py-4">
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
        <Flex className="border-border -mx-5 mt-4 border-t px-5 pt-4" gap={2}>
          <Button
            align="center"
            className="flex-1"
            color="tertiary"
            disabled={update.isPending}
            leftIcon={<CopyIcon className="size-4" />}
            onClick={() => void copy()}
            variant="outline"
          >
            Copy public link
          </Button>
          <a
            aria-label="Open public document in a new tab"
            className="border-border hover:bg-state-hover focus-visible:ring-ring flex size-10 items-center justify-center rounded-xl border outline-none focus-visible:ring-2"
            href={path}
            rel="noreferrer"
            target="_blank"
          >
            <ArrowUpRightIcon className="size-4" />
          </a>
        </Flex>
      ) : null}
    </Box>
  );
}
