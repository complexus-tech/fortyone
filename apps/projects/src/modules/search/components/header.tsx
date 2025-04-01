"use client";
import { Flex } from "ui";
import { SearchIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import { NewStoryButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <SearchIcon className="h-5" />
        <input
          className="w-full flex-1 bg-transparent placeholder:text-gray focus:outline-none"
          placeholder="Search"
          type="text"
        />
      </Flex>
      <NewStoryButton />
    </HeaderContainer>
  );
};
