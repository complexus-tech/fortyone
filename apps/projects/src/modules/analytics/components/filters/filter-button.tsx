"use client";
import React, { useState } from "react";
import { Box, Button, Flex, Popover, Text } from "ui";
import { ArrowDown2Icon } from "icons";
import type { FilterButtonProps } from "./types";

export const FilterButton = ({
  label,
  icon,
  text,
  popover,
}: FilterButtonProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Box>
      <Text className="mb-1 text-[0.95rem]" color="muted">
        {label}
      </Text>
      <Popover onOpenChange={setIsOpen} open={isOpen}>
        <Popover.Trigger asChild>
          <Button
            className="min-w-28 justify-between rounded-[0.7rem] md:h-[2.3rem]"
            color="tertiary"
            rightIcon={<ArrowDown2Icon className="h-4" strokeWidth={3} />}
            variant="outline"
          >
            <Flex align="center" gap={2}>
              {icon}
              <span>{text}</span>
            </Flex>
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          className="w-[23rem] bg-opacity-80 pb-2.5 dark:bg-opacity-80"
        >
          {popover}
        </Popover.Content>
      </Popover>
    </Box>
  );
};
