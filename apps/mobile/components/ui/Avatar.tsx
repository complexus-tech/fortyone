import React from "react";
import { View, ViewProps } from "react-native";
import { Image } from "expo-image";
import { VariantProps, cva } from "cva";
import { Text } from "./Text";
import { cn } from "@/lib/utils";
import { AssigneeIcon } from "../icons/assignee";
import { getAvatarInitials } from "./avatar-initials";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";

const avatarVariants = cva(
  "flex justify-center items-center aspect-square overflow-hidden text-center font-semibold shrink-0",
  {
    variants: {
      rounded: {
        full: "rounded-full",
        none: "rounded-none",
        sm: "rounded-sm",
        md: "rounded-md",
        lg: "rounded-lg",
        xl: "rounded-xl",
        "2xl": "rounded-2xl",
      },
      color: {
        primary: "bg-primary",
        secondary: "bg-secondary dark:bg-secondary-dark",
        tertiary: "bg-accent dark:bg-accent-dark",
        naked: "bg-transparent",
      },
      size: {
        xs: "size-5 text-xs",
        sm: "size-7 text-xs",
        md: "size-9 text-sm",
        lg: "size-11 text-base",
        xl: "size-14 text-lg",
      },
    },
    defaultVariants: {
      size: "md",
      rounded: "full",
      color: "tertiary",
    },
  },
);

export interface AvatarProps
  extends ViewProps,
    VariantProps<typeof avatarVariants> {
  src?: string | null;
  name?: string;
  textClassName?: string;
  fallbackIconSize?: number;
}

export const Avatar = ({
  className,
  src,
  name,
  color,
  size,
  rounded,
  textClassName,
  fallbackIconSize,
  ...props
}: AvatarProps) => {
  const { resolvedTheme } = useTheme();
  const iconColor = themeColors[resolvedTheme].icon;
  const classes = avatarVariants({ rounded, color, size });
  const asIcon = !src && !name;

  return (
    <View
      className={cn(
        classes,
        {
          "bg-transparent dark:bg-transparent": asIcon,
        },
        className,
      )}
      {...props}
    >
      {src && (
        <Image
          source={src}
          className={cn("aspect-square", {
            "rounded-full": rounded === "full",
            "rounded-sm": rounded === "sm",
            "rounded-md": rounded === "md",
            "rounded-lg": rounded === "lg",
          })}
          contentFit="cover"
          contentPosition="top center"
          style={{
            width: "100%",
            height: "100%",
          }}
        />
      )}
      {!src && name && (
        <Text
          color={
            color === "primary"
              ? "primaryForeground"
              : color === "secondary"
                ? "secondaryForeground"
                : undefined
          }
          className={cn(
            {
              "text-[0.6rem]": size === "xs",
              "text-sm": size === "sm",
              "text-md": size === "md",
              "text-lg": size === "lg",
            },
            textClassName,
          )}
          fontWeight="bold"
        >
          {getAvatarInitials(name)}
        </Text>
      )}
      {asIcon && (
        <AssigneeIcon
          size={
            fallbackIconSize ??
            (size === "xs"
              ? 18
              : size === "sm"
                ? 24
                : size === "md"
                  ? 25
                  : size === "lg"
                    ? 30
                    : 25)
          }
          color={iconColor}
        />
      )}
    </View>
  );
};
