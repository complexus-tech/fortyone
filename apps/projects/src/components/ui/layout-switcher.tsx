import { Button, Flex } from "ui";
import { cn } from "lib";
import { KanbanIcon, TableIcon } from "icons";
import type { IssuesLayout } from "@/components/ui";

export const LayoutSwitcher = ({
  layout,
  setLayout,
  className,
}: {
  layout: IssuesLayout;
  setLayout: (value: IssuesLayout) => void;
  className?: string;
}) => {
  return (
    <Flex
      className={cn(
        "mr-2 items-center gap-0.5 rounded-lg bg-gray-100/50 dark:bg-dark-200/60",
        className,
      )}
    >
      <Button
        className={cn("opacity-60", {
          "border-gray-200 bg-white opacity-100 dark:border-dark-100 dark:bg-dark-300":
            layout === "kanban",
        })}
        color="tertiary"
        leftIcon={<KanbanIcon className="h-5 w-auto" strokeWidth={2.5} />}
        onClick={() => {
          setLayout("kanban");
        }}
        size="sm"
        title="Kanban layout"
        variant={layout === "kanban" ? "outline" : "naked"}
      >
        <span className="sr-only">Board</span>
      </Button>
      <Button
        className={cn("opacity-60", {
          "border-gray-200 bg-white opacity-100 dark:border-dark-100 dark:bg-dark-300":
            layout === "list",
        })}
        color="tertiary"
        leftIcon={<TableIcon className="h-5 w-auto" />}
        onClick={() => {
          setLayout("list");
        }}
        size="sm"
        title="List layout"
        variant={layout === "list" ? "outline" : "naked"}
      >
        <span className="sr-only">List</span>
      </Button>
    </Flex>
  );
};
