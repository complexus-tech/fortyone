import { Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, ListIcon } from "icons";
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
        "mr-2 h-[2.25rem] items-center rounded-[0.46rem] border-[0.5px] border-gray-100/60 bg-gray-50/50 p-[0.18rem] dark:border-dark-100 dark:bg-dark-200/50",
        className,
      )}
    >
      <button
        className={cn(
          "flex h-full items-center gap-1 rounded-[0.35rem] px-2.5 font-medium dark:text-white/55 hover:dark:text-gray-100",
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
        <ListIcon className="h-[1.1rem] w-auto" />
        List
      </button>
      <button
        className={cn(
          "flex h-full items-center gap-1 rounded-[0.35rem] px-2.5 font-medium dark:text-white/55 hover:dark:text-gray-100",
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
        Board
      </button>
    </Flex>
  );
};
