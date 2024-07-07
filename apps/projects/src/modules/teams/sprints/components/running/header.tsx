"use client";
import { BreadCrumbs, Button } from "ui";
import { PlusIcon, SprintsIcon } from "icons";
import { HeaderContainer } from "../../../../../components/shared/header-container";

export const ActiveSprintsHeader = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Running Sprints",
            icon: <SprintsIcon className="h-5 w-auto" />,
          },
        ]}
      />
      <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="sm">
        New Sprint
      </Button>
    </HeaderContainer>
  );
};
