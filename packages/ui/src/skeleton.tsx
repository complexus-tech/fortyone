import { Box } from "./box";
import type { BoxProps } from "./box";
import { cn } from "lib";

export const Skeleton = ({ className, ...rest }: BoxProps) => {
  return (
    <Box
      className={cn("animate-pulse rounded-lg bg-skeleton", className)}
      {...rest}
    />
  );
};
