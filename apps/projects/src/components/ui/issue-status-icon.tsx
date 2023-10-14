import { cn } from "lib";
import {
  TbBrandParsinta,
  TbCircleCheckFilled,
  TbCircleDashed,
  TbCircleXFilled,
  TbProgress,
  TbWashDryclean,
} from "react-icons/tb";
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
        <TbCircleDashed
          className={cn("h-5 w-auto text-gray", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Todo" && (
        <TbWashDryclean className={cn("h-5 w-auto text-gray/60", className)} />
      )}
      {status === "In Progress" && (
        <TbProgress
          className={cn("h-5 w-auto text-warning", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Testing" && (
        <TbBrandParsinta
          className={cn("h-5 w-auto text-info", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Done" && (
        <TbCircleCheckFilled
          className={cn("h-5 w-auto text-success", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Canceled" && (
        <TbCircleXFilled
          className={cn("h-5 w-auto text-danger", className)}
          strokeWidth={2.3}
        />
      )}
      {status === "Duplicate" && (
        <TbCircleXFilled
          className={cn("h-5 w-auto text-warning", className)}
          strokeWidth={2.3}
        />
      )}
    </>
  );
};
