import React from "react";
import { View, ScrollView } from "react-native";
import { StorySkeleton } from "./story-skeleton";
import { Skeleton } from "./skeleton";

type StoriesListSkeletonProps = {
  sectionsCount?: number;
  storiesPerSection?: number;
};

export const StoriesListSkeleton = ({
  sectionsCount = 3,
  storiesPerSection = 3,
}: StoriesListSkeletonProps) => {
  return (
    <ScrollView
      showsVerticalScrollIndicator={false}
      contentContainerStyle={{ paddingBottom: 24 }}
    >
      {Array.from({ length: sectionsCount }).map((_, sectionIndex) => (
        <View key={sectionIndex}>
          {/* Section Header Skeleton */}
          <View
            style={{
              paddingBottom: 8,
              paddingTop: 12,
              paddingHorizontal: 16,
            }}
          >
            <Skeleton style={{ width: 120, height: 16 }} />
          </View>

          {/* Stories in Section */}
          {Array.from({ length: storiesPerSection }).map((_, storyIndex) => (
            <StorySkeleton key={`${sectionIndex}-${storyIndex}`} />
          ))}
        </View>
      ))}
    </ScrollView>
  );
};
