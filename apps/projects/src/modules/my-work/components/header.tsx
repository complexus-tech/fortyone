"use client";
import { Box, BreadCrumbs, Flex } from "ui";
import { HealthIcon, StoryIcon, UserIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesViewOptionsButton,
  LayoutSwitcher,
  NewStoryButton,
  StoriesFilterButton,
} from "@/components/ui";
import { useTerminology, useUserRole } from "@/hooks";
import { useMyWork } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { viewOptions, setViewOptions, filters, resetFilters, setFilters } =
    useMyWork();
  const { userRole } = useUserRole();
  const isAdmin = userRole === "admin";
  const adminTabs = ["pulse", "all", "assigned", "created"] as const;
  const storyTabs = ["all", "assigned", "created"] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(isAdmin ? adminTabs : storyTabs).withDefault(
      isAdmin ? "pulse" : "all",
    ),
  );
  const isPulseTab = tab === "pulse";
  const tabLabel = (() => {
    if (tab === "pulse") return "Pulse";
    if (tab === "all") {
      return `All ${getTermDisplay("storyTerm", { variant: "plural" })}`;
    }
    return tab;
  })();

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
                icon: <UserIcon />,
              },
            ]}
          />
        </Box>
        <Box className="hidden md:block">
          <BreadCrumbs
            breadCrumbs={[
              {
                name: `My ${getTermDisplay("storyTerm", { variant: "plural" })}`,
                icon: <UserIcon />,
              },
              {
                name: tabLabel,
                icon: isPulseTab ? (
                  <HealthIcon strokeWidth={2} />
                ) : (
                  <StoryIcon strokeWidth={2} />
                ),
                className: "capitalize",
              },
            ]}
          />
        </Box>
      </Flex>
      <Flex align="center" gap={2}>
        {!isPulseTab ? (
          <>
            <LayoutSwitcher layout={layout} setLayout={setLayout} />
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
          </>
        ) : null}
        <Box className="hidden md:block">
          <NewStoryButton data-header-new-story-button />
        </Box>
      </Flex>
    </HeaderContainer>
  );
};
