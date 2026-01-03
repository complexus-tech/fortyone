import { cn } from "lib";

export const Divider = ({ className = "" }) => {
  return <div className={cn("border-border border-t-[0.5px]", className)} />;
};
