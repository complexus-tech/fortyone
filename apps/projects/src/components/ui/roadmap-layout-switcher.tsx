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
        "bg-surface-muted h-[2.1rem] items-center gap-1 rounded-xl",
        {
          "opacity-50": disabled,
        },
        className,
      )}
    >
      <button
        className={cn(
          "text-text-secondary enabled:hover:text-text-primary flex h-full items-center gap-1 rounded-xl px-3 font-medium disabled:cursor-not-allowed",
          {
            "border-border bg-surface text-text-primary border-[0.5px]":
              layout === "gantt",
          },
        )}
        disabled={disabled}
        onClick={() => {
          setLayout("gantt");
        }}
        title={disabled ? undefined : "Timeline"}
        type="button"
      >
        <GanttIcon />
        <span className="hidden md:inline">Timeline</span>
      </button>
      <button
        className={cn(
          "text-text-secondary enabled:hover:text-text-primary flex h-full items-center gap-1.5 rounded-xl px-3 font-medium disabled:cursor-not-allowed",
          {
            "border-border bg-surface text-text-primary border-[0.5px]":
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
