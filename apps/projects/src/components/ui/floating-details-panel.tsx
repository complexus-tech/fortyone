import type { ComponentProps } from "react";
import { cn } from "lib";
import { Box } from "ui";

export const FloatingDetailsPanel = ({
  children,
  className,
  style,
  ...props
}: ComponentProps<typeof Box>) => (
  <Box
    className={cn(
      "border-border/70 dark:border-border dark:bg-surface absolute top-14 right-3 bottom-4 isolate z-40 w-[calc(100%-1.5rem)] rounded-xl border bg-white/95 shadow-xl backdrop-blur-xl md:top-[1.625rem] md:right-6 md:bottom-[4.875rem] md:w-[34rem]",
      className,
    )}
    style={style}
  >
    {/* Keep clipping separate from the squircle fallback's shadow layer. */}
    <Box
      className="h-full overflow-y-auto rounded-[calc(var(--radius-xl)-1px)] outline-none"
      {...props}
    >
      {children}
    </Box>
  </Box>
);
