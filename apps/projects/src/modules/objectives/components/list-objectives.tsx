import { Box, Button, Flex, Text } from "ui";
import { ObjectiveIcon, PlusIcon } from "icons";
import { useState } from "react";
import { BodyContainer } from "@/components/shared/body";
import { NewObjectiveDialog } from "@/components/ui";
import type { Objective } from "../types";
import { TableHeader } from "./heading";
import { ObjectiveCard } from "./card";

export const ListObjectives = ({
  objectives,
  isInTeam,
}: {
  objectives: Objective[];
  isInTeam?: boolean;
}) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <BodyContainer className="h-[calc(100vh-3.7rem)]">
      <TableHeader isInTeam={isInTeam} />
      {objectives.length === 0 ? (
        <Box className="flex h-[70vh] items-center justify-center">
          <Box className="flex flex-col items-center">
            <ObjectiveIcon className="h-12 w-auto" strokeWidth={1.3} />
            <Text className="mb-6 mt-8" fontSize="3xl">
              No objectives found
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              Oops! This team doesn&apos;t have any objectives yet. Create a new
              objective to get started.
            </Text>
            <Flex gap={2}>
              <Button
                color="tertiary"
                leftIcon={<PlusIcon className="h-[1.1rem]" />}
                onClick={() => {
                  setIsOpen(true);
                }}
                size="md"
              >
                Create new objective
              </Button>
            </Flex>
          </Box>
        </Box>
      ) : (
        objectives.map((objective) => (
          <ObjectiveCard
            key={objective.id}
            {...objective}
            isInTeam={isInTeam}
          />
        ))
      )}
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </BodyContainer>
  );
};
