"use client";
import { Box, BreadCrumbs, Flex } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import {
  StoriesViewOptionsButton,
  NewStoryButton,
  StoriesFilterButton,
} from "@/components/ui";
import { useTerminology } from "@/hooks";
import type { MyWorkLayout } from "../types";
import { MyWorkLayoutSwitcher } from "./my-work-layout-switcher";
import { useMyWork } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: MyWorkLayout;
  setLayout: (value: MyWorkLayout) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { viewOptions, setViewOptions, filters, resetFilters, setFilters } =
    useMyWork();
  const tabs = [
    "all",
    "today",
    "upcoming",
    "blocked",
    "assigned",
    "collaborating",
    "created",
  ] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  let tabLabel: string = tab;
  if (tab === "all") {
    tabLabel = `All ${getTermDisplay("storyTerm", { variant: "plural" })}`;
  }

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });

  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <Box className="md:hidden">
          <BreadCrumbs
            breadCrumbs={[
              {
                name: `My ${getTermDisplay("storyTerm", { variant: "plural" })}`,
              },
            ]}
          />
        </Box>
        <Box className="hidden md:block">
          <BreadCrumbs
            breadCrumbs={[
              {
                name: `My ${getTermDisplay("storyTerm", { variant: "plural" })}`,
              },
              {
                name: tabLabel,
                className: "capitalize",
              },
            ]}
          />
        </Box>
      </Flex>
      <Flex align="center" gap={2}>
        <MyWorkLayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesFilterButton
          filters={filters}
          resetFilters={resetFilters}
          setFilters={setFilters}
        />
        <StoriesViewOptionsButton
          groupByOptions={["status", "priority", "assignee"]}
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <span className="text-text-secondary hidden md:inline">|</span>
        <Box className="hidden md:block">
          <NewStoryButton data-header-new-story-button />
        </Box>
      </Flex>
    </HeaderContainer>
  );
};
