"use client";

import { Box, Flex, Text } from "ui";
import { ObjectiveCard } from "@/components/ui/objective/card";
import type { Objective } from "@/modules/objectives/types";
import { BodyContainer } from "@/components/shared/body";
import { ObjectivesHeader } from "./components/header";

const TableHeader = () => (
  <Box className="sticky top-0 z-[1] border-b-[0.5px] border-gray-100/60 bg-gray-50/60 py-3 backdrop-blur dark:border-dark-100 dark:bg-dark-300/90">
    <Flex align="center" className="px-6" justify="between">
      <Box className="w-[300px] shrink-0">
        <Text color="muted" fontWeight="medium">
          Name
        </Text>
      </Box>
      <Flex gap={4}>
        <Box className="w-[120px] shrink-0">
          <Text color="muted" fontWeight="medium">
            Team
          </Text>
        </Box>
        <Box className="w-[140px] shrink-0">
          <Text color="muted" fontWeight="medium">
            Owner
          </Text>
        </Box>
        <Box className="w-[120px] shrink-0">
          <Text color="muted" fontWeight="medium">
            KR Progress
          </Text>
        </Box>
        <Box className="w-[120px] shrink-0">
          <Text color="muted" fontWeight="medium">
            Target
          </Text>
        </Box>
        <Box className="w-[120px] shrink-0">
          <Text color="muted" fontWeight="medium">
            Created On
          </Text>
        </Box>
        <Box className="w-[100px] shrink-0">
          <Text color="muted" fontWeight="medium">
            Health
          </Text>
        </Box>
      </Flex>
    </Flex>
  </Box>
);

export const ObjectivesList = ({ objectives }: { objectives: Objective[] }) => {
  return (
    <>
      <ObjectivesHeader />
      <BodyContainer className="h-[calc(100vh-3.7rem)]">
        <TableHeader />
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
      </BodyContainer>
    </>
  );
};
