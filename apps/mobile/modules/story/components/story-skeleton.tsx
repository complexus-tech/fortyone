import React from "react";
import { ScrollView } from "react-native";
import { Row, Col, Skeleton } from "@/components/ui";
import { ActivitiesSkeleton } from "./activities-skeleton";

const TitleSkeleton = () => {
  return (
    <Col asContainer>
      <Skeleton className="h-5 w-3/4 my-2 rounded-lg" />
      <Skeleton className="h-5 w-1/2 rounded-lg" />
    </Col>
  );
};

const PropertiesSkeleton = () => {
  return (
    <Row asContainer>
      <Col className="my-7">
        <Skeleton className="h-4 w-20 mb-4 rounded" />
        <Row wrap gap={2}>
          <Skeleton className="h-6 w-24 rounded-full" />
          <Skeleton className="h-6 w-28 rounded-full" />
          <Skeleton className="h-6 w-32 rounded-full" />
          <Skeleton className="h-7 w-24 rounded-full" />
          <Skeleton className="h-6 w-24 rounded-full" />
          <Skeleton className="h-6 w-28 rounded-full" />
          <Skeleton className="h-6 w-26 rounded-full" />
          <Skeleton className="h-6 w-24 rounded-full" />
        </Row>
      </Col>
    </Row>
  );
};

const DescriptionSkeleton = () => {
  return (
    <Col gap={2} asContainer className="mb-6">
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-4/5" />
      <Skeleton className="h-4 w-3/5" />
    </Col>
  );
};

export const StorySkeleton = () => {
  return (
    <ScrollView
      className="flex-1"
      contentContainerStyle={{ paddingBottom: 100 }}
    >
      <TitleSkeleton />
      <PropertiesSkeleton />
      <DescriptionSkeleton />
      <ActivitiesSkeleton />
    </ScrollView>
  );
};
