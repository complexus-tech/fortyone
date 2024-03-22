"use client";
import { cn } from "lib";
import { useState } from "react";
import { Box, Button, Flex } from "ui";
import { ArrowDownIcon, PlusIcon } from "icons";
import { NewObjectiveDialog } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Objective } from "./objective";

export const Objectives = () => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isOpen, setIsOpen] = useLocalStorage<boolean>(
    "objectives-dropdown",
    true,
  );

  const objectives = [
    {
      id: 1,
      icon: "🚀",
      name: "Website design",
    },
    {
      id: 2,
      icon: "🇦🇫",
      name: "Data migration",
    },
    {
      id: 3,
      icon: "🏀",
      name: "CRM development",
    },
  ];

  return (
    <Box className="mt-4">
      <Flex
        align="center"
        className="group h-[2.6rem] select-none rounded-lg pl-2.5 pr-1 outline-none transition hover:bg-gray-50/70 focus:bg-gray-50/70 hover:dark:bg-dark-50/20 focus:dark:bg-dark-50/20"
        justify="between"
        onClick={() => {
          setIsOpen(!isOpen);
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            setIsOpen(!isOpen);
          }
        }}
        role="button"
        tabIndex={0}
      >
        <span className="flex items-center gap-2 font-medium">
          My Objectives
          <ArrowDownIcon
            className={cn(
              "relative top-[0.2px] h-4 w-auto -rotate-90 text-gray-300/60 transition-transform dark:text-gray",
              {
                "rotate-0": isOpen,
              },
            )}
            strokeWidth={3.5}
          />
        </span>
        <Button
          className="hidden group-hover:inline"
          color="tertiary"
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          onClick={() => {
            setIsDialogOpen(true);
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <span className="sr-only">Add new objective</span>
        </Button>
      </Flex>
      <Flex
        className={cn(
          "h-0 max-h-[54vh] overflow-y-auto pb-2 transition-all duration-300",
          {
            "h-max": isOpen,
          },
        )}
        direction="column"
      >
        {objectives.map(({ id, icon, name }) => (
          <Objective icon={icon} id={id} key={id} name={name} />
        ))}
      </Flex>

      <NewObjectiveDialog isOpen={isDialogOpen} setIsOpen={setIsDialogOpen} />
    </Box>
  );
};
