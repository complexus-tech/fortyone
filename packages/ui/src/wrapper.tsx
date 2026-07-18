import { cn } from "lib";
import { Box, BoxProps } from "./box";

export const Wrapper = ({ children, className, ...rest }: BoxProps) => {
  return (
    <Box
      className={cn(
        "rounded-2xl border-[0.5px] border-border bg-surface px-4 py-4 shadow-lg shadow-shadow md:px-5 dark:bg-surface/70",
        className,
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};
