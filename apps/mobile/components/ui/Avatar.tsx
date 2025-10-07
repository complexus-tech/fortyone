import React from "react";
import { View, ViewProps } from "react-native";
import { Image } from "expo-image";
import { VariantProps, cva } from "cva";
import { Text } from "./text";
import { cn } from "@/lib/utils";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

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
        xl: "rounded-xl",
        "2xl": "rounded-2xl",
      },
      color: {
        primary: "text-white bg-primary",
        secondary: "text-white bg-secondary",
        tertiary: "bg-gray-100 dark:bg-dark-200",
        naked: "bg-transparent",
      },
      size: {
        xs: "size-5 text-xs",
        sm: "size-7 text-xs",
        md: "size-9 text-sm",
        lg: "size-10 text-base",
        xl: "size-14 text-lg",
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
  textClassName?: string;
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
  textClassName,
  ...props
}: AvatarProps) => {
  const classes = avatarVariants({ rounded, color, size });
  const asIcon = !src && !name;

  return (
    <View
      className={cn(
        classes,
        {
          "bg-transparent dark:bg-transparent": asIcon,
        },
        className
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
            color === "primary" || color === "secondary" ? "white" : "black"
          }
          className={cn(
            {
              "text-[0.6rem]": size === "xs",
              "text-sm": size === "sm",
              "text-md": size === "md",
              "text-lg": size === "lg",
            },
            textClassName
          )}
          fontWeight="semibold"
        >
          {getInitials(name)}
        </Text>
      )}
      {asIcon && (
        <SymbolView
          name="person.crop.circle.dashed"
          size={
            size === "xs"
              ? 18
              : size === "sm"
                ? 24
                : size === "md"
                  ? 25
                  : size === "lg"
                    ? 30
                    : 25
          }
          tintColor={colors.gray.DEFAULT}
        />
      )}
    </View>
  );
};
