"use client";

import { useState } from "react";
import Link from "next/link";
import { Box, Button, Command, Divider, Flex, Popover, Text } from "ui";
import { CloseIcon, DocsIcon, ObjectiveIcon, StoryIcon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { useSearch } from "@/modules/search/hooks/use-search";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useDocumentRelationshipMutations } from "./hooks";
import type {
  DocumentRelationType,
  RelatedWork,
  WorkspaceDocument,
} from "./types";

const getRelatedWorkPath = (
  work: RelatedWork,
  withWorkspace: (path: string) => string,
) => {
  if (work.entityType === "story" && work.reference) {
    return withWorkspace(`/work/${work.reference}`);
  }
  return withWorkspace(`/teams/${work.teamId}/objectives/${work.entityId}`);
};

const WorkIcon = ({ type }: { type: DocumentRelationType }) =>
  type === "story" ? (
    <StoryIcon className="size-4" />
  ) : (
    <ObjectiveIcon className="size-4" />
  );

const RelationshipPicker = ({ document }: { document: WorkspaceDocument }) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const { data: teams = [] } = useTeams();
  const { data, isFetching } = useSearch({
    query: query.trim(),
    type: "all",
    pageSize: 12,
  });
  const { add } = useDocumentRelationshipMutations(document.id);
  const existing = new Set(
    document.relatedWork.map((work) => `${work.entityType}:${work.entityId}`),
  );
  const teamCodeById = new Map(teams.map((team) => [team.id, team.code]));
  const results = [
    ...(data?.stories ?? []).map((story) => {
      const teamCode = story.team?.code ?? teamCodeById.get(story.teamId);
      return {
        entityId: story.id,
        entityType: "story" as const,
        title: story.title,
        reference: teamCode
          ? `${teamCode}-${story.sequenceId}`
          : String(story.sequenceId),
      };
    }),
    ...(data?.objectives ?? [])
      .filter((objective) => Boolean(objective.teamId))
      .map((objective) => ({
        entityId: objective.id,
        entityType: "objective" as const,
        title: objective.name,
        reference: null,
      })),
  ].filter((item) => !existing.has(`${item.entityType}:${item.entityId}`));

  const select = (entityType: DocumentRelationType, entityId: string) => {
    add.mutate(
      { entityType, entityId },
      {
        onSuccess: () => {
          setOpen(false);
          setQuery("");
        },
      },
    );
  };

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <Button
          align="center"
          color="tertiary"
          fullWidth
          size="md"
          variant="outline"
        >
          Add new relationship
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        className="border-border-strong bg-surface-elevated w-[22rem] border"
      >
        <Command shouldFilter={false}>
          <Command.Input
            autoFocus
            onValueChange={setQuery}
            placeholder="Search stories and objectives..."
            value={query}
          />
          <Divider className="my-2" />
        </Command>
        <Box className="max-h-72 overflow-y-auto px-1.5">
          {!query.trim() ? (
            <Text className="px-2 py-4 text-center" color="muted">
              Start typing to find work.
            </Text>
          ) : null}
          {isFetching ? (
            <Text className="px-2 py-4 text-center" color="muted">
              Searching...
            </Text>
          ) : null}
          {!isFetching && query.trim() && results.length === 0 ? (
            <Text className="px-2 py-4 text-center" color="muted">
              No matching work.
            </Text>
          ) : null}
          {results.map((item) => (
            <button
              className="hover:bg-state-hover flex w-full items-center gap-2 rounded-md px-2 py-2 text-left"
              key={`${item.entityType}:${item.entityId}`}
              onClick={() => {
                select(item.entityType, item.entityId);
              }}
              type="button"
            >
              <WorkIcon type={item.entityType} />
              <Text className="min-w-0 flex-1 truncate">
                {item.reference ? (
                  <span className="text-text-muted mr-2">{item.reference}</span>
                ) : null}
                {item.title}
              </Text>
            </button>
          ))}
        </Box>
      </Popover.Content>
    </Popover>
  );
};

export const RelatedWorkPanel = ({
  document,
  onClose,
}: {
  document: WorkspaceDocument;
  onClose: () => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const { remove } = useDocumentRelationshipMutations(document.id);
  const grouped = {
    story: document.relatedWork.filter((work) => work.entityType === "story"),
    objective: document.relatedWork.filter(
      (work) => work.entityType === "objective",
    ),
  };

  return (
    <Box
      aria-labelledby="related-work-title"
      className="bg-surface/75 dark:bg-surface/75 flex h-full min-h-0 flex-col backdrop-blur-xl"
      role="complementary"
    >
      <Box className="shrink-0 px-6 pt-6">
        <Flex align="center" className="mb-6" justify="between">
          <Text fontSize="xl" fontWeight="semibold" id="related-work-title">
            Related work
          </Text>
          <Button
            aria-label="Close related work"
            asIcon
            color="tertiary"
            onClick={onClose}
            size="sm"
            variant="naked"
          >
            <CloseIcon className="size-5" />
          </Button>
        </Flex>
        {document.canEdit ? <RelationshipPicker document={document} /> : null}
      </Box>

      <Box className="min-h-0 flex-1 overflow-y-auto px-6 pb-6">
        {document.relatedWork.length === 0 ? (
          <Flex
            align="center"
            className="h-full min-h-64 px-4 text-center"
            direction="column"
            justify="center"
          >
            <DocsIcon
              className="text-text-muted mb-5 size-14"
              strokeWidth={1.5}
            />
            <Text className="mb-2" fontSize="lg" fontWeight="semibold">
              Nothing to see here!
            </Text>
            <Text className="max-w-sm" color="muted">
              Relate your work to this document and it&apos;ll appear here.
            </Text>
          </Flex>
        ) : null}

        {(["story", "objective"] as const).map((type) => {
          const items = grouped[type];
          if (items.length === 0) return null;
          return (
            <Box className="mt-6" key={type}>
              <Text className="mb-2" color="muted" fontWeight="semibold">
                {type === "story" ? "Stories" : "Objectives"}
              </Text>
              <Box>
                {items.map((work) => (
                  <Flex
                    align="center"
                    className="border-border/60 dark:border-border-strong/80 hover:bg-state-hover group border-b-[0.5px] py-2.5 last:border-b-0"
                    gap={2}
                    key={`${work.entityType}:${work.entityId}`}
                  >
                    <WorkIcon type={work.entityType} />
                    <Link
                      className="min-w-0 flex-1"
                      href={getRelatedWorkPath(work, withWorkspace)}
                    >
                      <Text className="truncate">
                        {work.reference ? (
                          <span className="text-text-muted mr-2">
                            {work.reference}
                          </span>
                        ) : null}
                        {work.title}
                      </Text>
                    </Link>
                    {document.canEdit ? (
                      <button
                        aria-label={`Remove ${work.title}`}
                        className="text-text-muted hover:text-foreground opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
                        onClick={() => {
                          remove.mutate({
                            entityType: work.entityType,
                            entityId: work.entityId,
                          });
                        }}
                        type="button"
                      >
                        <CloseIcon className="size-4" />
                      </button>
                    ) : null}
                  </Flex>
                ))}
              </Box>
            </Box>
          );
        })}
      </Box>
    </Box>
  );
};
