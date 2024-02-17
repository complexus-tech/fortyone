import { ChevronDown } from "lucide-react";
import { Box, Button, Flex, Text, Menu, Wrapper } from "ui";
import type { ActivityProps } from "@/components/ui";
import { Activity } from "@/components/ui";

export const Activities = () => {
  const activites: ActivityProps[] = [
    {
      id: 1,
      user: "josemukorivo",
      action: "changed status from",
      prevValue: "Todo",
      newValue: "In Progress",
      timestamp: "23 hours ago",
    },
    {
      id: 2,
      user: "janedoe",
      action: "created task",
      prevValue: "Issue 1",
      newValue: "Todo",
      timestamp: "2 days ago",
    },
    {
      id: 3,
      user: "johnsmith",
      action: "assigned task to",
      prevValue: "jackdoe",
      newValue: "janedoe",
      timestamp: "1 week ago",
    },
    {
      id: 4,
      user: "johndoe",
      action: "changed status from",
      prevValue: "In Progress",
      newValue: "Done",
      timestamp: "1 hour ago",
    },
    {
      id: 5,
      user: "josemukorivo",
      action: "changed status from",
      prevValue: "Todo",
      newValue: "In Progress",
      timestamp: "23 hours ago",
    },
    {
      id: 6,
      user: "janedoe",
      action: "created task",
      prevValue: "Issue 1",
      newValue: "Todo",
      timestamp: "2 days ago",
    },
    {
      id: 7,
      user: "johnsmith",
      action: "assigned task to",
      prevValue: "jackdoe",
      newValue: "janedoe",
      timestamp: "1 week ago",
    },
  ];

  return (
    <Wrapper>
      <Flex align="center" className="mb-5" justify="between">
        <Text fontSize="lg">Recent activities</Text>
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

      <Flex className="relative" direction="column" gap={5}>
        <Box className="pointer-events-none absolute left-4 top-0 z-0 h-full border-l border-gray-100 dark:border-dark-200" />
        {activites.map((activity) => (
          <Activity key={activity.id} {...activity} />
        ))}
      </Flex>
    </Wrapper>
  );
};
