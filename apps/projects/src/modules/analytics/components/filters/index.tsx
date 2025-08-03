"use client";
import { Flex } from "ui";
// import { DateRangeFilter } from "./date-range-filter";
import { TeamsFilter } from "./teams-filter";
import { SprintsFilter } from "./sprints-filter";
import { ObjectivesFilter } from "./objectives-filter";

export const Filters = () => {
  return (
    <Flex align="end" gap={4}>
      {/* <DateRangeFilter /> */}
      <TeamsFilter />
      <SprintsFilter />
      <ObjectivesFilter />
    </Flex>
  );
};

// Export individual components for potential standalone use
// export { DateRangeFilter } from "./date-range-filter";
export { TeamsFilter } from "./teams-filter";
export { SprintsFilter } from "./sprints-filter";
export { ObjectivesFilter } from "./objectives-filter";
export { FilterButton } from "./filter-button";
export * from "./types";
