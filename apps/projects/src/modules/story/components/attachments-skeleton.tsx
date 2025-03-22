"use client";
import { Flex, Skeleton, Wrapper } from "ui";
import { cn } from "lib";

export const AttachmentsSkeleton = ({ className }: { className?: string }) => {
  return (
    <Wrapper
      className={cn(
        "mb-4 mt-2.5 border-t border-gray-100/60 pt-2.5 dark:border-dark-100/80",
        className,
      )}
    >
      <Flex align="center" className="mb-4" justify="between">
        <Skeleton className="h-6 w-32 rounded" />
        <Skeleton className="h-8 w-8 rounded-full" />
      </Flex>
    </Wrapper>
  );
};
