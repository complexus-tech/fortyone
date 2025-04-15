"use client";
import { BreadCrumbs, Flex } from "ui";
import { DashboardIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { NewStoryButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Summary",
              icon: <DashboardIcon />,
            },
          ]}
        />
      </Flex>
      <NewStoryButton />
    </HeaderContainer>
  );
};
