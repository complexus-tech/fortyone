import { Flex, Badge, Button, Tooltip } from "ui";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { StoriesBoard } from "@/components/ui";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon } from "icons";
import { Story } from "@/modules/stories/types";
import { useState } from "react";
import { useStatuses } from "@/lib/hooks/statuses";
import { cn } from "lib";

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
  const { data: statuses = [] } = useStatuses();

  const completedStatus = statuses?.find(
    (status) => status?.category === "completed",
  );

  const completedStories =
    subStories.filter((story) => story.statusId === completedStatus?.id)
      ?.length ?? 0;

  return (
    <>
      <Flex
        align="center"
        justify={subStories.length > 0 ? "between" : "end"}
        className={cn({
          "border-b-[0.5px] border-gray-200 pb-2 dark:border-dark-200":
            !isSubStoriesOpen,
        })}
      >
        {subStories.length > 0 && (
          <Flex align="center" gap={2}>
            <Button
              color="tertiary"
              variant="naked"
              size="sm"
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
            >
              Sub stories
            </Button>
            <Badge color="tertiary" rounded="full" className="px-1.5">
              {completedStories}/{subStories.length} Done
            </Badge>
          </Flex>
        )}

        <Tooltip title={subStories.length > 0 ? "Add Sub Story" : null}>
          <Button
            color="tertiary"
            leftIcon={<PlusIcon />}
            size="sm"
            variant="naked"
            onClick={() => setIsCreateSubStoryOpen(true)}
          >
            <span className={cn({ "sr-only": subStories.length > 0 })}>
              Add Sub Story
            </span>
          </Button>
        </Tooltip>
      </Flex>
      <NewSubStory
        teamId={teamId}
        parentId={parentId}
        isOpen={isCreateSubStoryOpen}
        setIsOpen={setIsCreateSubStoryOpen}
      />
      {isSubStoriesOpen && subStories.length > 0 && (
        <StoriesBoard
          layout="list"
          stories={subStories}
          className="mt-2 h-auto border-t-[0.5px] border-gray-100/60 pb-0 dark:border-dark-200"
          viewOptions={{
            groupBy: "None",
            orderBy: "Priority",
            showEmptyGroups: false,
            displayColumns: ["ID", "Status", "Priority", "Assignee"],
          }}
        />
      )}
    </>
  );
};
