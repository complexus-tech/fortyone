import React from "react";
import { View, ViewProps } from "react-native";
import { VariantProps, cva } from "cva";
import { cn } from "@/lib/utils/classnames";

const badgeVariants = cva(
  "flex-row items-center justify-center gap-[4px] border",
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded-sm",
        md: "rounded-[8px]",
        lg: "rounded-xl",
        xl: "rounded-2xl",
        full: "rounded-full",
      },
      color: {
        primary: "bg-primary border-primary",
        success: "bg-success border-success",
        danger: "bg-danger border-danger",
        info: "bg-info border-info",
        warning: "bg-warning border-warning",
        tertiary:
          "border-border bg-surface-muted dark:border-border-dark dark:bg-surface-muted-dark",
        secondary:
          "bg-secondary border-secondary dark:bg-secondary-dark dark:border-secondary-dark",
        invert:
          "bg-background-inverse border-background-inverse dark:bg-background-inverse-dark dark:border-background-inverse-dark",
      },
      size: {
        sm: "min-h-[24px] px-[6px] py-[2px]",
        md: "min-h-[28px] px-[8px] py-[3px]",
        lg: "min-h-[36px] px-[12px] py-[5px]",
      },
    },
    compoundVariants: [
      {
        variant: "outline",
        className: "bg-transparent dark:bg-transparent",
      },
    ],
    defaultVariants: {
      size: "md",
      variant: "solid",
      color: "tertiary",
      rounded: "md",
    },
  },
);

export interface BadgeProps
  extends ViewProps,
    VariantProps<typeof badgeVariants> {
  children?: React.ReactNode;
}

export const Badge = ({
  className,
  variant,
  size,
  color,
  rounded,
  children,
  ...props
}: BadgeProps) => {
  return (
    <View
      className={cn(
        badgeVariants({ variant, color, size, rounded }),
        className,
      )}
      {...props}
    >
      {children}
    </View>
  );
};
