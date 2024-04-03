"use client";
import { BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  NewStoryButton,
  SideDetailsSwitch,
  StoriesViewOptionsButton,
} from "@/components/ui";
import { useTeamStories } from "./provider";

export const Header = ({
  isExpanded,
  setIsExpanded,
  layout,
  setLayout,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
}) => {
  const { viewOptions, setViewOptions } = useTeamStories();

  return (
    <HeaderContainer className="justify-between">
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
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
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
