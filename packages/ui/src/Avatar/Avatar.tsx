"use client";
import { VariantProps, cva } from "cva";
import { cn } from "lib";
import { FC, HTMLAttributes, useState } from "react";
import { BlurImage } from "../Image/Image";

const avatar = cva(
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
        gray: "text-black bg-gray",
        naked: "bg-transparent",
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
        <BlurImage
          src={path}
          priority
          alt={name}
          className={cn("w-full h-full aspect-square", {
            "rounded-full": rounded === "full",
            "rounded-sm": rounded === "sm",
            "rounded-md": rounded === "md",
            "rounded-lg": rounded === "lg",
          })}
        />
      )}
      {!path && name && <span title={name}>{getInitials(name)}</span>}
      {!path && !name && (
        <svg
          className={cn("h-5 w-auto", {
            "h-5": size === "sm",
            "h-auto": size === "xs",
          })}
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="2"
          />
          <path
            d="M7.5 17C9.8317 14.5578 14.1432 14.4428 16.5 17M14.4951 9.5C14.4951 10.8807 13.3742 12 11.9915 12C10.6089 12 9.48797 10.8807 9.48797 9.5C9.48797 8.11929 10.6089 7 11.9915 7C13.3742 7 14.4951 8.11929 14.4951 9.5Z"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
          />
        </svg>
      )}
    </div>
  );
};

Avatar.displayName = "Avatar";
