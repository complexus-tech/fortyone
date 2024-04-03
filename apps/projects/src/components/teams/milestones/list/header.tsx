"use client";
import { Avatar, BreadCrumbs, Button, Flex } from "ui";
import {
  PlusIcon,
  SearchIcon,
  PreferencesIcon,
  ArrowDownIcon,
  MilestonesIcon,
} from "icons";
import { HeaderContainer } from "@/components/shared";

export const MilestonesHeader = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Engineering",
            icon: "🚀",
          },
          {
            name: "Milestones",
            icon: <MilestonesIcon className="h-[1.1rem] w-auto" />,
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
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Preferences</span>
        </Button>
        <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="sm">
          New sprint
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
