"use client";

import { Box, Flex, BreadCrumbs, Tabs } from "ui";
import { AnalyticsIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { Filters } from "./filters";

const REPORT_TABS = [
  { value: "overview", label: "Overview" },
  { value: "workload", label: "Workload" },
  { value: "flow", label: "Flow" },
  { value: "planning", label: "Planning" },
  { value: "engagement", label: "Engagement" },
  { value: "custom-fields", label: "Custom fields" },
];

export const Header = ({ showFilters = true }: { showFilters?: boolean }) => (
  <HeaderContainer className="justify-between gap-4">
    <Flex align="center" className="shrink-0" gap={2}>
      <MobileMenuButton />
      <BreadCrumbs
        breadCrumbs={[{ name: "Analytics", icon: <AnalyticsIcon /> }]}
      />
    </Flex>
    <Flex align="center" className="min-w-0" gap={3}>
      <Box className="min-w-0">
        <Tabs.List
          aria-label="Report sections"
          className="hide-scrollbar mx-0 max-w-[calc(100vw-15rem)] flex-nowrap overflow-x-auto md:mx-0 md:max-w-[40vw] xl:max-w-none"
        >
          {REPORT_TABS.map(({ value, label }) => (
            <Tabs.Tab
              className="shrink-0 whitespace-nowrap"
              key={value}
              value={value}
            >
              {label}
            </Tabs.Tab>
          ))}
        </Tabs.List>
      </Box>
      {showFilters ? (
        <Flex align="center" className="shrink-0 gap-3">
          <span aria-hidden className="bg-border h-5 w-px" />
          <Filters compact />
        </Flex>
      ) : null}
    </Flex>
  </HeaderContainer>
);
