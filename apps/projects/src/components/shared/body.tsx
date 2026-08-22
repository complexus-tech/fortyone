import { cn } from "lib";
import type { BoxProps } from "ui";
import { Box } from "ui";

export const BodyContainer = ({ children, className }: BoxProps) => {
  return (
    <Box
      className={cn(
        "h-(--app-page-content-height) min-h-0 overflow-y-auto",
        className,
      )}
      data-body-container
    >
      {children}
    </Box>
  );
};
