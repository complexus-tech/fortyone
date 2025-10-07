import React from "react";
import { View } from "react-native";
import { Col, Row, Skeleton, Wrapper } from "@/components/ui";

export const OverviewSkeleton = () => {
  return (
    <Col asContainer className="mb-6">
      <Skeleton className="h-4 w-64 mb-4" />
      <Row wrap gap={3}>
        {Array.from({ length: 4 }).map((_, index) => (
          <View key={index} className="w-[48.5%]">
            <Wrapper>
              <View className="flex-col gap-2">
                <Skeleton className="h-8 w-12" />
                <Skeleton className="h-4 w-20" />
              </View>
              <Skeleton
                className="size-5 rounded-full"
                style={{ position: "absolute", top: 14, right: 16 }}
              />
            </Wrapper>
          </View>
        ))}
      </Row>
    </Col>
  );
};
