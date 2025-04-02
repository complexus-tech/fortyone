import { Box, Button, Flex, Text } from "ui";
import { ObjectiveIcon, PlusIcon } from "icons";
import { useState } from "react";
import { cn } from "lib";
import { BodyContainer } from "@/components/shared/body";
import { NewObjectiveDialog } from "@/components/ui";
import { useUserRole } from "@/hooks";
import type { Objective } from "../types";
import { TableHeader } from "./heading";
import { ObjectiveCard } from "./card";

export const ListObjectives = ({
  objectives,
  isInTeam,
  isInSearch,
}: {
  objectives: Objective[];
  isInTeam?: boolean;
  isInSearch?: boolean;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();

  return (
    <BodyContainer
      className={cn("h-[calc(100vh-3.7rem)]", {
        "h-auto": isInSearch,
      })}
    >
      {!isInSearch && <TableHeader isInTeam={isInTeam} />}
      {objectives.length === 0 ? (
        <>
          {isInSearch ? (
            <Box className="flex h-full items-center justify-center">
              <Box className="flex flex-col items-center">
                <ObjectiveIcon className="h-12 w-auto" strokeWidth={1.3} />
                <Text className="mb-6 mt-8" fontSize="3xl">
                  No results found
                </Text>
                <Text className="mb-6 max-w-md text-center" color="muted">
                  Oops! There are no results for your search. Try a different
                  search query.
                </Text>
              </Box>
            </Box>
          ) : (
            <Box className="flex h-full items-center justify-center">
              <Box className="flex flex-col items-center">
                <ObjectiveIcon className="h-12 w-auto" strokeWidth={1.3} />
                <Text className="mb-6 mt-8" fontSize="3xl">
                  No objectives found
                </Text>
                <Text className="mb-6 max-w-md text-center" color="muted">
                  Oops! This team doesn&apos;t have any objectives yet. Create a
                  new objective to get started.
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
                    Create new objective
                  </Button>
                </Flex>
              </Box>
            </Box>
          )}
        </>
      ) : (
        objectives.map((objective) => (
          <ObjectiveCard
            key={objective.id}
            {...objective}
            isInSearch={isInSearch}
            isInTeam={isInTeam}
          />
        ))
      )}
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </BodyContainer>
  );
};
