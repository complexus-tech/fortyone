"use client";
import { Box } from "ui";
import { cn } from "lib";
import { MainDetailsSkeleton } from "./main-details-skeleton";
import { OptionsSkeleton } from "./options-skeleton";

export const StorySkeleton = ({
  bodyOnly = false,
  isDialog = false,
}: {
  bodyOnly?: boolean;
  isDialog?: boolean;
}) => {
  if (bodyOnly) {
    return (
      <Box className="h-full min-h-0">
        <MainDetailsSkeleton isDialog={isDialog} />
      </Box>
    );
  }

  return (
    <Box className="h-full min-h-0">
      <Box
        className={cn("md:hidden", {
          "dark:bg-surface": isDialog,
        })}
      >
        <MainDetailsSkeleton isDialog={isDialog} />
      </Box>
      <Box className="hidden h-full min-h-0 md:flex">
        <Box
          className={cn("min-w-0 flex-1", {
            "dark:bg-surface": isDialog,
          })}
        >
          <MainDetailsSkeleton isDialog={isDialog} />
        </Box>
        <Box className="border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px]">
          <OptionsSkeleton />
        </Box>
      </Box>
    </Box>
  );
};
