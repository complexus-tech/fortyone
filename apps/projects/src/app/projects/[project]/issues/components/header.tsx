import { BreadCrumbs, Button, Flex } from "ui";
import { cn } from "lib";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";
import {
  KanbanIcon,
  TableIcon,
  PreferencesIcon,
  ProjectsIcon,
  IssuesIcon,
} from "icons";
import type { Layout } from "../types";

export const Header = ({
  layout = "kanban",
  setLayout,
}: {
  layout: Layout;
  setLayout: (value: Layout) => void;
}) => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Web design",
            icon: <ProjectsIcon className="h-4 w-auto" />,
          },
          {
            name: "Issues",
            icon: <IssuesIcon className="h-[1.1rem] w-auto" />,
          },
        ]}
      />
      <Flex gap={2}>
        <Flex className="mr-2 items-center gap-0.5 rounded-lg bg-gray-100/50 dark:bg-dark-200/60">
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

        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <NewIssueButton />
      </Flex>
    </HeaderContainer>
  );
};
