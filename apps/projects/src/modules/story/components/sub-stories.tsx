import { Flex, Badge, Button } from "ui";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { StoriesBoard } from "@/components/ui";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon } from "icons";
import { useLocalStorage } from "@/hooks";
import { Story } from "@/modules/stories/types";
import { useState } from "react";

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

  return (
    <>
      <Flex align="center" justify="between">
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
            1/{subStories.length} Done
          </Badge>
        </Flex>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          size="sm"
          variant="naked"
          onClick={() => setIsCreateSubStoryOpen(true)}
        >
          Add Sub Story
        </Button>
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
          className="mt-2 h-auto border-t border-gray-100/60 pb-0 dark:border-dark-100/80"
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
