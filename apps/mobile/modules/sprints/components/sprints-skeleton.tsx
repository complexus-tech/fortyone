import React from "react";
import { ScrollView } from "react-native";
import { SprintSkeleton } from "./sprint-skeleton";

type SprintsSkeletonProps = {
  count?: number;
};

export const SprintsSkeleton = ({ count = 5 }: SprintsSkeletonProps) => {
  return (
    <ScrollView showsVerticalScrollIndicator={false}>
      {Array.from({ length: count }).map((_, index) => (
        <SprintSkeleton key={index} />
      ))}
    </ScrollView>
  );
};
