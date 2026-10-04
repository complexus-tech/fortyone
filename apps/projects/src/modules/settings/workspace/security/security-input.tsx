import type { ComponentProps } from "react";
import { cn } from "lib";
import { Input } from "ui";

export const SecurityInput = ({
  className,
  ...props
}: ComponentProps<typeof Input>) => (
  <Input
    className={cn(
      "border-border h-[2.1rem] px-3 py-1 text-base leading-normal",
      className,
    )}
    {...props}
  />
);
