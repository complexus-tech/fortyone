import { Flex, Badge, Button, Tooltip } from "ui";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon, SubStoryIcon } from "icons";
import { useState } from "react";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/modules/stories/types";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTerminology, useUserRole } from "@/hooks";

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
  const { getTermDisplay } = useTerminology();
  const [isCreateSubStoryOpen, setIsCreateSubStoryOpen] = useState(false);
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { userRole } = useUserRole();
  const completedStatus = statuses.find(
    (status) => status.category === "completed",
  );

  const completedStories = subStories.filter(
    (story) => story.statusId === completedStatus?.id,
  ).length;

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
          "border-b-[0.5px] border-gray-200 pb-2 dark:border-dark-200":
            !isSubStoriesOpen,
        })}
        justify={subStories.length > 0 ? "between" : "end"}
      >
        {subStories.length > 0 && (
          <Flex align="center" gap={2}>
            <Button
              color="tertiary"
              leftIcon={<SubStoryIcon />}
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
              Sub {getTermDisplay("storyTerm", { variant: "plural" })}
            </Button>
            <Badge className="px-1.5" color="tertiary" rounded="full">
              {completedStories}/{subStories.length} Done
            </Badge>
          </Flex>
        )}

        {userRole !== "guest" && (
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
                Add Sub{" "}
                {getTermDisplay("storyTerm", {
                  capitalize: true,
                })}
              </span>
            </Button>
          </Tooltip>
        )}
      </Flex>

      {isSubStoriesOpen && subStories.length > 0 ? (
        <StoriesBoard
          className="mt-2 h-auto border-t-[0.5px] border-gray-100/60 pb-0 dark:border-dark-200"
          groupedStories={{
            groups: [
              {
                key: "none",
                totalCount: subStories.length,
                stories: subStories,
                loadedCount: subStories.length,
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
      <NewSubStory
        isOpen={isCreateSubStoryOpen}
        parentId={parentId}
        setIsOpen={setIsCreateSubStoryOpen}
        teamId={teamId}
      />
    </>
  );
};
