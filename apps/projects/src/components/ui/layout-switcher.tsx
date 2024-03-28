import { Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, TableIcon } from "icons";
import type { StoriesLayout } from "@/components/ui";

export const LayoutSwitcher = ({
  layout = "kanban",
  setLayout,
  className,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
  className?: string;
}) => {
  return (
    <Flex
      className={cn(
        "mr-2 h-[2.1rem] items-center rounded-lg bg-gray-50 dark:bg-dark-200/60",
        className,
      )}
    >
      <button
        className={cn("h-full rounded-lg px-2", {
          "border border-gray-100 bg-white dark:border-dark-50 dark:bg-dark-100/80":
            layout === "kanban",
        })}
        onClick={() => {
          setLayout("kanban");
        }}
        title="Board"
        type="button"
      >
        <KanbanIcon className="h-5 w-auto" strokeWidth={2.5} />
        <span className="sr-only">Board</span>
      </button>
      <button
        className={cn("h-full rounded-lg px-2", {
          "border border-gray-100 bg-white dark:border-dark-50 dark:bg-dark-100/80":
            layout === "list",
        })}
        onClick={() => {
          setLayout("list");
        }}
        title="List view"
        type="button"
      >
        <TableIcon className="h-5 w-auto" />
        <span className="sr-only">List</span>
      </button>
    </Flex>
  );
};
