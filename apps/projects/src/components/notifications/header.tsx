"use client";
import { Button, Flex, Menu, Text } from "ui";
import {
  MoreVerticalIcon,
  NotificationsIcon,
  NotificationsUnreadIcon,
} from "icons";

export const NotificationsHeader = () => {
  return (
    <Flex
      align="center"
      className="h-16 border-b border-gray-100/70 px-5 dark:border-dark-200"
      justify="between"
    >
      <Text className="flex items-center gap-2">
        <NotificationsIcon className="h-5 w-auto" strokeWidth={2} />
        Notifications
      </Text>
      <Flex align="center" gap={2}>
        <Menu>
          <Menu.Button>
            <Button
              className="aspect-square"
              color="tertiary"
              leftIcon={
                <svg
                  className="h-4 w-auto"
                  fill="none"
                  height="24"
                  viewBox="0 0 24 24"
                  width="24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    d="M4 5L20 5"
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2.5"
                  />
                  <path
                    d="M18 12L6 12"
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2.5"
                  />
                  <path
                    d="M8 19L16 19"
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2.5"
                  />
                </svg>
              }
              size="sm"
            >
              <div className="sr-only">Filter</div>
            </Button>
          </Menu.Button>
          <Menu.Items className="w-40">
            <Menu.Group>
              <Menu.Item>
                <NotificationsIcon className="h-5 w-auto" strokeWidth={2} />
                All
              </Menu.Item>
              <Menu.Item>
                <NotificationsUnreadIcon className="h-5 w-auto" />
                Unread
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <Button
          color="tertiary"
          rightIcon={<MoreVerticalIcon className="h-4 w-auto" />}
          size="sm"
        >
          <span className="sr-only">Filter</span>
        </Button>
      </Flex>
    </Flex>
  );
};
