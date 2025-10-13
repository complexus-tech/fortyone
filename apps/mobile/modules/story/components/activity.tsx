import { Col, Tabs, Text } from "@/components/ui";
import { DetailedStory } from "@/modules/stories/types";
import React, { useState } from "react";
import { useStoryActivitiesInfinite } from "../hooks/use-story-activities";
import { ActivityItem } from "./activity-item";
import { Comments } from "./comments";
import { ActivitiesSkeleton } from "./activities-skeleton";

export const Activity = ({ story }: { story: DetailedStory }) => {
  const [activeTab, setActiveTab] = useState("updates");
  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
    isPending,
  } = useStoryActivitiesInfinite(story.id);

  if (isPending) {
    return <ActivitiesSkeleton />;
  }

  const allActivities =
    infiniteData?.pages.flatMap((page) => page.activities) ?? [];

  const handleLoadMore = () => {
    fetchNextPage();
  };

  return (
    <Tabs
      defaultValue={activeTab}
      onValueChange={(value) => setActiveTab(value)}
    >
      <Tabs.List className="mb-2">
        <Tabs.Tab value="updates" className="py-1.5 px-4 rounded-[10px]">
          Updates
        </Tabs.Tab>
        <Tabs.Tab value="comments" className="py-1.5 px-4 rounded-[10px]">
          Comments
        </Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="comments">
        <Comments storyId={story.id} />
      </Tabs.Panel>
      <Tabs.Panel value="updates">
        {allActivities.length === 0 ? (
          <Text>No updates available</Text>
        ) : (
          <Col asContainer>
            {allActivities.map((activity, index) => (
              <ActivityItem
                key={activity.id}
                {...activity}
                teamId={story.teamId}
                isTimeShown={index === allActivities.length - 1 || index === 0}
              />
            ))}
            {hasNextPage && (
              <Text
                onPress={handleLoadMore}
                className="mt-4 pl-7"
                fontSize="sm"
              >
                {isFetchingNextPage ? "Loading..." : "Load more updates"}
              </Text>
            )}
          </Col>
        )}
      </Tabs.Panel>
    </Tabs>
  );
};
