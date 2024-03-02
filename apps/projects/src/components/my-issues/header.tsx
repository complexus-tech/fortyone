"use client";
import { BreadCrumbs, Flex } from "ui";
import { IssueIcon } from "icons";
import { HeaderContainer } from "@/components/layout";
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
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "My issues",
            icon: <IssueIcon className="h-5 w-auto" strokeWidth={2} />,
          },
          { name: "Assigned" },
        ]}
      />
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
