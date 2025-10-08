import React from "react";
import { Col, Text } from "@/components/ui";
import { useStoryCommentsInfinite } from "../hooks/use-story-comments";
import { CommentItem } from "./comment-item";
import { CommentsSkeleton } from "./comments-skeleton";

export const Comments = ({ storyId }: { storyId: string }) => {
  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
    isPending,
  } = useStoryCommentsInfinite(storyId);

  if (isPending) {
    return <CommentsSkeleton />;
  }

  const allComments =
    infiniteData?.pages.flatMap((page) => page.comments) ?? [];

  const handleLoadMore = () => {
    fetchNextPage();
  };

  return (
    <Col asContainer>
      {allComments.length === 0 ? (
        <Text>No comments yet</Text>
      ) : (
        <>
          {allComments.map((comment) => (
            <CommentItem key={comment.id} {...comment} />
          ))}
          {hasNextPage && (
            <Text onPress={handleLoadMore} className="mt-4 pl-7" fontSize="sm">
              {isFetchingNextPage ? "Loading..." : "Load more comments"}
            </Text>
          )}
        </>
      )}
    </Col>
  );
};
