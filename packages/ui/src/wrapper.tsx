import { cn } from "lib";
import { Box, BoxProps } from "./box";

export const Wrapper = ({ children, className, ...rest }: BoxProps) => {
  return (
    <Box
      className={cn(
        "rounded-2xl border-[0.5px] border-border py-4 px-4 md:px-5 shadow-lg shadow-shadow bg-surface",
        className,
      )}
      {...rest}
    >
      {children}
    </Box>
  );
};
