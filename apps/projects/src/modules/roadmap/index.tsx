"use client";

import { useState } from "react";
import { BreadCrumbs, Flex, Button, Box } from "ui";
import { RoadmapIcon, PlusIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLocalStorage, useTerminology, useUserRole } from "@/hooks";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import { ListObjectives } from "@/modules/objectives/components/list-objectives";
import { NewObjectiveDialog } from "@/components/ui";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import type { RoadmapLayoutType } from "./types";

export const RoadmapPage = () => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const [layout, setLayout] = useLocalStorage<RoadmapLayoutType>(
    "roadmapLayout",
    "gantt",
  );
  const { data: objectives = [] } = useObjectives();
  const [isOpen, setIsOpen] = useState(false);

  const renderContent = () => {
    switch (layout) {
      case "gantt":
        return (
          <RoadmapGanttBoard
            className="h-full"
            objectives={[
              ...objectives,
              ...objectives,
              ...objectives,
              ...objectives,
              ...objectives,
            ]}
          />
        );
      case "list":
        return <ListObjectives objectives={objectives} />;
      default:
        return null;
    }
  };

  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Roadmap",
                icon: (
                  <RoadmapIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                ),
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={1}>
          <RoadmapLayoutSwitcher layout={layout} setLayout={setLayout} />
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
            New {getTermDisplay("objectiveTerm", { capitalize: true })}
          </Button>
        </Flex>
      </HeaderContainer>

      <Box className="h-[calc(100dvh-4rem)]">{renderContent()}</Box>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
