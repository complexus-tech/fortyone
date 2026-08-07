"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Box, Button, Flex, Menu, Skeleton, Text, TimeAgo } from "ui";
import { DeleteIcon, DocsIcon, MoreHorizontalIcon } from "icons";
import { RowWrapper } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import { useDocumentRelationshipMutations, useRelatedDocuments } from "./hooks";
import type { DocumentRelationType, WorkspaceDocument } from "./types";

const RelatedDocumentRow = ({
  document,
  entityId,
  entityType,
}: {
  document: WorkspaceDocument;
  entityId: string;
  entityType: DocumentRelationType;
}) => {
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const { remove } = useDocumentRelationshipMutations(document.id);
  const documentPath = withWorkspace(`/docs/${document.id}`);

  return (
    <RowWrapper className="gap-8 px-1 py-2 md:px-1">
      <Link
        className="flex min-w-0 flex-1 items-center gap-2"
        href={documentPath}
      >
        <DocsIcon className="text-text-muted mx-0.5 h-[1.3rem] shrink-0" />
        <Text className="truncate font-medium" title={document.title}>
          {document.title}
        </Text>
      </Link>
      <Flex align="center" className="shrink-0" gap={3}>
        <Text color="muted">
          <TimeAgo timestamp={document.updatedAt} />
        </Text>
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${document.title}`}
              asIcon
              color="tertiary"
              rounded="full"
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="min-w-48">
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  router.push(documentPath);
                }}
              >
                <DocsIcon className="h-[1.15rem]" />
                Open document
              </Menu.Item>
            </Menu.Group>
            {document.canEdit ? (
              <>
                <Menu.Separator className="my-1.5" />
                <Menu.Group>
                  <Menu.Item
                    className="text-error tracking-wide"
                    disabled={remove.isPending}
                    onSelect={() => {
                      remove.mutate({ entityId, entityType });
                    }}
                  >
                    <DeleteIcon />
                    Remove association
                  </Menu.Item>
                </Menu.Group>
              </>
            ) : null}
          </Menu.Items>
        </Menu>
      </Flex>
    </RowWrapper>
  );
};

export const RelatedDocuments = ({
  entityId,
  entityType,
}: {
  entityId: string;
  entityType: DocumentRelationType;
}) => {
  const { data: documents = [], isPending } = useRelatedDocuments(
    entityType,
    entityId,
  );

  if (!isPending && documents.length === 0) return null;

  return (
    <Box className="mt-4">
      <Flex
        align="center"
        className="border-border border-b-[0.5px] pb-2"
        gap={2}
      >
        <DocsIcon className="text-text-muted size-5" />
        <Text fontWeight="semibold">Related documents</Text>
      </Flex>
      {isPending ? <Skeleton className="h-11 w-full rounded-none" /> : null}
      <Box>
        {documents.map((document) => (
          <RelatedDocumentRow
            document={document}
            entityId={entityId}
            entityType={entityType}
            key={document.id}
          />
        ))}
      </Box>
    </Box>
  );
};
