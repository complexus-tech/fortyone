import { VariantProps, cva } from "cva";
import { FC, HTMLAttributes } from "react";

import { cn } from "lib";

const badge = cva(
  "flex w-max items-center justify-center font-medium rounded-lg border-[0.5px] gap-1",
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded",
        md: "rounded-[0.4rem]",
        lg: "rounded-xl",
        full: "rounded-full",
      },
      color: {
        primary: "text-white bg-primary border-primary",
        success: "text-white bg-success border-success",
        danger: "text-white bg-danger border-danger",
        info: "text-white bg-info border-info",
        warning: "text-white bg-warning border-warning",
        tertiary:
          "text-gray bg-gray-50 border-gray-100/80 dark:bg-dark-300 dark:border-dark-50 dark:text-gray-200",
        secondary: "text-white bg-secondary border-secondary",
      },
      size: {
        sm: "h-5 min-w-[1.25rem] text-[80%] py-2 px-1",
        md: "h-6 text-[0.9rem] leading-6 p-2",
        lg: "h-8 px-3.5 text-[0.95rem]",
      },
    },
    compoundVariants: [
      {
        variant: "outline",
        color: "tertiary",
        className: "bg-gray-50 text-black dark:text-white dark:bg-dark-200/60",
      },
      {
        variant: "outline",
        color: "primary",
        className: "bg-transparent text-primary-200 dark:text-primary",
      },
      {
        variant: "outline",
        color: "secondary",
        className: "bg-transparent text-secondary",
      },
      {
        variant: "outline",
        color: "danger",
        className: "bg-transparent text-danger",
      },
      {
        variant: "outline",
        color: "info",
        className: "bg-transparent text-info",
      },
      {
        variant: "outline",
        color: "warning",
        className: "bg-transparent text-warning",
      },
    ],
    defaultVariants: {
      size: "md",
      variant: "solid",
      color: "primary",
      rounded: "md",
    },
  }
);

export interface BadgeProps
  extends Omit<HTMLAttributes<HTMLSpanElement>, "color">,
    VariantProps<typeof badge> {}

export const Badge: FC<BadgeProps> = (props) => {
  const { className, variant, size, color, rounded, children, ...rest } = props;
  const classes = cn(badge({ variant, color, size, rounded }), className);

  return (
    <span className={classes} {...rest}>
      {children}
    </span>
  );
};

Badge.displayName = "Badge";
