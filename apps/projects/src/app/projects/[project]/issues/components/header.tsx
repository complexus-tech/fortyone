import { BreadCrumbs, Button, Flex, Tooltip } from "ui";
import { Kanban, Settings2, TableProperties } from "lucide-react";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Web design",
          },
          { name: "Assigned" },
        ]}
      />
      <Flex gap={2}>
        <Flex
          className="mr-2 items-center rounded-lg bg-gray-100/70 dark:bg-dark-300"
          gap={1}
        >
          <Tooltip title="Kanban Layout">
            <Button
              className="border-gray-200 bg-white dark:border-dark-50 dark:bg-dark-200"
              color="tertiary"
              leftIcon={<Kanban className="h-5 w-auto" strokeWidth={2.5} />}
              size="sm"
              variant="outline"
            >
              <span className="sr-only">View as kanban board</span>
            </Button>
          </Tooltip>
          <Tooltip title="Table Layout">
            <Button
              className="opacity-80"
              color="tertiary"
              leftIcon={<TableProperties className="h-5 w-auto" />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">View as table</span>
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
