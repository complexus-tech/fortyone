import { BreadCrumbs, Button, Flex } from "ui";
import { cn } from "lib";
import type { Dispatch, SetStateAction } from "react";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";
import { KanbanIcon, TableIcon, PreferencesIcon } from "@/components/icons";
import type { Layout } from "../types";

export const Header = ({
  layout = "kanban",
  setLayout,
}: {
  layout: string;
  setLayout: Dispatch<SetStateAction<Layout>>;
}) => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Web design",
          },
          { name: "Issues" },
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
