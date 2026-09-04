"use client";
import { Flex, Text, Wrapper } from "ui";
import { useActivities } from "@/lib/hooks/activities";
import { Activity } from "@/components/ui/activity";
import { ActivityEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { useSummaryDateFilters } from "@/modules/summary/hooks/summary-date-filters";
import { ActivitiesSkeleton } from "./activities-skeleton";

export const Activities = () => {
  const filters = useSummaryDateFilters();
  const { data: activities = [], isPending } = useActivities(filters);
  if (isPending) {
    return <ActivitiesSkeleton />;
  }
  return (
    <Wrapper className="min-h-100 md:min-h-120">
      <Flex align="center" className="mb-5" justify="between">
        <Text fontSize="lg">Your Activities</Text>
      </Flex>
      {activities.length === 0 ? (
        <Flex
          align="center"
          className="h-100"
          direction="column"
          gap={3}
          justify="center"
        >
          <ActivityEmptyIllustration />
          <Text color="muted">You do not have any activities yet.</Text>
        </Flex>
      ) : (
        activities.map((activity) => (
          <Activity key={activity.id} {...activity} />
        ))
      )}
    </Wrapper>
  );
};
