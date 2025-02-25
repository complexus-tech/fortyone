import { Flex, Badge, Button, Tooltip } from "ui";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon } from "icons";
import { useState } from "react";
import { cn } from "lib";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/modules/stories/types";
import { useTeamStatuses } from "@/lib/hooks/statuses";

export const SubStories = ({
  subStories,
  parentId,
  teamId,
  setIsSubStoriesOpen,
  isSubStoriesOpen,
}: {
  subStories: Story[];
  parentId: string;
  teamId: string;
  setIsSubStoriesOpen: (value: boolean) => void;
  isSubStoriesOpen: boolean;
}) => {
  const [isCreateSubStoryOpen, setIsCreateSubStoryOpen] = useState(false);
  const { data: statuses = [] } = useTeamStatuses(teamId);

  const completedStatus = statuses.find(
    (status) => status.category === "completed",
  );

  const completedStories = subStories.filter(
    (story) => story.statusId === completedStatus?.id,
  ).length;

  return (
    <>
      <Flex
        align="center"
        className={cn({
          "border-b-[0.5px] border-gray-200 pb-2 dark:border-dark-200":
            !isSubStoriesOpen,
        })}
        justify={subStories.length > 0 ? "between" : "end"}
      >
        {subStories.length > 0 && (
          <Flex align="center" gap={2}>
            <Button
              color="tertiary"
              onClick={() => {
                setIsSubStoriesOpen(!isSubStoriesOpen);
              }}
              rightIcon={
                isSubStoriesOpen ? (
                  <ArrowDownIcon className="h-4 w-auto" />
                ) : (
                  <ArrowUpIcon className="h-4 w-auto" />
                )
              }
              size="sm"
              variant="naked"
            >
              Sub stories
            </Button>
            <Badge className="px-1.5" color="tertiary" rounded="full">
              {completedStories}/{subStories.length} Done
            </Badge>
          </Flex>
        )}

        <Tooltip title={subStories.length > 0 ? "Add Sub Story" : null}>
          <Button
            color="tertiary"
            leftIcon={<PlusIcon />}
            onClick={() => {
              setIsCreateSubStoryOpen(true);
            }}
            size="sm"
            variant="naked"
          >
            <span className={cn({ "sr-only": subStories.length > 0 })}>
              Add Sub Story
            </span>
          </Button>
        </Tooltip>
      </Flex>
      <NewSubStory
        isOpen={isCreateSubStoryOpen}
        parentId={parentId}
        setIsOpen={setIsCreateSubStoryOpen}
        teamId={teamId}
      />
      {isSubStoriesOpen && subStories.length > 0 ? (
        <StoriesBoard
          className="mt-2 h-auto border-t-[0.5px] border-gray-100/60 pb-0 dark:border-dark-200"
          layout="list"
          stories={subStories}
          viewOptions={{
            groupBy: "None",
            orderBy: "Priority",
            showEmptyGroups: false,
            displayColumns: ["ID", "Status", "Priority", "Assignee"],
          }}
        />
      ) : null}
    </>
  );
};
