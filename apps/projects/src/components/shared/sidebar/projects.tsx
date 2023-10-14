"use client";
import { cn } from "lib";
import { useState } from "react";
import { HiViewGrid } from "react-icons/hi";
import { TbCaretDownFilled, TbPlus } from "react-icons/tb";
import { Box, Button, Flex } from "ui";
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
        size="sm"
        variant="naked"
      >
        <span className="flex items-center gap-1">
          <HiViewGrid className="relative -top-[0.1px] h-[1.1rem] w-auto text-gray-300/80 dark:text-gray" />
          Projects
          <TbCaretDownFilled
            className={cn(
              "-rotate-90 text-gray-300/60 transition-transform dark:text-gray",
              {
                "rotate-0": isOpen,
              },
            )}
          />
        </span>

        <TbPlus className="hidden h-5 w-auto justify-self-end text-gray-300/60 group-hover:inline dark:dark:text-gray-200" />
      </Button>
      <Flex
        className={cn(
          "h-0 max-h-[55vh] overflow-y-auto transition-all duration-300",
          {
            "mt-1 h-max": isOpen,
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
