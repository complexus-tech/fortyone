import React from "react";
import {
  TouchableOpacity,
  TouchableOpacityProps,
  ActivityIndicator,
} from "react-native";
import { VariantProps, cva } from "cva";
import { Text } from "./text";
import { cn } from "@/lib/utils";

export const buttonVariants = cva(
  "flex items-center justify-center transition duration-200 ease-linear",
  {
    variants: {
      size: {
        sm: "px-2 h-[40px]",
        md: "px-3 h-[44px]",
        lg: "px-5 h-14",
      },
      color: {
        primary: "bg-primary",
        invert: "bg-dark dark:bg-white",
        tertiary: "bg-gray-100/70 dark:bg-dark-100/80",
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded",
        md: "rounded-[0.7rem]",
        lg: "rounded-[0.85rem]",
        xl: "rounded-3xl",
        full: "rounded-full",
      },
      disabled: {
        true: "opacity-40",
        false: "",
      },
      loading: {
        true: "opacity-80",
        false: "",
      },
    },
    defaultVariants: {
      size: "md",
      rounded: "md",
      color: "primary",
    },
  }
);

export interface ButtonProps
  extends Omit<TouchableOpacityProps, "disabled">,
    VariantProps<typeof buttonVariants> {
  href?: string;
  isDestructive?: boolean;
}

export const Button = ({
  rounded,
  size,
  loading,
  href,
  className,
  disabled,
  children,
  color,
  isDestructive,
  ...rest
}: ButtonProps) => {
  const isDisabled = disabled || loading;

  const classes = cn(
    buttonVariants({
      size,
      disabled: isDisabled,
      loading,
      rounded,
      color,
    }),
    className
  );

  return (
    <TouchableOpacity
      className={classes}
      disabled={isDisabled as boolean}
      activeOpacity={0.7}
      {...rest}
    >
      {loading ? (
        <ActivityIndicator
          size="small"
          className={cn({
            "text-white": color === "primary",
            "text-white dark:text-dark": color === "invert",
          })}
        />
      ) : (
        <Text
          className={cn({
            "text-white": color === "primary",
            "text-white dark:text-dark": color === "invert",
            "text-danger dark:text-danger": isDestructive,
          })}
        >
          {children}
        </Text>
      )}
    </TouchableOpacity>
  );
};
