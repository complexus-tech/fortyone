"use client";
import { Box, Button, Flex, Text, Menu, Wrapper } from "ui";
import { ArrowDownIcon } from "icons";

export const Activities = () => {
  return (
    <Wrapper>
      <Flex align="center" className="mb-5" justify="between">
        <Text fontSize="lg">Recent activities</Text>
        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
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
        {/* {activites.map((activity) => (
          <Activity key={activity.id} {...activity} />
        ))} */}
      </Flex>
    </Wrapper>
  );
};
