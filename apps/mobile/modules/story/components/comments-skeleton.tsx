import React from "react";
import { Col } from "@/components/ui";
import { CommentSkeleton } from "./comment-skeleton";

export const CommentsSkeleton = () => {
  return (
    <Col asContainer>
      {Array.from({ length: 6 }).map((_, index) => (
        <CommentSkeleton key={index} />
      ))}
    </Col>
  );
};
