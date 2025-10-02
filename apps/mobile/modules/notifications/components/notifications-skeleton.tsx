import React from "react";
import { View } from "react-native";
import { Row, Skeleton } from "@/components/ui";

export const NotificationsSkeleton = () => {
  return (
    <View>
      {Array.from({ length: 8 }).map((_, index) => (
        <View key={index} className="px-4 py-3 border-b border-gray-100">
          <Row align="center" justify="between" className="mb-2">
            <Skeleton className="h-4 w-3/5" />
            <Skeleton className="h-4 w-10" />
          </Row>

          <Row align="center" justify="between">
            <Row align="center" gap={2}>
              <Skeleton className="size-8 rounded-full" />
              <Skeleton className="h-4 w-40" />
            </Row>
            <Skeleton className="size-4" />
          </Row>
        </View>
      ))}
    </View>
  );
};
