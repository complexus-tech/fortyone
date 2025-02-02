import { HealthIcon } from "icons";
import { cn } from "lib";
import type { ObjectiveHealth } from "@/modules/objectives/types";

export const ObjectiveHealthIcon = ({
  health,
}: {
  health?: ObjectiveHealth;
}) => {
  return (
    <HealthIcon
      className={cn({
        "text-success dark:text-success": health === "On Track",
        "text-warning dark:text-warning": health === "At Risk",
        "text-danger dark:text-danger": health === "Off Track",
      })}
    />
  );
};
