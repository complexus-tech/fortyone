import { ReactNode } from "react";
import { Badge } from "../Badge/Badge";
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
      "h-6 min-w-6 rounded-[0.4rem] px-1 uppercase tracking-wider dark:border-dark-50",
      className
    )}
    color="tertiary"
    size="sm"
  >
    {children}
  </Badge>
);
