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
  "flex items-center justify-center border border-primary bg-primary transition duration-200 ease-linear",
  {
    variants: {
      size: {
        sm: "px-2 h-[40px]",
        md: "px-3 h-[44px]",
        lg: "px-5 py-[0.56rem]",
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
    },
  }
);

export interface ButtonProps
  extends Omit<TouchableOpacityProps, "disabled">,
    VariantProps<typeof buttonVariants> {
  href?: string;
  loadingText?: string;
}

export const Button = ({
  rounded,
  size,
  loading,
  loadingText,
  href,
  className,
  disabled,
  children,
  ...rest
}: ButtonProps) => {
  const isDisabled = disabled || loading;

  const classes = cn(
    buttonVariants({
      size,
      disabled: isDisabled,
      loading,
      rounded,
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
        <>
          <ActivityIndicator size="small" color="white" />
          {loadingText && (
            <Text color="white" fontWeight="medium" className="ml-2">
              {loadingText}
            </Text>
          )}
        </>
      ) : (
        <Text color="white">{children}</Text>
      )}
    </TouchableOpacity>
  );
};
