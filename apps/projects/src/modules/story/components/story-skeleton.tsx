"use client";
import { Box } from "ui";
import { MainDetailsSkeleton } from "./main-details-skeleton";
import { OptionsSkeleton } from "./options-skeleton";

export const StorySkeleton = ({
  isNotifications,
}: {
  isNotifications?: boolean;
}) => {
  return (
    <Box>
      <Box className="md:hidden">
        <MainDetailsSkeleton />
      </Box>
      <Box className="hidden md:flex">
        <Box className="min-w-0 flex-1">
            <MainDetailsSkeleton />
        </Box>
        <Box className="border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px]">
            <OptionsSkeleton />
        </Box>
      </Box>
    </Box>
  );
};
