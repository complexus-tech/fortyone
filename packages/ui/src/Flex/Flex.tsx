import { VariantProps, cva } from "cva";
import { cn } from "lib";
import { FC, HTMLAttributes, ComponentType } from "react";

const flex = cva("flex", {
  variants: {
    align: {
      start: "items-start",
      end: "items-end",
      center: "items-center",
      stretch: "items-stretch",
    },
    justify: {
      start: "justify-start",
      end: "justify-end",
      center: "justify-center",
      between: "justify-between",
      around: "justify-around",
    },
    direction: {
      row: "flex-row",
      "row-reverse": "flex-row-reverse",
      column: "flex-col",
      "column-reverse": "flex-col-reverse",
    },
    wrap: {
      true: "flex-wrap",
    },
    gap: {
      0: "gap-0",
      1: "gap-1",
      2: "gap-2",
      3: "gap-3",
      4: "gap-4",
      5: "gap-5",
      6: "gap-6",
      7: "gap-7",
      8: "gap-8",
      9: "gap-9",
      10: "gap-10",
      12: "gap-12",
      16: "gap-16",
      20: "gap-20",
    },
    gapX: {
      0: "gap-x-0",
      1: "gap-x-1",
      2: "gap-x-2",
      3: "gap-x-3",
      4: "gap-x-4",
      5: "gap-x-5",
      6: "gap-x-6",
      7: "gap-x-7",
      8: "gap-x-8",
      9: "gap-x-9",
      10: "gap-x-10",
      12: "gap-x-12",
      16: "gap-x-16",
      20: "gap-x-20",
    },
    gapY: {
      0: "gap-y-0",
      1: "gap-y-1",
      2: "gap-y-2",
      3: "gap-y-3",
      4: "gap-y-4",
      5: "gap-y-5",
      6: "gap-y-6",
      7: "gap-y-7",
      8: "gap-y-8",
      9: "gap-y-9",
      10: "gap-y-10",
      12: "gap-y-12",
      16: "gap-y-16",
      20: "gap-y-20",
    },
  },
});

interface Props extends HTMLAttributes<HTMLElement>, VariantProps<typeof flex> {
  as?: ComponentType<any>;
}

export const Flex: FC<Props> = ({
  as: Tag = "div",
  align,
  justify,
  direction,
  className,
  wrap,
  gap,
  gapX,
  gapY,
  children,
  ...rest
}) => {
  const classes = cn(
    flex({ align, justify, direction, gap, gapX, gapY, wrap }),
    className
  );

  return (
    <Tag className={classes} {...rest}>
      {children}
    </Tag>
  );
};
