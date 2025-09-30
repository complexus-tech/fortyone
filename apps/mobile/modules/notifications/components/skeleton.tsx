import React from "react";
import { View } from "react-native";
import { Row, Col, Skeleton } from "@/components/ui";

type NotificationSkeletonProps = {
  count?: number;
};

export const NotificationSkeleton = ({
  count = 8,
}: NotificationSkeletonProps) => {
  return (
    <View className="flex-1">
      {Array.from({ length: count }).map((_, index) => (
        <View
          key={index}
          className="border-b border-gray-100/60 px-4 py-[12px]"
        >
          <Col className="gap-2.5">
            <Row justify="between" align="center">
              <Skeleton className="h-4 flex-1 mr-8" />
              <Skeleton className="h-4 w-10" />
            </Row>
            <Row align="center" justify="between">
              <Row align="center" gap={2} className="flex-1">
                <Skeleton className="size-7 rounded-full" />
                <Skeleton className="h-3.5 flex-1 mr-4" />
              </Row>
              <Skeleton className="size-4" />
            </Row>
          </Col>
        </View>
      ))}
    </View>
  );
};
