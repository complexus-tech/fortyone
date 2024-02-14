import { cn } from "lib";
import { SignalMedium } from "lucide-react";
import type { IssuePriority } from "@/types/issue";

export const PriorityIcon = ({
  priority = "No Priority",
  className,
}: {
  priority?: IssuePriority;
  className?: string;
}) => {
  return (
    <>
      {priority === "No Priority" && (
        <SignalMedium
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "Urgent" && (
        <SignalMedium
          className={cn("h-[1.1rem] w-auto text-danger", className)}
        />
      )}
      {priority === "High" && (
        <SignalMedium
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "Medium" && (
        <SignalMedium
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "Low" && (
        <SignalMedium
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
    </>
  );
};
