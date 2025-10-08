import React from "react";
import { View } from "react-native";
import { Row, Skeleton } from "@/components/ui";

export const CommentSkeleton = () => {
  return (
    <View className="my-2.5 pl-0.5">
      <Row align="center">
        <Skeleton className="mr-2 size-6 rounded-full" />
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="ml-2 h-4 w-20" />
      </Row>
      <View className="mt-1 pl-8">
        <Skeleton className="h-4 w-6/12" />
      </View>
    </View>
  );
};
