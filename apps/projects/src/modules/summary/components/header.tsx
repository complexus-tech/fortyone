"use client";
import { BreadCrumbs, Flex } from "ui";
import { DashboardIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import { NewStoryButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
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
