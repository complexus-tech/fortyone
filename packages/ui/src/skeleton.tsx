import { HTMLAttributes } from "react";
import { Box } from "./box";
import { cn } from "lib";

export const Skeleton = ({
  className,
  ...rest
}: HTMLAttributes<HTMLDivElement>) => {
  return (
    <Box
      className={cn("animate-pulse rounded-lg bg-skeleton", className)}
      {...rest}
    />
  );
};
