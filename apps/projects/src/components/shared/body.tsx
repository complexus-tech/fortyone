import { cn } from "lib";
import type { BoxProps } from "ui";
import { Box } from "ui";

export const BodyContainer = ({ children, className }: BoxProps) => {
  return (
    <Box className={cn("h-[calc(100dvh-3.6rem)] overflow-y-auto", className)}>
      {children}
    </Box>
  );
};
