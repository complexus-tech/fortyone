import type { ReactNode } from "react";
import { Badge } from "./badge";
import { cn } from "lib";

export const Kbd = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => (
  <Badge
    className={cn(
      "inline-flex h-6 min-w-6 whitespace-nowrap rounded-md px-1 uppercase tracking-wider dark:bg-surface-prominent/65",
      className,
    )}
    color="tertiary"
    size="sm"
  >
    {children}
  </Badge>
);
