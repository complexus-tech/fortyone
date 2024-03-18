"use client";
import { Avatar, BreadCrumbs, Button, Flex } from "ui";
import { useState } from "react";
import {
  PlusIcon,
  ObjectiveIcon,
  SearchIcon,
  PreferencesIcon,
  ArrowDownIcon,
} from "icons";
import { HeaderContainer } from "@/components/shared";
import { StoryStatusIcon, NewProjectDialog } from "@/components/ui";

export const ProjectsHeader = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "All projects",
            icon: <ObjectiveIcon className="h-[1.1rem] w-auto" />,
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
          className="pl-1"
          color="tertiary"
          leftIcon={<Avatar color="naked" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          Lead
        </Button>
        <Button
          color="tertiary"
          leftIcon={<StoryStatusIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          Status
        </Button>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Preferences</span>
        </Button>
        <Button
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
        >
          New project
        </Button>
      </Flex>
      <NewProjectDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
