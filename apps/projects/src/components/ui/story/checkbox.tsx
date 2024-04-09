import { cn } from "lib";
import { Checkbox } from "ui";

export const TableCheckbox = ({ className }: { className?: string }) => {
  return <Checkbox className={cn("rounded-[0.35rem]", className)} />;
};
