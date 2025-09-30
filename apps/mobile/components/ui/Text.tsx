import { Text as RNText, TextProps as RNTextProps } from "react-native";
import { VariantProps, cva } from "cva";
import { cn } from "@/lib/utils";

const textVariants = cva("text-dark", {
  variants: {
    align: {
      left: "text-left",
      center: "text-center",
      right: "text-right",
    },
    color: {
      primary: "text-primary",
      secondary: "text-secondary",
      muted: "text-gray",
      danger: "text-danger",
      warning: "text-warning",
      info: "text-info",
      success: "text-success",
      black: "text-black",
      white: "text-white",
    },
    fontSize: {
      xs: "text-sm",
      sm: "text-base",
      md: "text-lg",
      lg: "text-xl",
      xl: "text-2xl",
      "2xl": "text-3xl",
      "3xl": "text-4xl",
      "4xl": "text-5xl",
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
  },
  defaultVariants: {
    align: "left",
    color: "black",
    fontSize: "md",
    fontWeight: "medium",
    transform: "none",
    fontStyle: "normal",
    decoration: "none",
  },
});

export interface TextProps
  extends RNTextProps,
    VariantProps<typeof textVariants> {
  className?: string;
}

export const Text = ({
  children,
  className,
  align,
  color,
  fontSize,
  fontWeight,
  transform,
  fontStyle,
  decoration,
  ...props
}: TextProps) => {
  const classes = textVariants({
    align,
    color,
    fontSize,
    fontWeight,
    transform,
    fontStyle,
    decoration,
  });

  return (
    <RNText className={cn(classes, className)} {...props}>
      {children}
    </RNText>
  );
};
