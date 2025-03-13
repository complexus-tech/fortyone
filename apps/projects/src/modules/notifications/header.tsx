"use client";
import { Button, Flex, Menu, Text } from "ui";
import {
  DeleteIcon,
  FilterIcon,
  MoreVerticalIcon,
  NotificationsCheckIcon,
  NotificationsIcon,
  NotificationsUnreadIcon,
  SettingsIcon,
} from "icons";

export const NotificationsHeader = () => {
  return (
    <Flex
      align="center"
      className="h-16 border-b-[0.5px] border-gray-100/70 px-4 dark:border-dark-50"
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
              leftIcon={<FilterIcon className="h-4 w-auto" />}
              size="sm"
            >
              <div className="sr-only">Filter</div>
            </Button>
          </Menu.Button>
          <Menu.Items className="w-54">
            <Menu.Group className="mb-3 mt-1 px-4">
              <Text color="muted" textOverflow="truncate">
                Filter notifications
              </Text>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <NotificationsIcon className="h-5 w-auto" strokeWidth={2} />
                All notifications
              </Menu.Item>
              <Menu.Item>
                <NotificationsUnreadIcon className="h-5 w-auto" />
                Unread notifications
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              rightIcon={<MoreVerticalIcon className="h-4 w-auto" />}
              size="sm"
            >
              <span className="sr-only">More options</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group className="mb-3 mt-1 px-4">
              <Text color="muted" textOverflow="truncate">
                Manage notifications
              </Text>
            </Menu.Group>
            <Menu.Separator className="mb-1.5" />
            <Menu.Group>
              <Menu.Item>
                <NotificationsCheckIcon className="h-5 w-auto" />
                Mark all as read
              </Menu.Item>
              <Menu.Item>
                <SettingsIcon className="h-5 w-auto" />
                Notification settings
              </Menu.Item>
              <Menu.Separator />
              <Menu.Item>
                <DeleteIcon className="h-5 w-auto" />
                Delete all notifications
              </Menu.Item>
              <Menu.Item>
                <NotificationsCheckIcon className="h-5 w-auto" />
                Delete read notifications
              </Menu.Item>
              <Menu.Item>
                <NotificationsUnreadIcon className="h-5 w-auto" />
                Delete unread notifications
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Flex>
  );
};
