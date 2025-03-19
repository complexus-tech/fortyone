"use client";
import { Flex, Text, Wrapper } from "ui";
import { ClockIcon } from "icons";
import { useActivities } from "@/lib/hooks/activities";
import { Activity } from "@/components/ui/activity";

export const Activities = () => {
  const { data: activities = [] } = useActivities();
  return (
    <Wrapper>
      <Flex align="center" className="mb-5" justify="between">
        <Text fontSize="lg">Recent activities</Text>
      </Flex>
      {activities.length === 0 ? (
        <Flex
          align="center"
          className="h-[25rem]"
          direction="column"
          gap={3}
          justify="center"
        >
          <ClockIcon className="h-24 opacity-70" />
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
