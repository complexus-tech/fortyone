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
import Link from "next/link";
import { useState } from "react";
import { ConfirmDialog } from "@/components/ui";
import { useNotifications } from "./hooks/notifications";
import { useUnreadNotifications } from "./hooks/unread";
import { useReadAllNotificationsMutation } from "./hooks/read-all-mutation";
import { useDeleteAllMutation } from "./hooks/delete-all-mutation";
import { useDeleteReadMutation } from "./hooks/delete-read-mutation";

export const NotificationsHeader = () => {
  const [isDeletingAll, setIsDeletingAll] = useState(false);
  const [isDeletingRead, setIsDeletingRead] = useState(false);
  const { data: notifications = [] } = useNotifications();
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const { mutate: markAllAsRead } = useReadAllNotificationsMutation();
  const { mutate: deleteAllNotifications } = useDeleteAllMutation();
  const { mutate: deleteReadNotifications } = useDeleteReadMutation();

  const handleMarkAllAsRead = () => {
    markAllAsRead();
  };

  const handleDeleteAllNotifications = () => {
    deleteAllNotifications();
    setIsDeletingAll(false);
  };

  const handleDeleteReadNotifications = () => {
    deleteReadNotifications();
    setIsDeletingRead(false);
  };

  return (
    <Flex
      align="center"
      className="h-16 border-b-[0.5px] border-gray-200/60 px-4 dark:border-dark-50"
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
                <NotificationsIcon />
                All notifications
              </Menu.Item>
              <Menu.Item>
                <NotificationsUnreadIcon />
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
              {unreadNotifications > 0 && (
                <Menu.Item onSelect={handleMarkAllAsRead}>
                  <NotificationsCheckIcon className="h-5 w-auto" />
                  Mark all as read
                </Menu.Item>
              )}
              <Menu.Item>
                <Link
                  className="flex items-center gap-2"
                  href="/settings/account/preferences"
                >
                  <SettingsIcon className="h-5 w-auto" />
                  Notification settings
                </Link>
              </Menu.Item>
            </Menu.Group>
            {notifications.length > 0 && (
              <Menu.Group>
                <Menu.Separator />
                <Menu.Item
                  onSelect={() => {
                    setIsDeletingAll(true);
                  }}
                >
                  <DeleteIcon className="h-5 w-auto" />
                  Delete all notifications
                </Menu.Item>
                {notifications.length > unreadNotifications && (
                  <Menu.Item
                    onSelect={() => {
                      setIsDeletingRead(true);
                    }}
                  >
                    <NotificationsCheckIcon className="h-5 w-auto" />
                    Delete read notifications
                  </Menu.Item>
                )}
              </Menu.Group>
            )}
          </Menu.Items>
        </Menu>
      </Flex>

      <ConfirmDialog
        description="Are you sure you want to delete all notifications? This action cannot be undone."
        isOpen={isDeletingAll}
        onClose={() => {
          setIsDeletingAll(false);
        }}
        onConfirm={handleDeleteAllNotifications}
        title="Delete all notifications"
      />

      <ConfirmDialog
        description="Are you sure you want to delete all read notifications? This action cannot be undone."
        isOpen={isDeletingRead}
        onClose={() => {
          setIsDeletingRead(false);
        }}
        onConfirm={handleDeleteReadNotifications}
        title="Delete all read notifications"
      />
    </Flex>
  );
};
