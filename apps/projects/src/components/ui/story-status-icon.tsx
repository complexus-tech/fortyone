import { cn } from "lib";
import { useStatuses } from "@/lib/hooks/statuses";
import type { StateCategory } from "@/types/states";
import { Dot } from "./dot";

export const StoryStatusIcon = ({
  statusId,
  className,
}: {
  statusId?: string;
  category?: StateCategory;
  className?: string;
}) => {
  const { data: statuses = [] } = useStatuses();
  if (!statuses.length) return null;
  const state =
    statuses.find((state) => state.id === statusId) || statuses.at(0);

  return <Dot className={cn("size-3", className)} color={state?.color} />;
};
