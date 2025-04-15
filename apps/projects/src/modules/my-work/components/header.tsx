"use client";
import { Box, BreadCrumbs, Flex } from "ui";
import { StoryIcon, UserIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useHotkeys } from "react-hotkeys-hook";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesViewOptionsButton,
  LayoutSwitcher,
  NewStoryButton,
} from "@/components/ui";
import { useTerminology } from "@/hooks";
import { useMyWork } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { viewOptions, setViewOptions } = useMyWork();
  const tabs = ["all", "assigned", "created"] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );

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
                name:
                  tab === "all"
                    ? `All ${getTermDisplay("storyTerm", { variant: "plural" })}`
                    : tab,
                icon: <StoryIcon strokeWidth={2} />,
                className: "capitalize",
              },
            ]}
          />
        </Box>
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesViewOptionsButton
          groupByOptions={["Status", "Priority", "Assignee", "None"]}
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <span className="hidden text-gray-200 dark:text-dark-100 md:inline">
          |
        </span>
        <Box className="hidden md:block">
          <NewStoryButton />
        </Box>
      </Flex>
    </HeaderContainer>
  );
};
