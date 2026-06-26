"use client";

import type { Dispatch, ReactNode, SetStateAction } from "react";
import { useState } from "react";
import { Box, Button, Flex, Input, Popover, Skeleton, Text } from "ui";
import {
  CloseIcon,
  CopyIcon,
  DuplicateIcon,
  ErrorIcon,
  LinkIcon,
  SearchIcon,
  WarningIcon,
} from "icons";
import { cn } from "lib";
import { NewStoryDialog } from "@/components/ui/new-story-dialog";
import { useTerminology } from "@/hooks";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useSearch } from "@/modules/search/hooks/use-search";
import type { Story } from "@/modules/stories/types";
import { useAddAssociationMutation } from "../hooks/add-association-mutation";
import type { DetailedStory, StoryAssociationType } from "../types";

type RelationshipDirection = "outgoing" | "incoming";

type RelationshipOption = {
  direction: RelationshipDirection;
  icon: ReactNode;
  label: string;
  type: StoryAssociationType;
};

const RELATIONSHIP_OPTIONS: RelationshipOption[] = [
  {
    direction: "outgoing",
    icon: <LinkIcon className="h-5" />,
    label: "Relates to",
    type: "related",
  },
  {
    direction: "outgoing",
    icon: <WarningIcon className="text-warning h-5" />,
    label: "Blocks",
    type: "blocking",
  },
  {
    direction: "incoming",
    icon: <ErrorIcon className="text-danger h-5" />,
    label: "Blocked by",
    type: "blocking",
  },
  {
    direction: "outgoing",
    icon: <DuplicateIcon className="text-warning h-5" />,
    label: "Duplicates",
    type: "duplicate",
  },
  {
    direction: "incoming",
    icon: <CopyIcon className="text-warning h-5" />,
    label: "Duplicated by",
    type: "duplicate",
  },
];

const SEARCH_PAGE_SIZE = 8;
const EMPTY_ASSOCIATION_STORY_IDS: string[] = [];

const StorySearchSkeleton = () => (
  <Flex
    align="center"
    className="px-1 py-2"
    data-testid="relationship-search-skeleton"
    gap={3}
  >
    <Skeleton className="size-5 shrink-0 rounded-full" />
    <Skeleton className="h-5 w-12 rounded" />
    <Skeleton className="h-5 w-56 max-w-[60%] rounded" />
  </Flex>
);

const getStoryKey = (story: Story, teamCode: string) =>
  `${story.team?.code ?? teamCode}-${story.sequenceId}`;

const StorySearchResults = ({
  isFetching,
  isPending,
  onSelectStory,
  query,
  statusColorById,
  stories,
  teamCode,
}: {
  isFetching: boolean;
  isPending: boolean;
  onSelectStory: (story: Story) => void;
  query: string;
  statusColorById: Map<string, string>;
  stories: Story[];
  teamCode: string;
}) => {
  if (!query) {
    return null;
  }

  if (isFetching) {
    return (
      <Box className="mt-3">
        <StorySearchSkeleton />
        <StorySearchSkeleton />
      </Box>
    );
  }

  if (stories.length === 0) {
    return (
      <Box className="mt-3">
        <Text className="px-1 py-2" color="muted" fontSize="sm">
          No matching stories in {teamCode}.
        </Text>
      </Box>
    );
  }

  return (
    <Box className="mt-3">
      {stories.map((story) => (
        <button
          className="hover:bg-state-hover focus-visible:bg-state-active flex w-full items-center gap-3 rounded-lg px-1 py-2 text-left focus-visible:outline-0"
          disabled={isPending}
          key={story.id}
          onClick={() => {
            onSelectStory(story);
          }}
          type="button"
        >
          <span
            aria-hidden="true"
            className="bg-warning size-3.5 shrink-0 rounded-full"
            style={{
              backgroundColor: statusColorById.get(story.statusId),
            }}
          />
          <Text className="line-clamp-1">
            <span className="text-text-muted mr-2">
              {getStoryKey(story, teamCode)}
            </span>
            {story.title}
          </Text>
        </button>
      ))}
    </Box>
  );
};

export const StoryRelationshipPicker = ({
  currentStoryId,
  currentStoryTitle,
  existingAssociationStoryIds = EMPTY_ASSOCIATION_STORY_IDS,
  teamCode,
  teamId,
}: {
  currentStoryId: string;
  currentStoryTitle: string;
  existingAssociationStoryIds?: string[];
  teamCode: string;
  teamId: string;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selectedOption, setSelectedOption] = useState<RelationshipOption>(
    RELATIONSHIP_OPTIONS[0],
  );
  const [isCreateStoryOpen, setIsCreateStoryOpen] = useState(false);
  const { getTermDisplay } = useTerminology();
  const addAssociation = useAddAssociationMutation();
  const trimmedQuery = query.trim();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data, isFetching } = useSearch({
    pageSize: SEARCH_PAGE_SIZE,
    query: trimmedQuery,
    teamId,
    type: "stories",
  });

  const existingIds = new Set(existingAssociationStoryIds);
  const stories = (data?.stories ?? []).filter(
    (story) => story.id !== currentStoryId && !existingIds.has(story.id),
  );
  const statusColorById = new Map(
    statuses.map((status) => [status.id, status.color]),
  );
  const storyTerm = getTermDisplay("storyTerm");

  const getAssociationPayload = (storyId: string) =>
    selectedOption.direction === "incoming"
      ? {
          fromStoryId: storyId,
          toStoryId: currentStoryId,
          type: selectedOption.type,
        }
      : {
          fromStoryId: currentStoryId,
          toStoryId: storyId,
          type: selectedOption.type,
        };

  const resetPicker = () => {
    setIsOpen(false);
    setQuery("");
    setSelectedOption(RELATIONSHIP_OPTIONS[0]);
  };

  const handleSelectStory = (story: Story) => {
    addAssociation.mutate(getAssociationPayload(story.id), {
      onSuccess: () => {
        resetPicker();
      },
    });
  };

  const handleOpenCreateStoryDialog = () => {
    setIsOpen(false);
    setQuery("");
    setIsCreateStoryOpen(true);
  };

  const handleCreatedStory = async (createdStory: DetailedStory) => {
    await addAssociation.mutateAsync(getAssociationPayload(createdStory.id));
  };

  const handleCreateStoryDialogOpenChange: Dispatch<SetStateAction<boolean>> = (
    open,
  ) => {
    setIsCreateStoryOpen((currentOpen) => {
      const nextOpen = typeof open === "function" ? open(currentOpen) : open;

      if (!nextOpen) {
        setSelectedOption(RELATIONSHIP_OPTIONS[0]);
      }

      return nextOpen;
    });
  };

  return (
    <>
      <Popover onOpenChange={setIsOpen} open={isOpen}>
        <Popover.Trigger asChild>
          <Button
            active={isOpen}
            color="tertiary"
            leftIcon={<LinkIcon className="h-4" />}
            size="sm"
            variant="naked"
          >
            Association
          </Button>
        </Popover.Trigger>

        <Popover.Content
          align="end"
          className="z-[100] w-max max-w-[calc(100vw-2rem)] overflow-hidden p-0"
          sideOffset={8}
        >
          <Box className="p-4">
            <Flex align="center" className="mb-4 min-w-0" justify="between">
              <Text
                className="line-clamp-1 min-w-0 flex-1 pr-3"
                color="muted"
                fontWeight="semibold"
                title={currentStoryTitle}
              >
                {currentStoryTitle}
              </Text>
              <Button
                aria-label="Close relationship picker"
                asIcon
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                }}
                rounded="full"
                size="xs"
                variant="naked"
              >
                <CloseIcon className="h-4" />
              </Button>
            </Flex>

            <Flex className="mb-4 flex-nowrap" gap={2}>
              {RELATIONSHIP_OPTIONS.map((option) => (
                <Button
                  align="left"
                  className={cn(
                    "dark:bg-surface-elevated/60 dark:hover:bg-surface-elevated/80 h-auto min-h-[4.75rem] min-w-24 shrink-0 flex-col items-start justify-center gap-1.5 px-3 py-2 text-left text-[0.95rem] leading-snug",
                    {
                      "dark:bg-surface-elevated/80 ring-ring bg-transparent ring-2 hover:bg-transparent":
                        selectedOption.label === option.label,
                    },
                  )}
                  color="tertiary"
                  key={option.label}
                  onClick={() => {
                    setSelectedOption(option);
                  }}
                  type="button"
                  variant="outline"
                >
                  {option.icon}
                  <span className="whitespace-nowrap">{option.label}</span>
                </Button>
              ))}
            </Flex>
            <Input
              autoFocus
              onChange={(event) => {
                setQuery(event.target.value);
              }}
              placeholder={`Search ${getTermDisplay("storyTerm", {
                capitalize: true,
              })} Title or ID`}
              rightIcon={<SearchIcon className="text-icon h-5" />}
              value={query}
            />
            {!trimmedQuery ? (
              <Flex align="center" className="mt-4" gap={2}>
                <Text color="muted">Search for an existing {storyTerm} or</Text>
                <Button
                  color="tertiary"
                  disabled={addAssociation.isPending}
                  onClick={handleOpenCreateStoryDialog}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  Create related {storyTerm}
                </Button>
              </Flex>
            ) : null}

            <StorySearchResults
              isFetching={isFetching}
              isPending={addAssociation.isPending}
              onSelectStory={handleSelectStory}
              query={trimmedQuery}
              statusColorById={statusColorById}
              stories={stories}
              teamCode={teamCode}
            />
          </Box>
        </Popover.Content>
      </Popover>
      <NewStoryDialog
        isOpen={isCreateStoryOpen}
        onCreated={handleCreatedStory}
        setIsOpen={handleCreateStoryDialogOpenChange}
        teamId={teamId}
      />
    </>
  );
};
