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
  "flex text-dark w-max items-center border gap-2 transition duration-200 ease-linear",
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
        naked: "bg-opacity-10 hover:bg-opacity-20",
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded",
        md: "rounded-[0.7rem]",
        lg: "rounded-[0.85rem]",
        xl: "rounded-3xl",
        full: "rounded-full",
      },
      color: {
        primary: "border-primary bg-primary",
        danger: "text-danger border-danger bg-danger",
        info: "text-info border-info bg-info",
        warning: "text-warning border-warning bg-warning",
        tertiary: "bg-gray-50/80 border-gray-100",
        secondary: "text-secondary border-secondary bg-secondary",
        white: "text-black border-white bg-white",
        invert: "text-white bg-dark border-dark",
        black: "text-white border-black bg-black",
      },
      size: {
        sm: "px-2 h-[40px] gap-1",
        md: "px-3 h-[44px]",
        lg: "px-5 py-[0.56rem]",
      },
      disabled: {
        true: "opacity-40",
        false: "",
      },
      loading: {
        true: "opacity-80",
        false: "",
      },
      fullWidth: {
        true: "w-full",
        false: "",
      },
      align: {
        center: "justify-center",
        left: "justify-start",
        right: "justify-end",
        between: "justify-between",
      },
      asIcon: {
        true: "px-0 aspect-square justify-center",
        false: "",
      },
    },
    compoundVariants: [
      // Solid variant
      {
        variant: "solid",
        color: ["primary", "secondary", "danger", "info"],
        className: "text-white",
      },
      {
        variant: "solid",
        color: "warning",
        className: "text-dark",
      },
      // Outline variant
      {
        variant: "outline",
        color: ["primary", "secondary", "warning", "danger", "info"],
        className: "bg-opacity-0 hover:bg-opacity-10",
      },
      {
        variant: "outline",
        color: "primary",
        className: "text-primary bg-transparent",
      },
      {
        variant: "outline",
        color: "black",
        className: "text-black border-black bg-transparent",
      },
      {
        variant: "outline",
        color: "tertiary",
        className: "bg-white border-gray-100",
      },
      // Naked variant
      {
        variant: "naked",
        color: [
          "primary",
          "secondary",
          "warning",
          "danger",
          "info",
          "tertiary",
        ],
        className: "bg-transparent border-none",
      },
      {
        variant: "naked",
        color: "secondary",
        className: "text-white",
      },
      {
        variant: "naked",
        color: "primary",
        className: "text-primary bg-primary/5 border-primary/5",
      },
      {
        variant: "naked",
        color: "danger",
        className: "text-danger bg-danger/5",
      },
    ],
    defaultVariants: {
      size: "md",
      variant: "solid",
      color: "primary",
      rounded: "md",
      align: "center",
    },
  }
);

export interface ButtonProps
  extends Omit<TouchableOpacityProps, "disabled">,
    VariantProps<typeof buttonVariants> {
  href?: string;
  loadingText?: string;
  rightIcon?: React.ReactNode;
  leftIcon?: React.ReactNode;
}

export const Button = ({
  variant,
  color,
  rounded,
  asIcon,
  size,
  loading,
  fullWidth,
  align,
  loadingText,
  href,
  rightIcon,
  leftIcon,
  className,
  disabled,
  children,
  ...rest
}: ButtonProps) => {
  const isDisabled = disabled || loading;

  const classes = cn(
    buttonVariants({
      variant,
      color,
      asIcon,
      size,
      disabled: isDisabled,
      loading,
      rounded,
      fullWidth,
      align,
    }),
    className
  );

  const getTextColor = () => {
    if (variant === "solid") {
      if (color === "warning") return "black";
      if (color === "tertiary") return "gray";
      return "white";
    }
    if (variant === "outline" || variant === "naked") {
      return color || "primary";
    }
    return "black";
  };

  return (
    <TouchableOpacity
      className={classes}
      disabled={isDisabled as boolean}
      activeOpacity={0.7}
      {...rest}
    >
      {loading ? (
        <>
          <ActivityIndicator
            size="small"
            color={
              variant === "solid" && color !== "warning" && color !== "tertiary"
                ? "white"
                : "black"
            }
          />
          {loadingText && (
            <Text
              color={getTextColor() as any}
              fontSize="sm"
              fontWeight="medium"
              className="ml-2"
            >
              {loadingText}
            </Text>
          )}
        </>
      ) : (
        <>
          {leftIcon}
          {children && (
            <Text
              color={getTextColor() as any}
              fontSize="sm"
              fontWeight="medium"
              className={leftIcon ? "ml-2" : rightIcon ? "mr-2" : ""}
            >
              {children}
            </Text>
          )}
          {rightIcon}
        </>
      )}
    </TouchableOpacity>
  );
};
