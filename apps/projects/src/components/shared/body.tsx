import { cn } from "lib";
import type { BoxProps } from "ui";
import { Box } from "ui";

export const BodyContainer = ({ children, className }: BoxProps) => {
  return (
    <Box className={cn("h-[calc(100vh-4rem)] overflow-y-auto", className)}>
      {children}
    </Box>
  );
};
