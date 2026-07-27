"use client";

import { useState } from "react";
import { BreadCrumbs, Flex, Button, Box, Text } from "ui";
import { ObjectiveIcon, PlusIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLocalStorage, useTerminology, useUserRole } from "@/hooks";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { ListObjectives } from "@/modules/objectives/components/list-objectives";
import { NewObjectiveDialog } from "@/components/ui";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import type { RoadmapLayoutType } from "./types";

export const WorkspaceObjectivesPage = () => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const [layout, setLayout] = useLocalStorage<RoadmapLayoutType>(
    "objectivesLayout",
    "list",
  );
  const { data: objectives = [], isPending } = useObjectives();
  const [isOpen, setIsOpen] = useState(false);

  const renderContent = () => {
    if (isPending) {
      return <BoardSkeleton className="h-full" layout={layout} />;
    }

    if (objectives.length === 0) {
      return (
        <Box className="flex h-full items-center justify-center">
          <Box className="flex flex-col items-center">
            <ObjectiveIcon className="h-12 w-auto" strokeWidth={1.3} />
            <Text className="mt-8 mb-6" fontSize="3xl">
              Set your first{" "}
              {getTermDisplay("objectiveTerm", { capitalize: true })}
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              Define what the workspace wants to achieve, then connect
              measurable key results and the work that moves them forward.
            </Text>
            <Flex gap={2}>
              <Button
                color="tertiary"
                disabled={userRole === "guest"}
                leftIcon={<PlusIcon className="h-[1.1rem]" />}
                onClick={() => {
                  if (userRole !== "guest") {
                    setIsOpen(true);
                  }
                }}
                size="md"
              >
                Set your first {getTermDisplay("objectiveTerm")}
              </Button>
            </Flex>
          </Box>
        </Box>
      );
    }

    switch (layout) {
      case "gantt":
        return <RoadmapGanttBoard className="h-full" objectives={objectives} />;
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
                name: getTermDisplay("objectiveTerm", {
                  variant: "plural",
                  capitalize: true,
                }),
                icon: (
                  <ObjectiveIcon
                    className="h-[1.1rem] w-auto"
                    strokeWidth={2}
                  />
                ),
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={1}>
          <RoadmapLayoutSwitcher
            className="hidden md:flex"
            layout={layout}
            setLayout={setLayout}
          />
          <Button
            color="invert"
            disabled={userRole === "guest"}
            leftIcon={
              <PlusIcon className="h-[1.1rem] text-current dark:text-current" />
            }
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

export const RoadmapPage = WorkspaceObjectivesPage;
