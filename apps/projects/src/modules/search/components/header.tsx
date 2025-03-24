"use client";
import { BreadCrumbs, Flex } from "ui";
import { SearchIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import { NewStoryButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Search",
              icon: <SearchIcon />,
            },
          ]}
        />
      </Flex>
      <NewStoryButton />
    </HeaderContainer>
  );
};
