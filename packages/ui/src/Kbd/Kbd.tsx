import { ReactNode } from "react";
import { Badge } from "../Badge/Badge";

export const Kbd = ({ children }: { children: ReactNode }) => (
  <Badge
    className="h-6 min-w-6 rounded-lg px-1 uppercase dark:border-dark-50"
    color="tertiary"
    size="sm"
  >
    {children}
  </Badge>
);
