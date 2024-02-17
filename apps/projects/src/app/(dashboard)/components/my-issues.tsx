import { Calendar, ChevronDown } from "lucide-react";
import { Box, Button, Flex, Tabs, Text, Menu, Wrapper } from "ui";
import { cn } from "lib";
import { RowWrapper } from "@/components/ui";
import { IssueContextMenu } from "@/components/ui/issue/context-menu";
import { StatusesMenu } from "@/components/ui/issue/statuses-menu";
import { PrioritiesMenu } from "@/components/ui/issue/priorities-menu";
import { IssueStatusIcon } from "@/components/ui/issue-status-icon";
import { PriorityIcon } from "@/components/ui/priority-icon";

export const MyIssues = () => {
  return (
    <Wrapper>
      <Flex align="center" className="mb-3" justify="between">
        <Text fontSize="lg">Assigned to me</Text>
        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              rightIcon={<ChevronDown className="h-5 w-auto" />}
              size="sm"
              variant="outline"
            >
              Due this week
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item>Last week</Menu.Item>
              <Menu.Item>Last month</Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      <Tabs defaultValue="open">
        <Tabs.List className="mx-0">
          <Tabs.Tab value="open">Open</Tabs.Tab>
          <Tabs.Tab value="closed">Closed</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="open">
          <Box className="mt-4 border-t border-gray-50 dark:border-dark-200">
            {new Array(7).fill(0).map((_, i) => (
              <IssueContextMenu key={i}>
                <RowWrapper
                  className={cn("px-1", {
                    "border-b-0": i === 7 - 1,
                  })}
                >
                  <Flex align="center" className="relative select-none" gap={2}>
                    <PrioritiesMenu>
                      <PrioritiesMenu.Trigger>
                        <button className="block" type="button">
                          <PriorityIcon priority="No Priority" />
                        </button>
                      </PrioritiesMenu.Trigger>
                      <PrioritiesMenu.Items priority="No Priority" />
                    </PrioritiesMenu>
                    <Flex align="center" gap={2}>
                      <Text
                        className="w-[55px] truncate"
                        color="muted"
                        fontWeight="medium"
                      >
                        COM-12
                      </Text>
                      <StatusesMenu>
                        <StatusesMenu.Trigger>
                          <button className="block" type="button">
                            <IssueStatusIcon status="Backlog" />
                          </button>
                        </StatusesMenu.Trigger>
                        <StatusesMenu.Items status="Backlog" />
                      </StatusesMenu>
                      <Text className="overflow-hidden text-ellipsis whitespace-nowrap pl-2 hover:opacity-90">
                        Design a new homepage
                      </Text>
                    </Flex>
                  </Flex>
                  <Flex align="center" gap={3}>
                    <Text className="flex items-center gap-1" color="muted">
                      Sep 27
                      <Calendar className="h-4 w-auto" />
                    </Text>
                  </Flex>
                </RowWrapper>
              </IssueContextMenu>
            ))}
          </Box>
        </Tabs.Panel>
        <Tabs.Panel value="closed">
          <Text className="pt-1">Closed</Text>
        </Tabs.Panel>
      </Tabs>
    </Wrapper>
  );
};
