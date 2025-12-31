import { cn } from "lib";
import { Box, BoxProps } from "./box";

export const Wrapper = ({ children, className, ...rest }: BoxProps) => {
  return (
    <Box
      className={cn(
        "rounded-2xl border border-gray-100/70 py-4 px-4 md:px-5 shadow-lg shadow-gray-50 dark:shadow-none dark:border-dark-100 bg-surface",
        className
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};
