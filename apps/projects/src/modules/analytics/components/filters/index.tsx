"use client";
import { Flex } from "ui";
import { DateRangeFilter } from "./date-range-filter";
import { TeamsFilter } from "./teams-filter";
import { ObjectivesFilter } from "./objectives-filter";

export const Filters = () => {
  return (
    <Flex align="end" gap={4}>
      <DateRangeFilter />
      <TeamsFilter />
      <ObjectivesFilter />
    </Flex>
  );
};
