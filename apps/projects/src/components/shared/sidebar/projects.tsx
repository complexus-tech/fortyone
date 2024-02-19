"use client";
import { cn } from "lib";
import { useState } from "react";
import { Box, Button, Flex } from "ui";
import { ChevronDown, Plus } from "lucide-react";
import { ProjectsIcon } from "@/components/icons";
import { Project } from "./project";

export const Projects = () => {
  const [isOpen, setIsOpen] = useState(true);
  return (
    <Box className="mt-8">
      <Button
        align="between"
        className="group"
        color="tertiary"
        fullWidth
        onClick={() => {
          setIsOpen(!isOpen);
        }}
        variant="naked"
      >
        <span className="flex items-center gap-2 font-medium">
          <ProjectsIcon className="relative h-5 w-auto text-gray-300/80 dark:text-gray" />
          Projects
          <ChevronDown
            className={cn(
              "relative top-[0.2px] h-4 w-auto -rotate-90 text-gray-300/60 transition-transform dark:text-gray",
              {
                "rotate-0": isOpen,
              },
            )}
            strokeWidth={3.5}
          />
        </span>
        <Plus className="hidden h-5 w-auto justify-self-end text-gray-300/60 group-hover:inline dark:dark:text-gray-200" />
      </Button>
      <Flex
        className={cn(
          "h-0 max-h-[55vh] overflow-y-auto transition-all duration-300",
          {
            "h-max": isOpen,
          },
        )}
        direction="column"
      >
        <Project icon="🚀" name="Website design" />
        <Project icon="🇦🇫" name="Data migration" />
        <Project icon="🏀" name="CRM development" />
      </Flex>
    </Box>
  );
};
