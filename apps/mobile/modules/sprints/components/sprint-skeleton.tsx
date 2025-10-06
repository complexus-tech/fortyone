import React from "react";
import { View } from "react-native";
import { Row, Col } from "@/components/ui";
import { Skeleton } from "@/components/ui/skeleton";

export const SprintSkeleton = () => {
  return (
    <View
      style={{
        paddingVertical: 12,
        paddingHorizontal: 16,
      }}
    >
      <Row align="center" justify="between" gap={3}>
        <Row align="center" gap={3} className="w-8/12">
          <Row className="bg-gray-100 dark:bg-dark-200 rounded-lg p-1.5">
            <Skeleton style={{ width: 20, height: 20 }} />
          </Row>
          <Col gap={1}>
            <Skeleton style={{ width: 120, height: 16 }} />
            <Skeleton style={{ width: 80, height: 14 }} />
          </Col>
        </Row>
        <Skeleton style={{ width: 70, height: 24, borderRadius: 12 }} />
      </Row>
    </View>
  );
};
