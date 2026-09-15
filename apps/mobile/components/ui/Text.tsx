import type { TextProps as RNTextProps } from "react-native";
import type { VariantProps } from "cva";
import { useContext } from "react";
import { Text as RNText } from "react-native";
import { cn } from "@/lib/utils/classnames";
import { textVariants } from "./text-variants";
import { TextStyleContext } from "./text-style-context";

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
  const inherited = useContext(TextStyleContext);
  const classes = textVariants({
    align,
    color: color ?? inherited?.color,
    fontSize,
    fontWeight: fontWeight ?? inherited?.fontWeight,
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
