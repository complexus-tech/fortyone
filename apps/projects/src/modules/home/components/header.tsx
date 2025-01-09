"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { HomeIcon, PlusIcon } from "icons";
import { useState } from "react";
import { NewStoryButton, NewObjectiveDialog } from "@/components/ui";
import { HeaderContainer } from "@/components/shared";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <HeaderContainer className="justify-between border-0">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Home",
            icon: <HomeIcon className="h-[1.35rem] w-auto" />,
          },
        ]}
      />
      <Flex gap={3}>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-[1.15rem] w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
          variant="naked"
        >
          New Objective
        </Button>
        <NewStoryButton />
      </Flex>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
