"use client";
import { BreadCrumbs, Flex } from "ui";
import { UserIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesViewOptionsButton,
  LayoutSwitcher,
  SideDetailsSwitch,
} from "@/components/ui";
import { useProfile } from "./provider";

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
  const { viewOptions, setViewOptions } = useProfile();
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Profile",
              icon: <UserIcon className="h-4 w-auto" />,
            },
            {
              name: "Joseph Mukorivo",
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <StoriesViewOptionsButton
          groupByOptions={["Status", "Priority"]}
          layout={layout}
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
