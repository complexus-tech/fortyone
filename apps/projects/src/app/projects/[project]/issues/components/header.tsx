import { BreadCrumbs, Button, Flex, Tooltip } from "ui";
import { Kanban, Settings2, TableProperties } from "lucide-react";
import { cn } from "lib";
import type { Dispatch, SetStateAction } from "react";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";
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
        <Flex
          className="mr-2 items-center rounded-lg bg-gray-100/70 dark:bg-dark-300"
          gap={1}
        >
          <Tooltip title="Kanban Layout">
            <Button
              className={cn("opacity-80", {
                "border-gray-200 bg-white opacity-100 dark:border-dark-50 dark:bg-dark-200":
                  layout === "kanban",
              })}
              color="tertiary"
              leftIcon={<Kanban className="h-5 w-auto" strokeWidth={2.5} />}
              onClick={() => {
                setLayout("kanban");
              }}
              size="sm"
              variant={layout === "kanban" ? "outline" : "naked"}
            >
              <span className="sr-only">View as kanban board</span>
            </Button>
          </Tooltip>
          <Tooltip title="List Layout">
            <Button
              className={cn("opacity-80", {
                "border-gray-200 bg-white opacity-100 dark:border-dark-50 dark:bg-dark-200":
                  layout === "list",
              })}
              color="tertiary"
              leftIcon={<TableProperties className="h-5 w-auto" />}
              onClick={() => {
                setLayout("list");
              }}
              size="sm"
              variant={layout === "list" ? "outline" : "naked"}
            >
              <span className="sr-only">View as list</span>
            </Button>
          </Tooltip>
        </Flex>

        <Button
          color="tertiary"
          leftIcon={<Settings2 className="h-4 w-auto" />}
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
