import { cn } from "lib";
import {
  AlertCircle,
  Ban,
  SignalHigh,
  SignalLow,
  SignalMedium,
} from "lucide-react";
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
        <Ban
          className={cn("h-[1.15rem] w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}

      {priority === "Urgent" && (
        <AlertCircle
          className={cn("h-[1.15rem] w-auto text-danger", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "High" && (
        <SignalHigh
          className={cn("relative -top-[2px] h-6 w-auto text-gray", className)}
          strokeWidth={2.6}
        />
      )}
      {priority === "Medium" && (
        <SignalMedium
          className={cn("relative -top-[2px] h-6 w-auto text-gray", className)}
          strokeWidth={2.6}
        />
      )}
      {priority === "Low" && (
        <SignalLow
          className={cn("relative -top-[2px] h-6 w-auto text-gray", className)}
          strokeWidth={2.6}
        />
      )}
    </>
  );
};
