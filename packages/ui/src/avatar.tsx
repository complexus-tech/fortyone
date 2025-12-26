"use client";
import { VariantProps, cva } from "cva";
import { cn } from "lib";
import { FC, HTMLAttributes } from "react";
import { BlurImage } from "./image";
import { AssigneeIcon } from "icons";

const avatar = cva(
  "flex justify-center items-center aspect-square overflow-hidden text-center font-medium shrink-0 text-dark dark:text-white",
  {
    variants: {
      rounded: {
        full: "rounded-full",
        none: "rounded-none",
        sm: "rounded-sm",
        md: "rounded-lg",
        lg: "rounded-[0.7rem]",
      },
      color: {
        primary: "text-white bg-primary",
        secondary: "text-white bg-secondary",
        tertiary: "dark:bg-dark-50 bg-gray-100",
        naked: "bg-transparent",
      },
      size: {
        xs: "h-5 text-[0.6rem] leading-6",
        sm: "h-7 text-[0.8rem] leading-7",
        md: "h-8 text-sm leading-8",
        lg: "h-12 text-lg leading-12",
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
  extends Omit<HTMLAttributes<HTMLDivElement>, "color">,
    VariantProps<typeof avatar> {
  src?: string | null;
  name?: string;
}

const getInitials = (name: string) => {
  if (!name) {
    return "U";
  }

  const names = name.split(" ");

  // If single word with 2 or more characters, return first two characters
  if (names.length === 1) {
    return names[0].slice(0, 2).toUpperCase();
  }

  let initials = "";
  initials += names[0][0]; // First initial of the first name

  if (names.length > 1) {
    initials += names[names.length - 1][0]; // First initial of the last name
  }

  return initials.toUpperCase();
};

export const Avatar: FC<AvatarProps> = (props) => {
  const { className, src, name, color, size, rounded, ...rest } = props;
  const classes = cn(avatar({ rounded, color, size }), className);
  const asIcon = !src && !name;

  return (
    <div
      className={cn(classes, {
        "bg-transparent dark:bg-transparent": asIcon,
      })}
      {...rest}
    >
      {src && (
        <BlurImage
          src={src}
          priority
          alt={name}
          className={cn("w-full h-full aspect-square", {
            "rounded-full": rounded === "full",
            "rounded-sm": rounded === "sm",
            "rounded-md": rounded === "md",
            "rounded-[0.6rem]": rounded === "lg",
          })}
          imageClassName="object-top object-cover"
        />
      )}
      {!src && name && <span title={name}>{getInitials(name.trim())}</span>}
      {asIcon && (
        <AssigneeIcon
          className={cn("h-5 w-auto opacity-70", {
            "h-6": size === "sm",
            "h-5": size === "xs",
          })}
          strokeWidth={2}
        />
      )}
    </div>
  );
};

Avatar.displayName = "Avatar";
