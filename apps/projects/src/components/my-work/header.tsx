"use client";
import { BreadCrumbs, Flex } from "ui";
import { StoryIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesViewOptionsButton,
  LayoutSwitcher,
  SideDetailsSwitch,
} from "@/components/ui";

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
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "My Work",
            icon: <StoryIcon className="h-5 w-auto" strokeWidth={2} />,
          },
          { name: "Assigned" },
        ]}
      />
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesViewOptionsButton />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
