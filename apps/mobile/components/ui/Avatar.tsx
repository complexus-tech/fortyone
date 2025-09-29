import React from "react";
import { View, Image, ViewProps } from "react-native";
import { VariantProps, cva } from "cva";
import { Text } from "./Text";
import { cn } from "@/lib/utils";

const avatarVariants = cva(
  "flex justify-center items-center aspect-square overflow-hidden text-center font-medium shrink-0",
  {
    variants: {
      rounded: {
        full: "rounded-full",
        none: "rounded-none",
        sm: "rounded-sm",
        md: "rounded-md",
        lg: "rounded-lg",
      },
      color: {
        primary: "text-white bg-primary",
        secondary: "text-white bg-secondary",
        tertiary: "bg-gray-100",
        naked: "bg-transparent",
      },
      size: {
        sm: "size-8 text-xs",
        md: "size-10 text-sm",
        lg: "size-12 text-base",
        xl: "size-16 text-lg",
      },
    },
    defaultVariants: {
      size: "md",
      rounded: "full",
      color: "tertiary",
    },
  }
);

export interface AvatarProps
  extends ViewProps,
    VariantProps<typeof avatarVariants> {
  src?: string | null;
  name?: string;
}

const getInitials = (name: string) => {
  if (!name) {
    return "U";
  }

  const names = name.split(" ");

  // If single word with 2 or more characters, return first two characters
  if (names.length === 1 && names[0].length >= 2) {
    return names[0].slice(0, 2).toUpperCase();
  }

  let initials = "";
  initials += names[0][0]; // First initial of the first name

  if (names.length > 1) {
    initials += names[names.length - 1][0]; // First initial of the last name
  }

  return initials.toUpperCase();
};

export const Avatar = ({
  className,
  src,
  name,
  color,
  size,
  rounded,
  ...props
}: AvatarProps) => {
  const classes = avatarVariants({ rounded, color, size });
  const asIcon = !src && !name;

  return (
    <View
      className={cn(
        classes,
        {
          "bg-transparent": asIcon,
        },
        className
      )}
      {...props}
    >
      {src && (
        <Image
          source={{ uri: src }}
          className={cn("w-full h-full aspect-square", {
            "rounded-full": rounded === "full",
            "rounded-sm": rounded === "sm",
            "rounded-md": rounded === "md",
            "rounded-lg": rounded === "lg",
          })}
          resizeMode="cover"
        />
      )}
      {!src && name && (
        <Text
          color={
            color === "primary" || color === "secondary" ? "white" : "black"
          }
          fontSize={
            size === "sm"
              ? "xs"
              : size === "md"
                ? "sm"
                : size === "lg"
                  ? "md"
                  : "lg"
          }
          fontWeight="semibold"
        >
          {getInitials(name)}
        </Text>
      )}
      {asIcon && (
        <Text color="muted" fontSize="sm">
          👤
        </Text>
      )}
    </View>
  );
};
