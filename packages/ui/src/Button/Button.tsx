import type { VariantProps } from "cva";
import { cva } from "cva";
import { LoadingIcon } from "icons";
import { cn } from "lib";
import Link from "next/link";
import { ReactNode, forwardRef, type ButtonHTMLAttributes } from "react";

export const buttonVariants = cva(
  "flex text-dark dark:text-gray-200 w-max items-center border gap-2 transition duration-200 ease-linear focus:outline-0",
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
        naked:
          "bg-opacity-10 dark:bg-opacity-10 hover:bg-opacity-20 dark:hover:bg-opacity-20",
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded",
        md: "rounded-lg",
        lg: "rounded-xl",
        xl: "rounded-3xl",
        full: "rounded-full",
      },
      color: {
        primary:
          "border-primary bg-primary ring-primary-200 enabled:hover:bg-primary-200 enabled:hover:border-primary-200 focus:bg-primary-200 focus:border-primary-200",
        danger:
          "text-danger border-danger bg-danger ring-danger hover:bg-danger/10",
        info: "text-info border-info bg-info ring-info enabled:hover:bg-info-300 enabled:hover:border-info-300 focus:bg-info-300 focus:border-info-300",
        warning:
          "text-warning border-warning bg-warning ring-warning enabled:hover:bg-warning-300 enabled:hover:border-warning-300 focus:bg-warning-300 focus:border-warning-300",
        tertiary:
          "dark:border-dark-50/80 dark:bg-dark-100/50 bg-gray-50 border-gray-100 focus:bg-gray-50 dark:focus:bg-dark-200 hover:bg-gray-100/50 active:bg-gray-50 dark:hover:bg-dark-200/80",
        secondary:
          "text-secondary border-secondary bg-secondary ring-secondary",
      },
      size: {
        xs: "px-1.5 h-[1.85rem] text-[0.95rem] gap-[2px]",
        sm: "px-2 h-[2.1rem] gap-1",
        md: "px-3 h-[2.4rem] md:h-[2.5rem]",
        lg: "px-5 md:px-7 py-2.5 md:py-3",
      },
      disabled: {
        true: "opacity-40 cursor-not-allowed",
      },
      active: {
        true: null,
      },
      loading: {
        true: "opacity-80 cursor-progress",
      },
      fullWidth: {
        true: "w-full",
      },
      align: {
        center: "justify-center",
        left: "justify-start",
        right: "justify-end",
        between: "justify-between",
      },
      asIcon: {
        true: "px-0 aspect-square justify-center",
      },
    },
    compoundVariants: [
      // Solid variant
      {
        variant: "solid",
        color: ["primary", "secondary", "warning", "danger", "info"],
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
        className:
          "enabled:bg-opacity-0 hover:enabled:bg-opacity-10 focus:enabled:bg-opacity-10",
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
        className: "bg-transparent dark:bg-transparent border-none",
      },
      {
        variant: "naked",
        color: ["secondary"],
        className: "dark:text-white",
      },
      {
        variant: "naked",
        color: "primary",
        className: "text-primary dark:text-primary",
      },
      {
        disabled: true,
        variant: ["outline", "naked"],
        className: "bg-opacity-100 text-white",
      },
      {
        disabled: true,
        variant: ["outline", "naked"],
        color: ["primary", "tertiary"],
        className: "bg-opacity-100 text-gray cursor-not-allowed",
      },
      {
        active: true,
        color: "tertiary",
        className: "bg-gray-100/80 bg-opacity-100 dark:bg-dark-50",
      },
    ],
    defaultVariants: {
      size: "md",
      variant: "solid",
      color: "primary",
      rounded: "md",
      align: "left",
    },
  }
);

export interface ButtonProps
  extends Omit<
      ButtonHTMLAttributes<HTMLButtonElement>,
      "color" | "disabled" | "active"
    >,
    VariantProps<typeof buttonVariants> {
  href?: string;
  loadingText?: string;
  target?: "_blank" | "_self" | "_parent" | "_top";
  rightIcon?: ReactNode;
  leftIcon?: ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (props, ref) => {
    const {
      variant,
      color,
      rounded,
      asIcon,
      size,
      target = "_self",
      loading,
      fullWidth,
      align,
      loadingText,
      href,
      active,
      rightIcon,
      leftIcon,
      className,
      disabled,
      children,
      ...rest
    } = props;

    const classes = cn(
      buttonVariants({
        variant,
        color,
        asIcon,
        size,
        disabled,
        loading,
        active,
        rounded,
        fullWidth,
        align,
      }),
      className
    );

    return (
      <>
        {href ? (
          <Link className={classes} href={href} target={target}>
            {leftIcon}
            {children}
            {rightIcon}
          </Link>
        ) : (
          <button
            className={classes}
            disabled={disabled as boolean}
            ref={ref}
            {...rest}
          >
            {loading ? (
              <>
                <LoadingIcon className="animate-spin" />
                {loadingText || "Loading..."}
              </>
            ) : (
              <>
                {leftIcon}
                {children}
                {rightIcon}
              </>
            )}
          </button>
        )}
      </>
    );
  }
);

Button.displayName = "Button";
