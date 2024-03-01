"use client";
import { BreadCrumbs, Flex } from "ui";
import { IssueIcon } from "icons";
import { HeaderContainer } from "@/components/layout";
import { IssuesFiltersButton, SideDetailsSwitch } from "@/components/ui";

export const Header = ({
  isExpanded,
  setIsExpanded,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
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
