import React from "react";
import { Row, Skeleton } from "@/components/ui";

export const HeaderSkeleton = () => {
  return (
    <Row align="center" justify="between" className="mb-3">
      <Row align="center" gap={2}>
        <Skeleton className="size-10 rounded-full" />
        <Skeleton className="h-8 w-32" />
      </Row>
      <Skeleton className="size-7" />
    </Row>
  );
};
