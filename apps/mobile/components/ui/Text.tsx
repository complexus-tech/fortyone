import { Text as RNText, TextProps as RNTextProps } from "react-native";
import { VariantProps, cva } from "cva";
import { cn } from "@/lib/utils";

const textVariants = cva("text-dark dark:text-white", {
  variants: {
    align: {
      left: "text-left",
      center: "text-center",
      right: "text-right",
    },
    color: {
      primary: "text-primary dark:text-primary",
      secondary: "text-secondary dark:text-secondary",
      muted: "text-gray dark:text-gray-300/80",
      danger: "text-danger dark:text-danger",
      warning: "text-warning dark:text-warning",
      info: "text-info dark:text-info",
      success: "text-success dark:text-success",
      black: "text-black dark:text-black",
      white: "text-white dark:text-white",
    },
    fontSize: {
      xs: "text-sm",
      sm: "text-[15px]",
      md: "text-[16px]",
      lg: "text-[18px]",
      xl: "text-[20px]",
      "2xl": "text-[24px]",
      "3xl": "text-[30px]",
      "4xl": "text-[36px]",
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
    fontSize: "md",
    fontWeight: "medium",
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
