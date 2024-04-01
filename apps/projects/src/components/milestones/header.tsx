"use client";
import { BreadCrumbs, Button } from "ui";
import { PlusIcon, MilestonesIcon } from "icons";
import { HeaderContainer } from "../shared/header-container";

export const ActiveMilestonesHeader = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Active Milestones",
            icon: <MilestonesIcon className="h-[1.15rem] w-auto" />,
          },
        ]}
      />
      <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="sm">
        New sprint
      </Button>
    </HeaderContainer>
  );
};
