"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { useState } from "react";
import { PlusIcon, ObjectiveIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import { NewObjectiveDialog } from "@/components/ui";
import { useUserRole } from "@/hooks";

export const ObjectivesHeader = () => {
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Objectives",
              icon: <ObjectiveIcon className="h-[1.05rem]" strokeWidth={2} />,
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <Button
          color="tertiary"
          disabled={userRole === "guest"}
          leftIcon={<PlusIcon className="h-[1.1rem]" />}
          onClick={() => {
            if (userRole !== "guest") {
              setIsOpen(true);
            }
          }}
          size="sm"
        >
          New Objective
        </Button>
      </Flex>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
