import { Circle } from "lucide-react";
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
        <Circle
          className={cn("h-[1.15rem] w-auto text-gray", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Todo" && (
        <Circle className={cn("h-[1.15rem] w-auto text-gray/60", className)} />
      )}
      {status === "In Progress" && (
        <Circle
          className={cn("h-[1.15rem] w-auto text-warning", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Testing" && (
        <Circle
          className={cn("h-[1.15rem] w-auto text-info", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Done" && (
        <Circle
          className={cn("h-[1.15rem] w-auto text-success", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Canceled" && (
        <Circle
          className={cn("h-5 w-auto text-danger", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Duplicate" && (
        <Circle
          className={cn("h-5 w-auto text-warning", className)}
          strokeWidth={2.3}
        />
      )}
    </>
  );
};
