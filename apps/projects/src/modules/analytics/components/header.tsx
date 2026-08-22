"use client";
import { Flex, BreadCrumbs } from "ui";
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
            },
          ]}
        />
      </Flex>
      <NewStoryButton className="md:hidden" />
    </HeaderContainer>
  );
};
