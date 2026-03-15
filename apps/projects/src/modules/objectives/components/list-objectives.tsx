import { Box, Button, Flex, Text } from "ui";
import { ObjectiveIcon, PlusIcon } from "icons";
import { useState } from "react";
import { cn } from "lib";
import { BodyContainer } from "@/components/shared/body";
import { NewObjectiveDialog } from "@/components/ui";
import { useTerminology, useUserRole } from "@/hooks";
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
  const { getTermDisplay } = useTerminology();

  return (
    <BodyContainer
      className={cn("h-[calc(100dvh-3.7rem)] overflow-x-auto", {
        "h-auto": isInSearch,
      })}
    >
      {objectives.length === 0 ? (
        <>
          {isInSearch ? null : (
            <Box className="flex h-full items-center justify-center">
              <Box className="flex flex-col items-center">
                <ObjectiveIcon className="h-12 w-auto" strokeWidth={1.6} />
                <Text className="mt-8 mb-6" fontSize="3xl">
                  No {getTermDisplay("objectiveTerm", { variant: "plural" })}{" "}
                  found
                </Text>
                <Text className="mb-6 max-w-md text-center" color="muted">
                  Oops! This team doesn&apos;t have any{" "}
                  {getTermDisplay("objectiveTerm", { variant: "plural" })} yet.
                  Create a new {getTermDisplay("objectiveTerm")} to get started.
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
                    Create new {getTermDisplay("objectiveTerm")}
                  </Button>
                </Flex>
              </Box>
            </Box>
          )}
        </>
      ) : (
        <Box
          className={cn("min-w-5xl", {
            "min-w-272": isInTeam,
            "min-w-6xl": !isInTeam,
          })}
        >
          {!isInSearch && <TableHeader isInTeam={isInTeam} />}
          {objectives.map((objective) => (
            <ObjectiveCard
              key={objective.id}
              {...objective}
              isInSearch={isInSearch}
              isInTeam={isInTeam}
            />
          ))}
        </Box>
      )}
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </BodyContainer>
  );
};
