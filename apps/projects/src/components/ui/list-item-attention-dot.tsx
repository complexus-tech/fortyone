import { cn } from "lib";

export const ListItemAttentionDot = ({ className }: { className?: string }) => (
  <span
    aria-hidden="true"
    className={cn("bg-primary size-2 shrink-0 rounded-full", className)}
  />
);
