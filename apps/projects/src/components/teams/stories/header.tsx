"use client";
import { Badge, BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  NewStoryButton,
  SideDetailsSwitch,
  StoriesFilterButton,
  StoriesViewOptionsButton,
} from "@/components/ui";
import { useTeamStories } from "./provider";

export const Header = ({
  isExpanded,
  allStories,
  setIsExpanded,
  layout,
  setLayout,
}: {
  isExpanded: boolean | null;
  allStories: number;
  setIsExpanded: (isExpanded: boolean) => void;
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { viewOptions, setViewOptions } = useTeamStories();

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Engineering",
              icon: "🚀",
            },
            {
              name: "Stories",
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
        <Badge className="bg-opacity-50" color="tertiary" rounded="full">
          {allStories} stories
        </Badge>
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesFilterButton
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <StoriesViewOptionsButton
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
        />
        <NewStoryButton />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
