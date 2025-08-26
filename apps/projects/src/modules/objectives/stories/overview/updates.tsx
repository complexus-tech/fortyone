import { useObjectiveActivitiesInfinite } from "../../hooks/use-objective-activities";

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
  return (
    <div>
      updates
      {allActivities.map((activity) => (
        <div key={activity.id}>{activity.field}</div>
      ))}
    </div>
  );
};
