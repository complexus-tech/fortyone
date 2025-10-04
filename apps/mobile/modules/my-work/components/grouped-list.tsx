import React from "react";
import { ScrollView, View } from "react-native";
import { Text, StoriesListSkeleton } from "@/components/ui";
import { SectionWithLoadMore } from "./section-with-load-more";
import type { MyWorkSection } from "../types";
import type { GroupStoryParams } from "@/modules/stories/types";

type GroupedStoriesListProps = {
  sections: MyWorkSection[];
  groupFilters: Omit<GroupStoryParams, "groupKey">;
  isLoading?: boolean;
};

export const GroupedStoriesList = ({
  sections,
  groupFilters,
  isLoading = false,
}: GroupedStoriesListProps) => {
  if (isLoading) {
    return <StoriesListSkeleton />;
  }

  if (sections.length === 0) {
    return (
      <View
        style={{
          flex: 1,
          justifyContent: "center",
          alignItems: "center",
          paddingTop: 48,
        }}
      >
        <Text fontSize="lg" color="muted">
          No stories found
        </Text>
      </View>
    );
  }

  return (
    <ScrollView
      showsVerticalScrollIndicator={false}
      contentContainerStyle={{ paddingBottom: 24 }}
    >
      {sections.map((section) => (
        <SectionWithLoadMore
          key={section.key}
          section={section}
          groupFilters={groupFilters}
        />
      ))}
    </ScrollView>
  );
};
