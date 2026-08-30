import { Flex, Badge, Button, Tooltip, Box } from "ui";
import {
  ArrowDown2Icon,
  ArrowUp2Icon,
  PlusIcon,
  SubStoryIcon,
  AiIcon,
} from "icons";
import { useState } from "react";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { RowWrapper, StoriesBoard } from "@/components/ui";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTerminology, useUserRole, useWorkspacePath } from "@/hooks";
import { Thinking } from "@/components/ui/chat/thinking";
import { useChatContext } from "@/context/chat-context";
import { useTeams } from "@/modules/teams/hooks/teams";
import type { DetailedStory } from "../types";
import { StoryRelationshipPicker } from "./story-relationship-picker";
import { SubstorySuggestions } from "./substory-suggestions";
import { useSubstorySuggestions } from "./use-substory-suggestions";

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
  const { data: statuses = [] } = useTeamStatuses(parent.teamId);
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === parent.teamId);
  const { openChat } = useChatContext();
  const { userRole } = useUserRole();
  const { workspaceSlug } = useWorkspacePath();
  const completedStatus = statuses.find(
    (status) => status.category === "completed",
  );
  const defaultStatus =
    statuses.find((status) => status.isDefault) || statuses.at(0);
  const storyTerms = {
    plural: getTermDisplay("storyTerm", { variant: "plural" }),
    pluralCapitalized: getTermDisplay("storyTerm", {
      capitalize: true,
      variant: "plural",
    }),
    singular: getTermDisplay("storyTerm"),
    singularCapitalized: getTermDisplay("storyTerm", {
      capitalize: true,
      variant: "singular",
    }),
  };
  const {
    canCreateSuggestedSubstories,
    cancelSuggestions,
    createSelectedSubstories,
    isCreatingSuggestedSubstories,
    isLoadingSuggestions,
    isShowingSuggestionError,
    requestSuggestions,
    selectedSubstories,
    showSuggestions,
    suggestedSubstories,
    toggleSelectedSubstory,
  } = useSubstorySuggestions({
    defaultStatusId: defaultStatus?.id,
    storyId: parent.id,
    teamId: parent.teamId,
    workspaceSlug,
  });

  const completedStories = parent.subStories.filter(
    (story) => story.statusId === completedStatus?.id,
  ).length;

  useHotkeys("c", () => {
    if (userRole !== "guest") {
      setIsCreateSubStoryOpen(true);
    }
  });

  return (
    <Box className="mt-6">
      <Flex
        align="center"
        className={cn("@container w-full min-w-0 pb-1.5", {
          "border-border border-b-[0.5px]": parent.subStories.length > 0,
        })}
        gap={2}
        justify={parent.subStories.length > 0 ? "between" : "end"}
        wrap
      >
        {parent.subStories.length > 0 && (
          <Flex align="center" className="justify-end" gap={2} wrap>
            <Button
              className="font-semibold"
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
              Sub {getTermDisplay("storyTerm", { variant: "plural" })}{" "}
              {parent.subStories.length > 0
                ? `(${parent.subStories.length})`
                : ""}
            </Button>
            <Badge className="px-1.5" color="tertiary">
              {completedStories}/{parent.subStories.length} Done
            </Badge>
          </Flex>
        )}

        {userRole !== "guest" && (
          <Flex
            align="center"
            className="ml-auto max-w-full min-w-0 justify-end"
            gap={2}
            wrap
          >
            <Button
              className="shrink-0"
              color="tertiary"
              leftIcon={<AiIcon className="text-primary dark:text-primary" />}
              onClick={() => {
                openChat(
                  `Improve the description of ${getTermDisplay("storyTerm")} ${team?.code}-${parent.sequenceId}`,
                );
              }}
              size="sm"
              variant="naked"
            >
              <span className="@lg:hidden">Improve</span>
              <span className="hidden @lg:inline">Improve description</span>
            </Button>
            <Button
              className="shrink-0"
              color="tertiary"
              disabled={isLoadingSuggestions}
              leftIcon={<AiIcon className="text-primary dark:text-primary" />}
              onClick={requestSuggestions}
              size="sm"
              variant="naked"
            >
              {isLoadingSuggestions ? (
                <Thinking message="Maya is thinking" />
              ) : (
                <>
                  <span className="@lg:hidden">Suggest</span>
                  <span className="hidden @lg:inline">
                    Suggest sub{" "}
                    {getTermDisplay("storyTerm", {
                      variant: "plural",
                    })}
                  </span>
                </>
              )}
            </Button>
            <StoryRelationshipPicker
              currentStoryId={parent.id}
              currentStoryTitle={parent.title}
              existingAssociationStoryIds={parent.associations.map(
                (association) => association.story.id,
              )}
              teamCode={parent.teamCode}
              teamId={parent.teamId}
            />
            <Tooltip title={`Add sub ${getTermDisplay("storyTerm")}`}>
              <Button
                aria-label={`Add sub ${getTermDisplay("storyTerm")}`}
                className="aspect-square shrink-0 justify-center px-0 @lg:aspect-auto @lg:px-2"
                color="tertiary"
                onClick={() => {
                  setIsCreateSubStoryOpen(true);
                }}
                size="sm"
                variant="naked"
              >
                <PlusIcon />
                <span className="hidden @lg:inline">
                  Add sub {getTermDisplay("storyTerm")}
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
      <SubstorySuggestions
        SuggestionRow={RowWrapper}
        canCreateSuggestedSubstories={canCreateSuggestedSubstories}
        isCreatingSuggestedSubstories={isCreatingSuggestedSubstories}
        isShowingSuggestionError={isShowingSuggestionError}
        onCancelSuggestions={cancelSuggestions}
        onCreateSelectedSubstories={createSelectedSubstories}
        onRequestSuggestions={requestSuggestions}
        onToggleSelectedSubstory={toggleSelectedSubstory}
        selectedSubstories={selectedSubstories}
        showSuggestions={showSuggestions}
        suggestedSubstories={suggestedSubstories}
        terms={storyTerms}
      />

      {isSubStoriesOpen && parent.subStories.length > 0 ? (
        <StoriesBoard
          className="h-auto pb-0"
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
            orderDirection: "desc",
            showEmptyGroups: false,
            showSubStories: false,
            displayColumns: [
              "ID",
              "Status",
              "Assignee",
              "Estimate",
              "Time needed",
              "Priority",
            ],
          }}
        />
      ) : null}
    </Box>
  );
};
