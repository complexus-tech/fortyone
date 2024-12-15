"use client";
import { BreadCrumbs, Flex, Badge } from "ui";
import { StoryIcon, UserIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesViewOptionsButton,
  LayoutSwitcher,
  SideDetailsSwitch,
} from "@/components/ui";
import { useMyWork } from "./provider";
import { useMyStories } from "../hooks/my-stories";
import { parseAsStringLiteral, useQueryState } from "nuqs";

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
  const { data } = useMyStories();
  const { viewOptions, setViewOptions } = useMyWork();
  const tabs = ["assigned", "created", "subscribed"] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("assigned"),
  );
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "My Work",
              icon: <StoryIcon className="h-5 w-auto" strokeWidth={2} />,
            },
            {
              name: tab,
              icon: <UserIcon className="h-4 w-auto" />,
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
