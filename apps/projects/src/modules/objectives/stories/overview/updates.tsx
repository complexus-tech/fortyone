import { Box, Flex, Text, Button, Skeleton } from "ui";
import { useObjectiveActivitiesInfinite } from "../../hooks/use-objective-activities";
import { ObjectiveActivityComponent } from "../../components/objective-activity";

export const Updates = ({ objectiveId }: { objectiveId: string }) => {
  const {
    data: infiniteData,
    isPending,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useObjectiveActivitiesInfinite(objectiveId);

  const allActivities =
    infiniteData?.pages.flatMap((page) => page.activities) ?? [];

  const handleLoadMore = () => {
    fetchNextPage();
  };

  if (isPending) {
    return (
      <Box className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Box className="flex gap-3" key={i}>
            <Skeleton className="h-8 w-8 rounded-full" />
            <Box className="flex-1 space-y-2">
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-4 w-48" />
            </Box>
          </Box>
        ))}
      </Box>
    );
  }

  return (
    <Box className="min-h-44 pl-1">
      <Flex direction="column">
        {allActivities.length === 0 && (
          <Text className="text-muted-foreground text-sm">
            No updates available
          </Text>
        )}
        {allActivities.map((activity) => (
          <ObjectiveActivityComponent key={activity.id} {...activity} />
        ))}
      </Flex>
      {hasNextPage ? (
        <Box className="mt-2">
          <Button
            className="ml-6 px-3 text-[0.95rem]"
            color="tertiary"
            disabled={isFetchingNextPage}
            onClick={handleLoadMore}
            size="sm"
            variant="naked"
          >
            {isFetchingNextPage ? "Loading..." : "Load more updates"}
          </Button>
        </Box>
      ) : null}
    </Box>
  );
};
