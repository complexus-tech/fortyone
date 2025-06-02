"use client";

import { useCallback } from "react";
import { BreadCrumbs, Flex, Text } from "ui";
import { RoadmapIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useLocalStorage } from "@/hooks";
import {
  RoadmapLayout,
  type RoadmapLayoutType,
} from "@/components/ui/roadmap-layout";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import { ListObjectives } from "@/modules/objectives/components/list-objectives";

export const RoadmapPage = () => {
  const [layout, setLayout] = useLocalStorage<RoadmapLayoutType>(
    "roadmapLayout",
    "gantt",
  );
  const { data: objectives = [] } = useObjectives();

  const handleLayoutChange = useCallback(
    (newLayout: RoadmapLayoutType) => {
      setLayout(newLayout);
    },
    [setLayout],
  );

  const renderContent = () => {
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
                name: "Roadmap",
                icon: (
                  <RoadmapIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                ),
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={2}>
          <Text className="text-gray-500 text-sm">
            {objectives.length} objective{objectives.length !== 1 ? "s" : ""}
          </Text>
        </Flex>
      </HeaderContainer>

      <RoadmapLayout
        className="h-[calc(100dvh-4rem)]"
        layout={layout}
        onLayoutChange={handleLayoutChange}
      >
        {renderContent()}
      </RoadmapLayout>
    </>
  );
};
