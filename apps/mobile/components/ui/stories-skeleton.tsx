import React from "react";
import { ScrollView } from "react-native";
import { StorySkeleton } from "./story-skeleton";

type StoriesSkeletonProps = {
  count?: number;
};

export const StoriesSkeleton = ({ count = 5 }: StoriesSkeletonProps) => {
  return (
    <ScrollView
      showsVerticalScrollIndicator={false}
      contentContainerStyle={{ paddingBottom: 24 }}
    >
      {Array.from({ length: count }).map((_, index) => (
        <StorySkeleton key={index} />
      ))}
    </ScrollView>
  );
};
