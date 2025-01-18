import { Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, ListIcon } from "icons";
import type { StoriesLayout } from "@/components/ui";

export const LayoutSwitcher = ({
  layout = "list",
  setLayout,
  className,
  disabled,
}: {
  layout: StoriesLayout;
  setLayout: (value: StoriesLayout) => void;
  className?: string;
  disabled?: boolean;
}) => {
  return (
    <Flex
      className={cn(
        "mr-2 h-[2.26rem] items-center gap-1 rounded-lg bg-gray-50 dark:bg-dark-200/70",
        {
          "opacity-50": disabled,
        },
        className,
      )}
    >
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
        <ListIcon className="h-[1.1rem] w-auto" strokeWidth={3} />
        List
      </button>
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
        Kanban
      </button>
    </Flex>
  );
};
