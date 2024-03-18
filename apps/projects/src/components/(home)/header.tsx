"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { HomeIcon, PlusIcon } from "icons";
import { useState } from "react";
import { NewStoryButton, NewProjectDialog } from "@/components/ui";
import { HeaderContainer } from "@/components/shared";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Home",
            icon: <HomeIcon className="h-[1.35rem] w-auto" />,
          },
        ]}
      />
      <Flex gap={2}>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-[1.15rem] w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
        >
          New project
        </Button>
        <NewStoryButton />
      </Flex>
      <NewProjectDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
