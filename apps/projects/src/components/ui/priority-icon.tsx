import { cn } from "lib";
import { BsFillExclamationSquareFill } from "react-icons/bs";
import {
  TbAntennaBars1,
  TbAntennaBars2,
  TbAntennaBars3,
  TbAntennaBars4,
} from "react-icons/tb";
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
        <TbAntennaBars1
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "Urgent" && (
        <BsFillExclamationSquareFill
          className={cn("h-[1.1rem] w-auto text-danger", className)}
        />
      )}
      {priority === "High" && (
        <TbAntennaBars4
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "Medium" && (
        <TbAntennaBars3
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
      {priority === "Low" && (
        <TbAntennaBars2
          className={cn("h-6 w-auto text-gray", className)}
          strokeWidth={2.5}
        />
      )}
    </>
  );
};
