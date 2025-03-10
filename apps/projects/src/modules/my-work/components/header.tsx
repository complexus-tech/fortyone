"use client";
import { BreadCrumbs, Flex, Badge } from "ui";
import { StoryIcon, UserIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesViewOptionsButton,
  LayoutSwitcher,
  NewStoryButton,
} from "@/components/ui";
import { useMyStories } from "../hooks/my-stories";
import { useMyWork } from "./provider";

export const Header = ({
  layout,
  setLayout,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { data } = useMyStories();
  const { viewOptions, setViewOptions } = useMyWork();
  const tabs = ["all", "assigned", "created"] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "My Work",
              icon: <UserIcon />,
            },
            {
              name: tab === "all" ? "All stories" : tab,
              icon: <StoryIcon strokeWidth={2} />,
              className: "capitalize",
            },
          ]}
        />

        <Badge className="bg-opacity-50" color="tertiary" rounded="full">
          {data?.length} stories
        </Badge>
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesViewOptionsButton
          groupByOptions={["Status", "Priority", "Assignee", "None"]}
          layout={layout}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <NewStoryButton />
      </Flex>
    </HeaderContainer>
  );
};
