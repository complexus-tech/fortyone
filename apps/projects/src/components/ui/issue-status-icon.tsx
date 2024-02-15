import {
  CheckCircle2,
  Circle,
  CircleDashed,
  PauseCircle,
  XCircle,
} from "lucide-react";
import { cn } from "lib";
import type { IssueStatus } from "../../types/issue";

export const IssueStatusIcon = ({
  status = "Backlog",
  className,
}: {
  status?: IssueStatus;
  className?: string;
}) => {
  return (
    <>
      {status === "Backlog" && (
        <CircleDashed
          className={cn("h-[1.15rem] w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {status === "Todo" && (
        <Circle
          className={cn("h-[1.15rem] w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {status === "In Progress" && (
        <Circle
          className={cn("h-[1.15rem] w-auto text-warning", className)}
          strokeWidth={2.5}
        />
      )}
      {status === "Testing" && (
        <Circle
          className={cn("h-[1.15rem] w-auto text-info", className)}
          strokeWidth={2.5}
        />
      )}
      {status === "Paused" && (
        <PauseCircle
          className={cn(
            "h-[1.15rem] w-auto text-dark-50 dark:text-gray-200",
            className,
          )}
          strokeWidth={2.5}
        />
      )}
      {status === "Done" && (
        <CheckCircle2
          className={cn("h-[1.15rem] w-auto text-success", className)}
          strokeWidth={2.5}
        />
      )}
      {status === "Canceled" && (
        <XCircle
          className={cn("h-[1.15rem] w-auto text-danger", className)}
          strokeWidth={2.5}
        />
      )}
      {status === "Duplicate" && (
        <XCircle
          className={cn("h-[1.15rem] w-auto text-dark-50", className)}
          strokeWidth={2.5}
        />
      )}
    </>
  );
};
