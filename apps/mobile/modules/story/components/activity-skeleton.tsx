import React from "react";
import { View } from "react-native";
import { Row, Skeleton } from "@/components/ui";

export const ActivitySkeleton = () => {
  return (
    <View className="my-2.5 pl-0.5">
      <Row align="center">
        <Skeleton className="mr-2 size-6 rounded-full" />
        <Skeleton className="h-4 w-32" />
        <Skeleton className="ml-2 h-4 w-1/2" />
      </Row>
    </View>
  );
};
