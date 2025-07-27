import { Flex, Badge, Button, Tooltip, Box, Text, Checkbox } from "ui";
import {
  ArrowDown2Icon,
  ArrowUp2Icon,
  PlusIcon,
  SubStoryIcon,
  AiIcon,
} from "icons";
import { useState, useEffect } from "react";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { experimental_useObject as useObject } from "@ai-sdk/react";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { RowWrapper, StoriesBoard } from "@/components/ui";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTerminology, useUserRole } from "@/hooks";
import { substoryGenerationSchema } from "@/modules/stories/schemas";
import type { DetailedStory } from "../types";

export const SubStories = ({
  parent,
  setIsSubStoriesOpen,
  isSubStoriesOpen,
}: {
  parent: DetailedStory;
  setIsSubStoriesOpen: (value: boolean) => void;
  isSubStoriesOpen: boolean;
}) => {
  const { getTermDisplay } = useTerminology();
  const [isCreateSubStoryOpen, setIsCreateSubStoryOpen] = useState(false);
  const [selectedSubstories, setSelectedSubstories] = useState<Set<string>>(
    new Set(),
  );
  const [showSuggestions, setShowSuggestions] = useState(true);
  const { data: statuses = [] } = useTeamStatuses(parent.teamId);
  const { userRole } = useUserRole();
  const completedStatus = statuses.find(
    (status) => status.category === "completed",
  );

  const { object, submit, isLoading } = useObject({
    api: "/api/suggest-substories",
    schema: substoryGenerationSchema,
  });

  // Set all substories as selected by default when they're loaded
  useEffect(() => {
    if (object?.substories && object.substories.length > 0) {
      const titles = object.substories
        .map((s) => s?.title)
        .filter((title): title is string => Boolean(title));
      setSelectedSubstories(new Set(titles));
      setShowSuggestions(true);
    }
  }, [object?.substories]);

  const completedStories = parent.subStories.filter(
    (story) => story.statusId === completedStatus?.id,
  ).length;

  const toggleSelection = (substoryTitle: string) => {
    setSelectedSubstories((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(substoryTitle)) {
        newSet.delete(substoryTitle);
      } else {
        newSet.add(substoryTitle);
      }
      return newSet;
    });
  };

  const clearSelection = () => {
    setSelectedSubstories(new Set());
  };

  const handleAddSelected = () => {
    // eslint-disable-next-line no-console -- Temporary logging for development
    console.log("Selected substories:", Array.from(selectedSubstories));
    clearSelection();
    setShowSuggestions(false);
  };

  const handleCancel = () => {
    clearSelection();
    setShowSuggestions(false);
  };

  useHotkeys("c", () => {
    if (userRole !== "guest") {
      setIsCreateSubStoryOpen(true);
    }
  });

  return (
    <>
      <Flex
        align="center"
        className={cn({
          "border-b-[0.5px] border-gray-100 pb-2 dark:border-dark-200":
            !isSubStoriesOpen,
        })}
        justify={parent.subStories.length > 0 ? "between" : "end"}
      >
        {parent.subStories.length > 0 && (
          <Flex align="center" gap={2}>
            <Button
              color="tertiary"
              leftIcon={<SubStoryIcon />}
              onClick={() => {
                setIsSubStoriesOpen(!isSubStoriesOpen);
              }}
              rightIcon={
                isSubStoriesOpen ? (
                  <ArrowDown2Icon className="h-4 w-auto" />
                ) : (
                  <ArrowUp2Icon className="h-4 w-auto" />
                )
              }
              size="sm"
              variant="naked"
            >
              Sub {getTermDisplay("storyTerm", { variant: "plural" })}
            </Button>
            <Badge className="px-1.5" color="tertiary" rounded="full">
              {completedStories}/{parent.subStories.length} Done
            </Badge>
          </Flex>
        )}

        {userRole !== "guest" && (
          <Flex align="center" gap={2}>
            <Button
              color="tertiary"
              leftIcon={<AiIcon className="text-primary dark:text-primary" />}
              loading={isLoading}
              loadingText="Maya is thinking..."
              onClick={() => {
                submit(parent);
              }}
              size="sm"
              variant="naked"
            >
              Suggest Sub{" "}
              {getTermDisplay("storyTerm", {
                capitalize: true,
                variant: "plural",
              })}
            </Button>
            <Tooltip
              title={parent.subStories.length > 0 ? "Add Sub Story" : null}
            >
              <Button
                color="tertiary"
                leftIcon={<PlusIcon />}
                onClick={() => {
                  setIsCreateSubStoryOpen(true);
                }}
                size="sm"
                variant="naked"
              >
                <span
                  className={cn({ "sr-only": parent.subStories.length > 0 })}
                >
                  Add Sub{" "}
                  {getTermDisplay("storyTerm", {
                    capitalize: true,
                  })}
                </span>
              </Button>
            </Tooltip>
          </Flex>
        )}
      </Flex>
      <NewSubStory
        isOpen={isCreateSubStoryOpen}
        parentId={parent.id}
        setIsOpen={setIsCreateSubStoryOpen}
        teamId={parent.teamId}
      />
      {object?.substories && object.substories.length > 0 && showSuggestions ? (
        <Box>
          <Box className="mt-2 rounded-lg border-[0.5px] border-gray-100 dark:border-dark-100">
            {object.substories.map((substory) => {
              if (!substory?.title) return null;
              return (
                <RowWrapper
                  className="gap-6 px-2 last-of-type:border-b-0 md:px-4"
                  key={substory.title}
                >
                  <Flex align="center" className="flex-1" gap={2}>
                    <AiIcon className="shrink-0" />
                    <Text
                      className="line-clamp-1"
                      color={
                        selectedSubstories.has(substory.title)
                          ? undefined
                          : "muted"
                      }
                    >
                      {substory.title}
                    </Text>
                  </Flex>
                  <Checkbox
                    checked={selectedSubstories.has(substory.title)}
                    className="shrink-0"
                    onCheckedChange={() => {
                      toggleSelection(substory.title!);
                    }}
                  />
                </RowWrapper>
              );
            })}
          </Box>
          <Flex className="mt-2" gap={2} justify="end">
            <Button color="tertiary" onClick={handleCancel} variant="naked">
              Cancel
            </Button>
            <Button
              disabled={selectedSubstories.size === 0}
              onClick={handleAddSelected}
            >
              Create {selectedSubstories.size} Sub{" "}
              {getTermDisplay("storyTerm", {
                capitalize: true,
                variant: selectedSubstories.size === 1 ? "singular" : "plural",
              })}
            </Button>
          </Flex>
        </Box>
      ) : null}

      {isSubStoriesOpen && parent.subStories.length > 0 ? (
        <StoriesBoard
          className="mt-2 h-auto border-t-[0.5px] border-gray-100/60 pb-0 dark:border-dark-200"
          groupedStories={{
            groups: [
              {
                key: "none",
                totalCount: parent.subStories.length,
                stories: parent.subStories,
                loadedCount: parent.subStories.length,
                hasMore: false,
                nextPage: 1,
              },
            ],
            meta: {
              totalGroups: 1,
              filters: {},
              groupBy: "none",
              orderBy: "priority",
              orderDirection: "desc",
            },
          }}
          layout="list"
          rowClassName="pr-0 md:pr-0.5 md:pl-7"
          viewOptions={{
            groupBy: "none",
            orderBy: "priority",
            showEmptyGroups: false,
            displayColumns: ["ID", "Status", "Priority", "Assignee"],
          }}
        />
      ) : null}
    </>
  );
};
