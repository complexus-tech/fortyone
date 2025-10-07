import React from "react";
import { View } from "react-native";
import { Row, Col, Skeleton } from "@/components/ui";

type NotificationSkeletonProps = {
  count?: number;
};

export const NotificationsSkeleton = ({
  count = 8,
}: NotificationSkeletonProps) => {
  return (
    <View>
      {Array.from({ length: count }).map((_, index) => (
        <View
          key={index}
          style={{
            paddingVertical: 12,
            paddingHorizontal: 16,
          }}
        >
          <Row align="center" gap={2}>
            <Skeleton className="size-10 rounded-full shrink-0" />
            <Col flex={1} gap={2}>
              <Row justify="between" align="center">
                <Skeleton className="h-4 flex-1 mr-2" />
                <Skeleton className="h-4 w-10 shrink-0" />
              </Row>
              <Row align="center" justify="between" gap={2}>
                <Row align="center" gap={2} className="flex-1">
                  <Skeleton className="h-3.5 flex-1" />
                </Row>
                <Skeleton className="size-4" />
              </Row>
            </Col>
          </Row>
        </View>
      ))}
    </View>
  );
};
