"use client";
import { useState } from "react";
import { cn } from "lib";
import { Box, Button, Flex, Popover, Text } from "ui";
import type { FilterButtonProps } from "./types";

export const FilterButton = ({
  label,
  icon,
  text,
  popover,
  showLabel = true,
  compact = false,
}: FilterButtonProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Box>
      {showLabel ? (
        <Text className="mb-1" color="muted">
          {label}
        </Text>
      ) : null}
      <Popover onOpenChange={setIsOpen} open={isOpen}>
        <Popover.Trigger asChild>
          <Button
            aria-label={`${label}: ${text}`}
            className={cn(
              "justify-between rounded-xl md:h-[2.3rem]",
              compact ? "min-w-0 px-2 xl:px-3" : "min-w-28",
            )}
            color="tertiary"
            variant="outline"
          >
            <Flex align="center" gap={2}>
              {icon}
              <span className={compact ? "hidden xl:inline" : undefined}>
                {text}
              </span>
            </Flex>
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="end"
          className="mr-0 max-h-[var(--radix-popover-content-available-height)] w-92 max-w-[calc(100vw-2rem)] overflow-y-auto pb-2.5"
        >
          {popover}
        </Popover.Content>
      </Popover>
    </Box>
  );
};
