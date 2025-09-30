import React from "react";
import { View, ViewProps } from "react-native";
import { VariantProps, cva } from "cva";
import { cn } from "@/lib/utils";

const rowVariants = cva("flex-row", {
  variants: {
    align: {
      start: "items-start",
      center: "items-center",
      end: "items-end",
      stretch: "items-stretch",
      baseline: "items-baseline",
    },
    justify: {
      start: "justify-start",
      center: "justify-center",
      end: "justify-end",
      between: "justify-between",
      around: "justify-around",
      evenly: "justify-evenly",
    },
    gap: {
      0: "gap-0",
      1: "gap-1",
      2: "gap-2",
      3: "gap-3",
      4: "gap-4",
      5: "gap-5",
      6: "gap-6",
      8: "gap-8",
      10: "gap-10",
      12: "gap-12",
    },
    wrap: {
      true: "flex-wrap",
      false: "flex-nowrap",
    },
    asContainer: {
      true: "px-4.5",
    },
  },
  defaultVariants: {
    align: "start",
    justify: "start",
    gap: 0,
    wrap: false,
  },
});

export interface RowProps extends ViewProps, VariantProps<typeof rowVariants> {}

export const Row = ({
  className,
  align,
  justify,
  gap,
  wrap,
  children,
  asContainer,
  ...props
}: RowProps) => {
  return (
    <View
      className={cn(
        rowVariants({ align, justify, gap, wrap, asContainer }),
        className
      )}
      {...props}
    >
      {children}
    </View>
  );
};
