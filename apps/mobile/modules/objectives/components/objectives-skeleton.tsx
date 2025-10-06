import React from "react";
import { ScrollView } from "react-native";
import { ObjectiveSkeleton } from "./objective-skeleton";

type ObjectivesSkeletonProps = {
  count?: number;
};

export const ObjectivesSkeleton = ({ count = 5 }: ObjectivesSkeletonProps) => {
  return (
    <ScrollView showsVerticalScrollIndicator={false}>
      {Array.from({ length: count }).map((_, index) => (
        <ObjectiveSkeleton key={index} />
      ))}
    </ScrollView>
  );
};
