import { cn } from "lib";
import type { BoxProps } from "ui";
import { Box } from "ui";

export const BodyContainer = ({ children, className }: BoxProps) => {
  return (
    <Box className={cn("h-screen overflow-y-auto pt-16", className)}>
      {children}
    </Box>
  );
};
