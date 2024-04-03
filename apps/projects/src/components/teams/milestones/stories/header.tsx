"use client";
import { BreadCrumbs, Flex } from "ui";
import { MilestonesIcon, StoryIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesViewOptionsButton,
  SideDetailsSwitch,
} from "@/components/ui";
import { useMilestoneStories } from "./provider";

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
  const { viewOptions, setViewOptions } = useMilestoneStories();
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Engineering",
            icon: "🚀",
          },
          {
            name: "Milestones",
            icon: <MilestonesIcon className="h-4 w-auto" />,
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
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
