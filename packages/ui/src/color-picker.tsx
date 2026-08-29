"use client";

import { useState, type ReactNode } from "react";
import { Box } from "./box";
import { Button, type ButtonProps } from "./button";
import { Popover } from "./popover";
import { cn, colors } from "lib";

type ColorPickerProps = {
  value?: string;
  onChange?: (color: string) => void;
  onClick?: () => void;
  className?: string;
  children?: ReactNode;
  disabled?: ButtonProps["disabled"];
  ariaLabel?: string;
  size?: ButtonProps["size"];
  style?: ButtonProps["style"];
};

export const ColorPicker = ({
  value = "#F8F9FA",
  onChange,
  onClick,
  className,
  children,
  disabled,
  ariaLabel = "Choose color",
  size = "sm",
  style,
}: ColorPickerProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <Popover.Trigger asChild>
        <Button
          aria-label={ariaLabel}
          asIcon
          className={className}
          color="tertiary"
          disabled={disabled}
          onClick={onClick}
          size={size}
          style={style}
          type="button"
        >
          {children ?? (
            <span
              className="size-3.5 cursor-pointer rounded-sm"
              style={{ backgroundColor: value }}
            />
          )}
        </Button>
      </Popover.Trigger>
      <Popover.Content className="rounded-lg p-2.5">
        <Box className="grid grid-cols-6 gap-1.5">
          {colors.map((color) => (
            <Box
              tabIndex={0}
              role="button"
              aria-label="Select color"
              key={color}
              className={cn(
                "ring-primary size-8 cursor-pointer rounded-md transition-transform hover:ring-2 focus:outline-none",
                {
                  "ring-2 ring-offset-background": color === value,
                },
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
