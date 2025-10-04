import React from "react";
import { SectionList, View, ActivityIndicator } from "react-native";
import { Story, Text } from "@/components/ui";
import { StoryGroupHeader } from "./story-group-header";
import type { MyWorkSection } from "../types";
import { colors } from "@/constants";

type GroupedStoriesListProps = {
  sections: MyWorkSection[];
  isLoading?: boolean;
};

export const GroupedStoriesList = ({
  sections,
  isLoading = false,
}: GroupedStoriesListProps) => {
  if (isLoading) {
    return (
      <View
        style={{
          flex: 1,
          justifyContent: "center",
          alignItems: "center",
          paddingTop: 48,
        }}
      >
        <ActivityIndicator size="large" color={colors.gray.DEFAULT} />
      </View>
    );
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
    <SectionList
      sections={sections}
      keyExtractor={(item) => item.id}
      renderItem={({ item }) => <Story {...item} />}
      renderSectionHeader={({ section }) => (
        <StoryGroupHeader title={section.title} color={section.color} />
      )}
      stickySectionHeadersEnabled={true}
      showsVerticalScrollIndicator={false}
      ItemSeparatorComponent={() => <View />}
      contentContainerStyle={{ paddingBottom: 24 }}
    />
  );
};
