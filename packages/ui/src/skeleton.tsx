import { HTMLAttributes } from "react";
import { Box } from "./box";
import { cn } from "lib";

export const Skeleton = ({
  className,
  ...rest
}: HTMLAttributes<HTMLDivElement>) => {
  return (
    <Box
      className={cn(
        "animate-pulse rounded-[0.6rem] bg-gray-100/80 dark:bg-dark-200/80",
        className
      )}
      {...rest}
    />
  );
};
