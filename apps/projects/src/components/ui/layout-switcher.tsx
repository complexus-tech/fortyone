import { Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, ListIcon, GanttIcon } from "icons";
import type { StoriesLayout } from "@/components/ui";

export const LayoutSwitcher = ({
  layout = "list",
  options = ["list", "kanban"],
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
        "bg-surface-muted h-[2.2rem] items-center gap-1 rounded-[0.7rem] md:mr-2",
        {
          "opacity-50": disabled,
        },
        className,
      )}
    >
      {options.includes("list") && (
        <button
          className={cn(
            "text-text-secondary enabled:hover:text-text-primary flex h-full items-center gap-1.5 rounded-[0.7rem] px-3 font-medium disabled:cursor-not-allowed",
            {
              "border-border bg-surface text-text-primary border":
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
            "text-text-secondary enabled:hover:text-text-primary flex h-full items-center gap-1.5 rounded-[0.7rem] px-3 font-medium disabled:cursor-not-allowed",
            {
              "border-border bg-surface text-text-primary border":
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
            "text-text-secondary enabled:hover:text-text-primary flex h-full items-center gap-1 rounded-[0.7rem] px-3 font-medium disabled:cursor-not-allowed",
            {
              "border-border bg-surface text-text-primary border":
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
