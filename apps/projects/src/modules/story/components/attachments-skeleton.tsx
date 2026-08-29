"use client";
import { Flex, Skeleton, Wrapper } from "ui";
import { cn } from "lib";

export const AttachmentsSkeleton = ({ className }: { className?: string }) => {
  return (
    <Wrapper
      className={cn("border-border mt-4 border-b-[0.5px] pb-2", className)}
    >
      <Flex align="center" justify="between">
        <Skeleton className="h-6 w-32 rounded" />
        <Skeleton className="h-8 w-8 rounded-full" />
      </Flex>
    </Wrapper>
  );
};
