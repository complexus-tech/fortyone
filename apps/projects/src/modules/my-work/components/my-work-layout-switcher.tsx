import { CalendarIcon, KanbanIcon, ListIcon } from "icons";
import { cn } from "lib";
import { Flex } from "ui";
import type { ReactNode } from "react";
import type { MyWorkLayout } from "../types";

const options = [
  {
    icon: <ListIcon className="h-[1.1rem] w-auto" strokeWidth={3.5} />,
    label: "List",
    value: "list",
  },
  {
    icon: <KanbanIcon className="h-5 w-auto" />,
    label: "Board",
    value: "kanban",
  },
  {
    icon: <CalendarIcon className="h-[1.1rem] w-auto" />,
    label: "Calendar",
    value: "calendar",
  },
] as const satisfies readonly {
  icon: ReactNode;
  label: string;
  value: MyWorkLayout;
}[];

export const MyWorkLayoutSwitcher = ({
  layout,
  setLayout,
  showCalendar = true,
}: {
  layout: MyWorkLayout;
  setLayout: (value: MyWorkLayout) => void;
  showCalendar?: boolean;
}) => (
  <Flex className="bg-surface-muted h-[2.2rem] items-center gap-1 rounded-xl md:mr-2">
    {options
      .filter((option) => showCalendar || option.value !== "calendar")
      .map((option) => (
        <button
          aria-pressed={layout === option.value}
          className={cn(
            "text-text-secondary hover:text-text-primary flex h-full items-center gap-1.5 rounded-xl px-3 font-medium",
            layout === option.value &&
              "border-border text-text-primary dark:bg-surface border bg-white",
          )}
          key={option.value}
          onClick={() => {
            setLayout(option.value);
          }}
          title={`${option.label} view`}
          type="button"
        >
          {option.icon}
          <span className="hidden md:inline">{option.label}</span>
        </button>
      ))}
  </Flex>
);
