import { cn } from "lib";
import { Box, BoxProps } from "../Box/Box";

export const Wrapper = ({ children, className, ...rest }: BoxProps) => {
  return (
    <Box
      className={cn(
        "rounded-xl border bg-gray-50/10 border-gray-100/60 p-4 shadow-sm dark:border-dark-200 dark:bg-dark-300/60",
        className
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};
