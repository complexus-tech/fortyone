"use client";
import { useState } from "react";
import { Avatar, Button, Flex, Menu, Text } from "ui";
import {
  CheckIcon,
  LogoutIcon,
  NewIssueIcon,
  PlusIcon,
  SearchIcon,
  SettingsIcon,
  UserIcon,
  UsersAddIcon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { NewIssueDialog } from "@/components/ui";
import { useLocalStorage } from "@/hooks";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");

  return (
    <>
      <Flex align="center" className="h-16" justify="between">
        <Menu>
          <Menu.Button>
            <Button
              className="pl-1"
              color="tertiary"
              leftIcon={
                <Avatar
                  name="Complexus Technologies"
                  rounded="md"
                  size="xs"
                  src="/complexus.png"
                />
              }
              size="sm"
              variant="naked"
            >
              Complexus
            </Button>
          </Menu.Button>
          <Menu.Items align="start" className="w-72">
            <Menu.Group className="mb-2 px-4">
              <Text>Workspaces</Text>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item className="justify-between">
                <span className="flex items-center gap-2">
                  <Avatar
                    name="Complexus Technologies"
                    rounded="md"
                    size="xs"
                    src="/complexus.png"
                  />
                  Complexus
                </span>
                <CheckIcon
                  className="h-5 w-auto text-primary"
                  strokeWidth={2.1}
                />
              </Menu.Item>
              <Menu.Item>
                <Avatar
                  color="secondary"
                  name="Fin Kenya"
                  rounded="md"
                  size="xs"
                />
                Fin Kenya
              </Menu.Item>
              <Menu.Item asChild>
                <Button
                  color="tertiary"
                  leftIcon={<PlusIcon className="h-5 w-auto" />}
                  variant="naked"
                >
                  Create workspace
                </Button>
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <SettingsIcon className="h-5 w-auto" />
                  Workspace settings
                </Link>
              </Menu.Item>
              <Menu.Item>
                <UsersAddIcon className="h-5 w-auto" />
                Invite members
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <LogoutIcon className="h-5 w-auto" />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>

        <Menu>
          <Menu.Button>
            <Button
              className="px-1"
              color="tertiary"
              leftIcon={
                <Avatar
                  name="Joseph Mukorivo"
                  size="sm"
                  src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                />
              }
              size="sm"
              variant="naked"
            >
              <span className="sr-only">Joseph Mukorivo</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="start" className="w-64 pb-1">
            <Menu.Group className="mb-3 mt-1 px-4">
              <Text color="muted" textOverflow="truncate">
                josemukorivo@gmail.com
              </Text>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/profile/josemukorivo"
                >
                  <UserIcon className="h-5 w-auto" />
                  View profile
                </Link>
              </Menu.Item>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings/account"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <SettingsIcon className="h-5 w-auto" />
                  Settings
                </Link>
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator className="mb-1.5" />
            <Menu.Group>
              <Menu.Item>
                <LogoutIcon className="h-5 w-auto" />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Flex align="center" className="mb-4" gap={2} justify="between">
        <Button
          className="shadow-sm"
          color="tertiary"
          fullWidth
          leftIcon={<NewIssueIcon className="h-5 w-auto" />}
          onClick={() => {
            setIsOpen(!isOpen);
          }}
          variant="outline"
        >
          New issue
        </Button>
        <Button
          align="center"
          className="px-[0.6rem] shadow-sm"
          color="tertiary"
          leftIcon={<SearchIcon className="h-[1.1rem] w-auto" />}
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button>
      </Flex>
      <NewIssueDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
