import { VariantProps, cva } from "cva";
import { FC, HTMLAttributes } from "react";

import { cn } from "lib";

const badge = cva(
  "flex w-max items-center justify-center font-medium rounded-[0.6rem] border gap-1",
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded-[0.4rem]",
        md: "rounded-[0.55rem]",
        lg: "rounded-xl",
        xl: "rounded-2xl",
        full: "rounded-full",
      },
      color: {
        primary: "text-white bg-primary border-primary",
        success: "text-white bg-success border-success",
        danger: "text-white bg-danger border-danger",
        info: "text-white bg-info border-info",
        warning: "text-white bg-warning border-warning",
        tertiary: "text-text-primary bg-surface border-border",
        secondary: "text-white bg-secondary border-secondary",
        invert: "bg-background-inverse text-foreground-inverse",
      },
      size: {
        xs: "h-[1.15rem] min-w-4 text-[70%] py-1 px-0.5",
        sm: "h-5 min-w-5 text-[80%] py-2 px-1",
        md: "h-6 text-[0.9rem] leading-6 p-2",
        lg: "h-8 px-3.5 text-[0.95rem]",
      },
    },
    compoundVariants: [
      {
        variant: "outline",
        color: "primary",
        className: "bg-transparent text-primary dark:text-primary",
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
