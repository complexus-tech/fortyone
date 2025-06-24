import { cn } from "lib";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { Dot } from "./dot";

export const ObjectiveStatusIcon = ({
  statusId,
  className,
}: {
  statusId?: string;
  className?: string;
}) => {
  const { data: statuses = [] } = useObjectiveStatuses();
  if (!statuses.length) return null;
  const state =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  return <Dot className={cn("size-3", className)} color={state?.color} />;
};
