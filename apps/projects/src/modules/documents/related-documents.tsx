"use client";

import Link from "next/link";
import { Box, Flex, Skeleton, Text } from "ui";
import { DocsIcon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { useRelatedDocuments } from "./hooks";
import type { DocumentRelationType } from "./types";

export const RelatedDocuments = ({
  entityId,
  entityType,
}: {
  entityId: string;
  entityType: DocumentRelationType;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const { data: documents = [], isPending } = useRelatedDocuments(
    entityType,
    entityId,
  );

  if (!isPending && documents.length === 0) return null;

  return (
    <Box className="border-border/70 mt-6 border-t pt-5">
      <Flex align="center" className="mb-3" gap={2}>
        <DocsIcon className="text-text-muted size-5" />
        <Text fontWeight="semibold">Related documents</Text>
      </Flex>
      {isPending ? <Skeleton className="h-12 w-full rounded-xl" /> : null}
      <Box className="space-y-1">
        {documents.map((document) => (
          <Link
            className="hover:bg-state-hover flex items-center gap-2.5 rounded-xl px-3 py-2.5"
            href={withWorkspace(`/docs/${document.id}`)}
            key={document.id}
          >
            <DocsIcon className="text-text-muted size-4 shrink-0" />
            <Box className="min-w-0">
              <Text className="truncate" fontWeight="medium">
                {document.title}
              </Text>
              <Text className="line-clamp-1" color="muted">
                {document.contentText || "Empty document"}
              </Text>
            </Box>
          </Link>
        ))}
      </Box>
    </Box>
  );
};
