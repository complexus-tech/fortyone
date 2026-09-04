"use client";
import { Box, BreadCrumbs, Flex, Tabs } from "ui";
import { StoryIcon, UserIcon } from "icons";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { StoriesViewOptionsButton, StoriesFilterButton } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
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
  const {
    filters,
    resetFilters,
    setFilters,
    setTab,
    setViewOptions,
    tab,
    viewOptions,
    visibleTabs,
  } = useMyWork();
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

  const tabLabels = {
    all: `All ${getTermDisplay("storyTerm", { variant: "plural" })}`,
    assigned: "Assigned",
    blocked: "Blocked",
    collaborating: "Collaborating",
    created: "Created",
    today: "Today",
    upcoming: "Upcoming",
  } as const;

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
                icon: (
                  <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                ),
                className: "capitalize",
              },
            ]}
          />
        </Box>
      </Flex>
      <Flex align="center" className="min-w-0" gap={2}>
        <Box data-walkthrough-target={walkthroughTargets.myWorkTabs}>
          <Tabs
            onValueChange={(value) => {
              setTab(value as typeof tab);
            }}
            value={tab}
          >
            <Tabs.List className="hide-scrollbar mx-0 max-w-[48vw] flex-nowrap overflow-x-auto md:mx-0">
              {visibleTabs.map((visibleTab) => (
                <Tabs.Tab key={visibleTab} value={visibleTab}>
                  {tabLabels[visibleTab]}
                </Tabs.Tab>
              ))}
            </Tabs.List>
          </Tabs>
        </Box>
        <span aria-hidden="true" className="bg-border mx-1 h-5 w-px shrink-0" />
        <Box data-walkthrough-target={walkthroughTargets.myWorkViewControls}>
          <MyWorkLayoutSwitcher layout={layout} setLayout={setLayout} />
        </Box>
        <Box data-walkthrough-target={walkthroughTargets.myWorkFilters}>
          <StoriesFilterButton
            filters={filters}
            resetFilters={resetFilters}
            setFilters={setFilters}
          />
        </Box>
        <Box data-walkthrough-target={walkthroughTargets.myWorkDisplayOptions}>
          <StoriesViewOptionsButton
            groupByOptions={["status", "priority", "assignee"]}
            layout={layout}
            setViewOptions={setViewOptions}
            viewOptions={viewOptions}
          />
        </Box>
      </Flex>
    </HeaderContainer>
  );
};
