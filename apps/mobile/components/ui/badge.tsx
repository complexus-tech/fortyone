import React from "react";
import { View, ViewProps } from "react-native";
import { VariantProps, cva } from "cva";
import { cn } from "@/lib/utils/classnames";

const badgeVariants = cva(
  "flex-row items-center justify-center border gap-1.5 font-medium",
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded-sm",
        md: "rounded-lg",
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
          "bg-gray-50 border-gray-200/70 dark:bg-dark-200 dark:border-dark-100",
        secondary: "bg-secondary border-secondary",
        invert: "bg-black border-black",
      },
      size: {
        sm: "px-1.5 text-sm h-7 gap-1",
        md: "px-2 h-8 text-sm",
        lg: "px-3.5 h-10",
      },
    },
    compoundVariants: [
      {
        variant: "outline",
        color: "tertiary",
        className: "bg-white",
      },
      {
        variant: "outline",
        color: "primary",
        className: "bg-transparent",
      },
      {
        variant: "outline",
        color: "secondary",
        className: "bg-transparent",
      },
      {
        variant: "outline",
        color: "danger",
        className: "bg-transparent",
      },
      {
        variant: "outline",
        color: "info",
        className: "bg-transparent",
      },
      {
        variant: "outline",
        color: "warning",
        className: "bg-transparent",
      },
    ],
    defaultVariants: {
      size: "md",
      variant: "solid",
      color: "primary",
      rounded: "xl",
    },
  }
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
        className
      )}
      {...props}
    >
      {children}
    </View>
  );
};
