"use client";
import { Flex, BreadCrumbs } from "ui";
import { AnalyticsIcon } from "icons";
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
              name: "Analytics",
              icon: <AnalyticsIcon />,
            },
          ]}
        />
      </Flex>
      <NewStoryButton />
    </HeaderContainer>
  );
};
