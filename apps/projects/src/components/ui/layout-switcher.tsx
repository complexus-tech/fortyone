import { Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, TableIcon } from "icons";
import type { StoriesLayout } from "@/components/ui";

export const LayoutSwitcher = ({
  layout = "list",
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
        "mr-2 h-[2.3rem] items-center rounded-[0.45rem] bg-gray-100/30 p-[0.2rem] dark:bg-dark-200/50",
        className,
      )}
    >
      <button
        className={cn(
          "flex h-full items-center gap-1 rounded-[0.45rem] px-2.5 font-medium dark:text-white/60 hover:dark:text-gray-100",
          {
            "border-[0.5px] border-gray-200/80 bg-white dark:border-dark-50 dark:bg-dark-200/80 dark:text-gray-100":
              layout === "list",
          },
        )}
        onClick={() => {
          setLayout("list");
        }}
        title="List view"
        type="button"
      >
        <TableIcon className="h-4 w-auto" />
        List view
      </button>
      <button
        className={cn(
          "flex h-full items-center gap-1 rounded-[0.45rem] px-2.5 font-medium dark:text-white/60 hover:dark:text-gray-100",
          {
            "border-[0.5px] border-gray-200/80 bg-white dark:border-dark-50 dark:bg-dark-200/80 dark:text-gray-100":
              layout === "kanban",
          },
        )}
        onClick={() => {
          setLayout("kanban");
        }}
        title="Board"
        type="button"
      >
        <KanbanIcon className="h-5 w-auto" strokeWidth={2.5} />
        Kanban
      </button>
    </Flex>
  );
};
