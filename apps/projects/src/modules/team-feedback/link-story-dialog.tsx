"use client";

import { useState } from "react";
import { Box, Dialog, Flex, Input, Skeleton, Text } from "ui";
import { SearchIcon } from "icons";
import { useDebouncedCallback } from "@/hooks/debounce";
import { Dot } from "@/components/ui";
import { useSearch } from "@/modules/search/hooks/use-search";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { usePlanTeamFeedback } from "./hooks/use-plan-feedback";

const SEARCH_PAGE_SIZE = 8;

const StorySearchSkeleton = () => (
  <Flex align="center" className="px-2 py-2.5" gap={3}>
    <Skeleton className="size-3.5 shrink-0 rounded-sm" />
    <Skeleton className="h-4 w-16 rounded" />
    <Skeleton className="h-4 w-52 max-w-[55%] rounded" />
  </Flex>
);

export const LinkFeedbackStoryDialog = ({
  feedbackId,
  isOpen,
  onLinked,
  onOpenChange,
  teamId,
}: {
  feedbackId: string;
  isOpen: boolean;
  onLinked: (storyId: string) => void;
  onOpenChange: (open: boolean) => void;
  teamId: string;
}) => {
  const [query, setQuery] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const { callback: updateSearchQuery, cancel: cancelSearchQuery } =
    useDebouncedCallback<string>(setSearchQuery, 300);
  const { data, isFetching } = useSearch({
    pageSize: SEARCH_PAGE_SIZE,
    query: searchQuery,
    teamId,
    type: "stories",
  });
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const linkStory = usePlanTeamFeedback();
  const statusColorById = new Map(
    statuses.map((status) => [status.id, status.color]),
  );
  const stories = data?.stories ?? [];

  const handleOpenChange = (open: boolean) => {
    onOpenChange(open);
    if (!open) {
      cancelSearchQuery();
      setQuery("");
      setSearchQuery("");
    }
  };

  const renderResults = () => {
    if (!searchQuery) {
      return (
        <Text className="px-2 py-6 text-center" color="muted">
          Search by story title or identifier.
        </Text>
      );
    }

    if (isFetching) {
      return (
        <>
          <StorySearchSkeleton />
          <StorySearchSkeleton />
          <StorySearchSkeleton />
        </>
      );
    }

    if (stories.length === 0) {
      return (
        <Text className="px-2 py-6 text-center" color="muted">
          No matching stories in this team.
        </Text>
      );
    }

    return (
      <Box className="max-h-72 overflow-y-auto">
        {stories.map((story) => {
          const identifier = `${story.team?.code ?? "STORY"}-${story.sequenceId}`;
          return (
            <button
              className="hover:bg-state-hover focus-visible:bg-state-active flex w-full items-center gap-3 rounded-lg px-2 py-2.5 text-left transition outline-none"
              disabled={linkStory.isPending}
              key={story.id}
              onClick={() => {
                linkStory.mutate(
                  {
                    feedbackId,
                    payload: { teamId, storyId: story.id },
                  },
                  {
                    onSuccess: (response) => {
                      if (!response.error?.message && response.data?.storyId) {
                        handleOpenChange(false);
                        onLinked(response.data.storyId);
                      }
                    },
                  },
                );
              }}
              type="button"
            >
              <Dot
                className="size-3.5"
                color={statusColorById.get(story.statusId)}
              />
              <Text className="line-clamp-1 min-w-0">
                <span className="text-text-muted mr-2">{identifier}</span>
                {story.title}
              </Text>
            </button>
          );
        })}
      </Box>
    );
  };

  return (
    <Dialog onOpenChange={handleOpenChange} open={isOpen}>
      <Dialog.Content className="max-w-xl">
        <Dialog.Header className="px-6 pt-6 pb-2">
          <Dialog.Title className="text-lg">Link existing story</Dialog.Title>
        </Dialog.Header>
        <Dialog.Description>
          Connect this feedback to work that your team is already planning.
        </Dialog.Description>
        <Dialog.Body className="pt-4">
          <Flex
            align="center"
            className="border-border bg-surface-muted/40 rounded-xl border px-3"
            gap={2}
          >
            <SearchIcon className="text-text-muted h-4 shrink-0" />
            <Input
              aria-label="Search team stories"
              autoComplete="off"
              autoFocus
              className="h-10 border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
              name="feedback-story-search"
              onChange={(event) => {
                setQuery(event.target.value);
                updateSearchQuery(event.target.value.trim());
              }}
              placeholder="Search stories…"
              type="search"
              value={query}
            />
          </Flex>
          <Box className="mt-3 min-h-28">{renderResults()}</Box>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
