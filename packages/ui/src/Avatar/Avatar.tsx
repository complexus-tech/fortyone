"use client";
import { VariantProps, cva } from "cva";
import { cn } from "lib";
import { User } from "lucide-react";
import { FC, HTMLAttributes, useState } from "react";

const avatar = cva(
  "inline-flex justify-center items-center aspect-square overflow-hidden text-center font-medium",
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
        gray: "text-black bg-gray",
      },
      size: {
        xs: "h-6 text-xs leading-6",
        sm: "h-7 text-sm leading-7",
        md: "h-8 text-base leading-8",
        lg: "h-12 text-lg leading-[3rem]",
      },
    },
    defaultVariants: {
      size: "md",
      rounded: "full",
      color: "primary",
    },
  }
);

export interface AvatarProps
  extends Omit<HTMLAttributes<HTMLDivElement>, "color">,
    VariantProps<typeof avatar> {
  src?: string;
  name?: string;
}

const getInitials = (name: string) => {
  if (!name) {
    return "U";
  }

  const names = name.split(" ");
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
  const [path, setPath] = useState(src);

  return (
    <div className={classes} {...rest}>
      {path && (
        <img
          src={path}
          alt={name}
          onError={(_) => {
            setPath("");
          }}
          className={cn("w-full h-auto object-cover", {
            "rounded-full": rounded === "full",
            "rounded-sm": rounded === "sm",
            "rounded-md": rounded === "md",
            "rounded-lg": rounded === "lg",
          })}
        />
      )}
      {!path && name && <span title={name}>{getInitials(name)}</span>}
      {!path && !name && (
        <User
          className={cn("h-5 w-auto", {
            "h-5": size === "sm",
            "h-auto": size === "xs",
          })}
        />
      )}
    </div>
  );
};

Avatar.displayName = "Avatar";
