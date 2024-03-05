"use client";
import { BreadCrumbs, Flex } from "ui";
import { UserIcon } from "icons";
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
