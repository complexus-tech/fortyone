"use client";
import { Box, BreadCrumbs, Flex } from "ui";
import { CalendarIcon, StoryIcon, UserIcon } from "icons";
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
  showCalendar,
}: {
  layout: MyWorkLayout;
  setLayout: (value: MyWorkLayout) => void;
  showCalendar: boolean;
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
  const isCalendar = layout === "calendar";
  let tabLabel: string = tab;
  if (isCalendar) {
    tabLabel = "Calendar";
  } else if (tab === "all") {
    tabLabel = `All ${getTermDisplay("storyTerm", { variant: "plural" })}`;
  }

  useHotkeys("v+l", () => {
    setLayout("list");
  });

  useHotkeys("v+k", () => {
    setLayout("kanban");
  });

  useHotkeys(
    "v+c",
    () => {
      setLayout("calendar");
    },
    { enabled: showCalendar },
  );

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
                icon: isCalendar ? (
                  <CalendarIcon strokeWidth={2} />
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
        <MyWorkLayoutSwitcher
          layout={layout}
          setLayout={setLayout}
          showCalendar={showCalendar}
        />
        {!isCalendar ? (
          <>
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
          </>
        ) : null}
        <span className="text-text-secondary hidden md:inline">|</span>
        <Box className="hidden md:block">
          <NewStoryButton data-header-new-story-button />
        </Box>
      </Flex>
    </HeaderContainer>
  );
};
