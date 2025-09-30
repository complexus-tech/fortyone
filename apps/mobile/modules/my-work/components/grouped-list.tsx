import React from "react";
import { SectionList, View } from "react-native";
import { Story } from "@/components/ui";
import { StoryGroupHeader } from "./story-group-header";

type StoryType = {
  id: string;
  title: string;
  status: {
    id: string;
    name: string;
    color: string;
  };
  priority: "Urgent" | "High" | "Medium" | "Low" | "No Priority";
  assignee?: {
    id: string;
    name: string;
    avatarUrl?: string;
  };
};

type StorySection = {
  title: string;
  color?: string;
  data: StoryType[];
};

type GroupedStoriesListProps = {
  sections: StorySection[];
  onStoryPress?: (storyId: string) => void;
};

export const GroupedStoriesList = ({
  sections,
  onStoryPress,
}: GroupedStoriesListProps) => {
  return (
    <SectionList
      sections={sections}
      keyExtractor={(item) => item.id}
      renderItem={({ item }) => <Story story={item} onPress={onStoryPress} />}
      renderSectionHeader={({ section }) => (
        <StoryGroupHeader title={section.title} />
      )}
      stickySectionHeadersEnabled={true}
      showsVerticalScrollIndicator={false}
      ItemSeparatorComponent={() => <View />}
    />
  );
};
