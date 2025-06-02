import { Flex } from "ui";
import { cn } from "lib";
import { ListIcon, GanttIcon } from "icons";
import type { RoadmapLayoutType } from "@/modules/roadmap/types";

export const RoadmapLayoutSwitcher = ({
  layout = "gantt",
  setLayout,
  className,
  disabled,
}: {
  layout: RoadmapLayoutType;
  setLayout: (value: RoadmapLayoutType) => void;
  className?: string;
  disabled?: boolean;
}) => {
  return (
    <Flex
      className={cn(
        "h-[2.26rem] items-center gap-1 rounded-lg bg-gray-50 dark:bg-dark-200/70 md:mr-2",
        {
          "opacity-50": disabled,
        },
        className,
      )}
    >
      <button
        className={cn(
          "flex h-full items-center gap-1 rounded-lg px-3 font-medium disabled:cursor-not-allowed dark:text-white/55 enabled:hover:dark:text-gray-100",
          {
            "border border-gray-100 bg-white dark:border-dark-50 dark:bg-dark-200/80 dark:text-gray-100":
              layout === "gantt",
          },
        )}
        disabled={disabled}
        onClick={() => {
          setLayout("gantt");
        }}
        title={disabled ? undefined : "Gantt Chart"}
        type="button"
      >
        <GanttIcon />
        <span className="hidden md:inline">Gantt</span>
      </button>
      <button
        className={cn(
          "flex h-full items-center gap-1.5 rounded-lg px-3 font-medium disabled:cursor-not-allowed dark:text-white/55 enabled:hover:dark:text-gray-100",
          {
            "border border-gray-100 bg-white dark:border-dark-50 dark:bg-dark-200/80 dark:text-gray-100":
              layout === "list",
          },
        )}
        disabled={disabled}
        onClick={() => {
          setLayout("list");
        }}
        title={disabled ? undefined : "List view"}
        type="button"
      >
        <ListIcon className="h-[1.1rem] w-auto" strokeWidth={3.5} />
        <span className="hidden md:inline">List</span>
      </button>
    </Flex>
  );
};
