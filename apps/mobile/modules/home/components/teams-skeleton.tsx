import React from "react";
import { View } from "react-native";
import { Col, Skeleton } from "@/components/ui";

export const TeamsSkeleton = () => {
  return (
    <Col asContainer>
      <Skeleton className="h-5 w-24 mb-4" />
      {Array.from({ length: 3 }).map((_, index) => (
        <View key={index} className="py-3.5 pl-0.5 min-h-[44px]">
          <View className="flex-row items-center justify-between">
            <View className="flex-row items-center">
              <Skeleton className="size-3 rounded-full mr-2" />
              <Skeleton className="h-4 w-24" />
            </View>
            <Skeleton className="size-3" />
          </View>
        </View>
      ))}
    </Col>
  );
};
