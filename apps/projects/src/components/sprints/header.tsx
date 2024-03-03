"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { PlusIcon, PreferencesIcon, SearchIcon, SprintsIcon } from "icons";
import { HeaderContainer } from "../shared/header-container";

export const ActiveSprintsHeader = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Active sprints",
            icon: <SprintsIcon className="h-[1.15rem] w-auto" />,
          },
        ]}
      />
      <Flex gap={3}>
        <Button
          align="center"
          className="px-[0.6rem]"
          color="tertiary"
          leftIcon={<SearchIcon className="h-[1.1rem] w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="sm">
          New sprint
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
