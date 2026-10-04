"use client";
import { useParams } from "next/navigation";
import { Box, Button, Text } from "ui";
import { StoriesBoard, type StoriesLayout } from "@/components/ui";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { useMyStoriesGrouped } from "@/modules/stories/public/my-work";
import { getMyWorkScopeFilterParams } from "@/shared/story/my-work-scope";
import type { MyWorkViewScope } from "@/shared/story/view-configuration";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useTeamOptions } from "./provider";

const TeamStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions, setViewOptions, filters } = useTeamOptions();
  const { data: groupedStories, isPending } = useTeamStoriesGrouped(
    teamId,
    viewOptions.groupBy,
    {
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      ...getGroupedStoryFilterParams(filters),
      showSubStories: viewOptions.showSubStories ? true : undefined,
      teamIds: [teamId],
    },
  );

  if (isPending) {
    return <StoriesSkeleton className="h-full" layout={layout} />;
  }

  return (
    <StoriesBoard
      className="h-full"
      groupedStories={groupedStories}
      layout={layout}
      setViewOptions={setViewOptions}
      viewOptions={viewOptions}
    />
  );
};

const MyWorkScopedStories = ({
  layout,
  scope,
}: {
  layout: StoriesLayout;
  scope: MyWorkViewScope;
}) => {
  const { viewOptions, setViewOptions, filters } = useTeamOptions();
  const query = useMyStoriesGrouped(
    viewOptions.groupBy,
    {
      ...getMyWorkScopeFilterParams(filters, scope),
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      showSubStories: viewOptions.showSubStories ? true : undefined,
    },
    { accountScoped: true },
  );
  if (query.isError)
    return (
      <Box className="space-y-3 p-6">
        <Text role="alert">Tasks for this view could not be loaded.</Text>
        <Button
          color="tertiary"
          onClick={() => void query.refetch()}
          variant="outline"
        >
          Try again
        </Button>
      </Box>
    );
  if (!query.isFetchedAfterMount)
    return <StoriesSkeleton className="h-full" layout={layout} />;
  return (
    <StoriesBoard
      className="h-full"
      groupedStories={query.data}
      layout={layout}
      setViewOptions={setViewOptions}
      viewOptions={viewOptions}
    />
  );
};

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { viewMetadata } = useTeamOptions();
  return viewMetadata.scope ? (
    <MyWorkScopedStories layout={layout} scope={viewMetadata.scope} />
  ) : (
    <TeamStories layout={layout} />
  );
};
