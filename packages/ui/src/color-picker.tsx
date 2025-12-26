"use client";

import { useState } from "react";
import { Box } from "./box";
import { Button } from "./button";
import { Popover } from "./popover";
import { cn, colors } from "lib";

type ColorPickerProps = {
  value?: string;
  onChange?: (color: string) => void;
  onClick?: () => void;
  className?: string;
};

export const ColorPicker = ({
  value = "#F8F9FA",
  onChange,
  onClick,
  className,
}: ColorPickerProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <Popover.Trigger asChild>
        <Button
          type="button"
          className={className}
          asIcon
          color="tertiary"
          onClick={onClick}
          rounded="full"
          size="sm"
        >
          <span
            className="size-3.5 cursor-pointer rounded-full"
            style={{ backgroundColor: value }}
          />
        </Button>
      </Popover.Trigger>
      <Popover.Content className="p-2.5 rounded-2xl">
        <Box className="grid grid-cols-6 gap-1.5">
          {colors.map((color) => (
            <Box
              tabIndex={0}
              role="button"
              aria-label="Select color"
              key={color}
              className={cn(
                "size-8 cursor-pointer rounded-full transition-transform focus:outline-none hover:ring-2 dark:hover:ring-offset-dark-100",
                {
                  "ring-2 ring-offset-2 dark:ring-offset-dark-100":
                    color === value,
                }
              )}
              onClick={() => {
                onChange?.(color);
                setIsOpen(false);
              }}
              style={{ backgroundColor: color }}
            />
          ))}
        </Box>
      </Popover.Content>
    </Popover>
  );
};
