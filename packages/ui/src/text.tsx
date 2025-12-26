import { VariantProps, cva } from "cva";
import { HTMLAttributes, JSX, createElement } from "react";

import { cn } from "lib";

const text = cva("text-dark dark:text-gray-200", {
  variants: {
    align: {
      left: "text-left",
      center: "text-center",
      right: "text-right",
    },
    color: {
      primary: "text-primary dark:text-primary",
      muted: "text-gray dark:text-gray-300/80",
      danger: "text-danger dark:text-danger",
      gradient:
        "bg-gradient-to-r from-primary dark:via-gray-200 dark:to-info to-secondary bg-clip-text text-transparent dark:text-transparent",
      gradientDark:
        "bg-gradient-to-r from-dark dark:from-white dark:to-gray/70 to-dark/30 bg-clip-text text-transparent dark:text-transparent",
      black: "text-gray-300",
      white: "text-gray-200",
      warning: "text-warning",
      info: "text-info",
      secondary: "text-secondary",
    },
    fontSize: {
      xs: "text-xs",
      sm: "text-sm",
      md: "text-base",
      lg: "text-lg",
      xl: "text-xl",
      "2xl": "text-2xl",
      "3xl": "text-3xl",
      "4xl": "text-4xl",
      inherit: "text-inherit",
    },
    fontWeight: {
      light: "font-light",
      normal: "font-normal",
      medium: "font-medium",
      semibold: "font-semibold",
      bold: "font-bold",
    },
    transform: {
      uppercase: "uppercase",
      lowercase: "lowercase",
      capitalize: "capitalize",
      none: "normal-case",
    },
    fontStyle: {
      italic: "italic",
      normal: "not-italic",
    },
    decoration: {
      underline: "underline",
      lineThrough: "line-through",
      none: "no-underline",
    },
    textOverflow: {
      ellipsis: "text-ellipsis",
      clip: "text-clip",
      truncate: "truncate",
    },
  },
});

export interface TextProps
  extends Omit<
      HTMLAttributes<
        HTMLHeadingElement | HTMLSpanElement | HTMLParagraphElement
      >,
      "color"
    >,
    VariantProps<typeof text> {
  as?: keyof JSX.IntrinsicElements;
  html?: string;
}

export const Text = ({
  as: Tag = "p",
  children,
  className,
  html,
  align,
  color,
  fontSize,
  fontWeight,
  transform,
  fontStyle,
  decoration,
  textOverflow,
  ...rest
}: TextProps) => {
  const htmlProps = html
    ? {
        dangerouslySetInnerHTML: { __html: html },
      }
    : {};

  const classes = cn(
    text({
      align,
      color,
      fontSize,
      fontWeight,
      transform,
      fontStyle,
      decoration,
      textOverflow,
    }),
    className
  );

  return createElement(Tag, {
    className: classes,
    ...rest,
    ...htmlProps,
    children,
  });
};
