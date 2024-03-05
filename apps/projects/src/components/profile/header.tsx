"use client";
import { Avatar, Flex, Text } from "ui";
import { HeaderContainer } from "@/components/shared";
import type { IssuesLayout } from "@/components/ui";
import {
  IssuesFiltersButton,
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
  layout: IssuesLayout;
  setLayout: (value: IssuesLayout) => void;
}) => {
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <Avatar
          name="Joseph Mukorivo"
          size="sm"
          src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
        />
        <Text>User Profile</Text>
      </Flex>
      <Flex align="center" gap={2}>
        <LayoutSwitcher layout={layout} setLayout={setLayout} />
        <IssuesFiltersButton />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
