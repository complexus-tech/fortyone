import React from "react";
import { View } from "react-native";
import { Row } from "./row";
import { Skeleton } from "./skeleton";

export const StorySkeleton = () => {
  return (
    <View
      style={{
        paddingVertical: 11,
        paddingHorizontal: 16,
      }}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" gap={2} className="flex-1">
          <Skeleton style={{ width: 18, height: 18 }} />
          <Skeleton className="flex-1" style={{ height: 16 }} />
        </Row>
        <Row align="center" gap={3}>
          <Skeleton style={{ width: 12, height: 12, borderRadius: 6 }} />
          <Skeleton style={{ width: 32, height: 32, borderRadius: 16 }} />
        </Row>
      </Row>
    </View>
  );
};
