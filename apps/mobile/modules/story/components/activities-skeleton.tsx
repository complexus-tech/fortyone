import React from "react";
import { Col } from "@/components/ui";
import { ActivitySkeleton } from "./activity-skeleton";

export const ActivitiesSkeleton = () => {
  return (
    <Col asContainer>
      {Array.from({ length: 3 }).map((_, index) => (
        <ActivitySkeleton key={index} />
      ))}
    </Col>
  );
};
