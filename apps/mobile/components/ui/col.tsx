import React from "react";
import { View, ViewProps } from "react-native";
import { VariantProps, cva } from "cva";
import { cn } from "@/lib/utils";

const colVariants = cva("flex-col", {
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
    flex: {
      1: "flex-1",
      auto: "flex-auto",
      initial: "flex-initial",
      none: "flex-none",
    },
  },
  defaultVariants: {
    align: "start",
    justify: "start",
    gap: 0,
  },
});

export interface ColProps extends ViewProps, VariantProps<typeof colVariants> {}

export const Col = ({
  className,
  align,
  justify,
  gap,
  flex,
  children,
  ...props
}: ColProps) => {
  return (
    <View
      className={cn(colVariants({ align, justify, gap, flex }), className)}
      {...props}
    >
      {children}
    </View>
  );
};
