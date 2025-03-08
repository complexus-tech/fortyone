import { HTMLAttributes } from "react";
import { Box } from "../Box/Box";
import { cn } from "lib";

export const Skeleton = ({
  className,
  ...rest
}: HTMLAttributes<HTMLDivElement>) => {
  return (
    <Box
      className={cn(
        "animate-pulse rounded-lg bg-gray-100 dark:bg-dark-200",
        className
      )}
      {...rest}
    />
  );
};
