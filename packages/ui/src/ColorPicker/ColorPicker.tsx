"use client";

import { Box } from "../Box/Box";
import { Flex } from "../Flex/Flex";
import { Popover } from "../Popover/Popover";
import { cn } from "lib";
import { forwardRef } from "react";

const colors = [
  // Light colors row
  "#F8F9FA",
  "#FFE066",
  "#FF6B6B",
  "#C0392B",
  // Pink/Purple row
  "#E056FD",
  "#686DE0",
  "#E67E22",
  "#A8E6CF",
  // Green row
  "#6BCB77",
  "#4ECDC4",
  "#4A90E2",
  "#95A5A6",
  // Dark row
  "#30336B",
  "#B4A6AB",
  "#DFE6E9",
  "#636E72",
  // Extra dark
  "#2D3436",
];

type ColorPickerProps = {
  value?: string;
  onChange?: (color: string) => void;
  className?: string;
};

export const ColorPicker = forwardRef<HTMLDivElement, ColorPickerProps>(
  ({ value = "#F8F9FA", onChange, className }, ref) => {
    return (
      <Popover>
        <Popover.Trigger>
          <div
            className={cn(
              "size-8 cursor-pointer rounded border border-gray-100 dark:border-dark-100",
              className
            )}
            ref={ref}
            style={{ backgroundColor: value }}
          />
        </Popover.Trigger>
        <Popover.Content align="start" className="w-[232px] p-2">
          <Flex className="flex-wrap gap-1">
            {colors.map((color) => (
              <Box
                key={color}
                className={cn(
                  "size-12 cursor-pointer rounded transition-transform hover:scale-110",
                  {
                    "ring-2 ring-primary ring-offset-2 dark:ring-offset-dark-100":
                      color === value,
                  }
                )}
                onClick={() => onChange?.(color)}
                style={{ backgroundColor: color }}
              />
            ))}
          </Flex>
        </Popover.Content>
      </Popover>
    );
  }
);

ColorPicker.displayName = "ColorPicker";
