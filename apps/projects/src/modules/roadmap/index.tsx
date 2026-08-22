"use client";

import { useState } from "react";
import { BreadCrumbs, Flex, Button, Box, Text } from "ui";
import { PlusIcon, RoadmapIcon, WarningIcon } from "icons";
import {
  HeaderContainer,
  MobileMenuButton,
  useAppCommandAction,
} from "@/components/shared";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLocalStorage, useTerminology, useUserRole } from "@/hooks";
import { NewObjectiveDialog } from "@/components/ui";
import { RoadmapLayoutSwitcher } from "@/components/ui/roadmap-layout-switcher";
import { RoadmapEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import type { ZoomLevel } from "@/components/ui/base-gantt";
import { isObjectiveForecastAtRisk } from "@/modules/objectives/components/objective-forecast-risk-utils";
import { ObjectiveViews } from "./components/objective-views";
import { ObjectiveViewOptionsButton } from "./components/objective-view-options-button";
import {
  DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  type ObjectiveViewOptions,
} from "./objective-board-utils";
import { getRoadmapLayoutLabel } from "./types";
import { useRoadmapLayout } from "./use-roadmap-layout";

export const RoadmapPage = () => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const { layout, setLayout } = useRoadmapLayout();
  const { data: objectives = [], isPending } = useObjectives();
  const [showForecastRisksOnly, setShowForecastRisksOnly] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [zoomLevel, setZoomLevel] = useLocalStorage<ZoomLevel>(
    "roadmapZoomLevel",
    "months",
  );
  const [viewOptions, setViewOptions] = useLocalStorage<ObjectiveViewOptions>(
    "objectivesViewOptions",
    DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  );

  useAppCommandAction({
    disabled: userRole === "guest",
    id: "roadmap:create-objective",
    label: `Create ${getTermDisplay("objectiveTerm")}`,
    onSelect: () => {
      setIsOpen(true);
    },
  });
  const forecastRiskObjectives = objectives.filter(isObjectiveForecastAtRisk);
  const isForecastRiskFilterActive =
    showForecastRisksOnly && forecastRiskObjectives.length > 0;
  const displayedObjectives = isForecastRiskFilterActive
    ? forecastRiskObjectives
    : objectives;
  const forecastRiskCountLabel = `${forecastRiskObjectives.length} ${
    forecastRiskObjectives.length === 1 ? "item needs" : "items need"
  } attention`;

  const emptyState = (
    <Box className="flex h-full items-center justify-center">
      <Box className="flex flex-col items-center">
        <RoadmapEmptyIllustration />
        <Text className="mt-8 mb-6" fontSize="3xl">
          Set your first {getTermDisplay("objectiveTerm", { capitalize: true })}
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Define what the workspace wants to achieve, then connect measurable
          key results and the work that moves them forward.
        </Text>
        <Flex gap={2}>
          <Button
            className="md:hidden"
            color="primary"
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
              {
                name: getRoadmapLayoutLabel(layout),
              },
            ]}
          />
          {forecastRiskObjectives.length > 0 ? (
            <Button
              aria-pressed={isForecastRiskFilterActive}
              className={
                isForecastRiskFilterActive
                  ? "text-primary-foreground dark:text-primary-foreground hidden gap-1.5 sm:flex"
                  : "text-primary dark:text-primary hidden gap-1.5 sm:flex"
              }
              color="primary"
              leftIcon={
                <WarningIcon
                  className={
                    isForecastRiskFilterActive
                      ? "text-primary-foreground dark:text-primary-foreground h-4 w-auto"
                      : "text-primary dark:text-primary h-4 w-auto"
                  }
                />
              }
              onClick={() => {
                setShowForecastRisksOnly((current) => !current);
              }}
              size="sm"
              type="button"
              variant={isForecastRiskFilterActive ? "solid" : "naked"}
            >
              {forecastRiskCountLabel}
            </Button>
          ) : null}
        </Flex>
        <Flex align="center" gap={2}>
          <RoadmapLayoutSwitcher
            className="hidden md:flex"
            layout={layout}
            setLayout={setLayout}
          />
          {layout !== "gantt" ? (
            <ObjectiveViewOptionsButton
              setViewOptions={setViewOptions}
              viewOptions={viewOptions}
            />
          ) : null}
          <Button
            className="md:hidden"
            color="primary"
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

      <Box className="h-[calc(100%-3.6rem)] min-w-0">
        <ObjectiveViews
          emptyState={emptyState}
          isPending={isPending}
          layout={layout}
          objectives={displayedObjectives}
          onCreateObjective={() => {
            if (userRole !== "guest") setIsOpen(true);
          }}
          onZoomLevelChange={setZoomLevel}
          setViewOptions={setViewOptions}
          viewOptions={viewOptions}
          zoomLevel={zoomLevel}
        />
      </Box>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
