import { Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, ListIcon, GanttIcon } from "icons";
import type { StoriesLayout } from "@/components/ui";

export const LayoutSwitcher = ({
  layout = "list",
  options = ["list", "kanban", "gantt"],
  setLayout,
  className,
  disabled,
}: {
  layout: StoriesLayout;
  options?: StoriesLayout[];
  setLayout: (value: StoriesLayout) => void;
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
      {options.includes("list") && (
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
      )}

      {options.includes("kanban") && (
        <button
          className={cn(
            "flex h-full items-center gap-1.5 rounded-lg px-3 font-medium disabled:cursor-not-allowed dark:text-white/55 enabled:hover:dark:text-gray-100",
            {
              "border border-gray-100 bg-white dark:border-dark-50 dark:bg-dark-200/80 dark:text-gray-100":
                layout === "kanban",
            },
          )}
          disabled={disabled}
          onClick={() => {
            setLayout("kanban");
          }}
          title={disabled ? undefined : "Kanban Board"}
          type="button"
        >
          <KanbanIcon className="h-5 w-auto" />
          <span className="hidden md:inline">Kanban</span>
        </button>
      )}
      {options.includes("gantt") && (
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
          title={disabled ? undefined : "Timeline"}
          type="button"
        >
          <GanttIcon />
          <span className="hidden md:inline">Timeline</span>
        </button>
      )}
    </Flex>
  );
};
