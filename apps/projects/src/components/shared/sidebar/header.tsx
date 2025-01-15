"use client";
import { useState } from "react";
import { Avatar, Button, Flex, Menu, Text } from "ui";
import {
  ArrowDownIcon,
  CheckIcon,
  LogoutIcon,
  NewStoryIcon,
  PlusIcon,
  SettingsIcon,
  UsersAddIcon,
  SearchIcon,
  MoonIcon,
  ArrowRightIcon,
  SunIcon,
  SystemIcon,
  PreferencesIcon,
} from "icons";
import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { NewStoryDialog } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useWorkspace } from "@/lib/hooks/workspace";
import { logOut } from "./actions";

export const Header = () => {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { theme, setTheme } = useTheme();
  const { data: workspace } = useWorkspace();
  const [isOpen, setIsOpen] = useState(false);
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");

  const callbackUrl = `${pathname}?${searchParams.toString()}`;

  useHotkeys("c", () => {
    setIsOpen(true);
  });

  const handleLogout = async () => {
    await logOut(callbackUrl).catch(() => {
      toast.error("Failed to log out");
    });
  };

  return (
    <>
      <Flex align="center" className="h-16" justify="between">
        <Menu>
          <Menu.Button>
            <Button
              className="gap-2 pl-1"
              color="tertiary"
              leftIcon={
                <Avatar
                  className="h-6"
                  name={workspace?.name}
                  rounded="md"
                  src="/complexus.png"
                />
              }
              rightIcon={
                <ArrowDownIcon className="relative top-[0.5px] h-3.5 w-auto text-gray dark:text-gray-300" />
              }
              size="sm"
              variant="naked"
            >
              <span className="max-w-[18ch] truncate">{workspace?.name}</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="start" className="w-72 pt-0">
            <Menu.Group className="px-4 py-2.5">
              <Text>Workspaces</Text>
            </Menu.Group>
            <Menu.Separator className="my-0" />
            <Menu.Group className="pt-1.5">
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
                  name="Amaka Studio"
                  rounded="md"
                  size="xs"
                />
                Amaka Studio
              </Menu.Item>
              <Menu.Item asChild>
                <Button
                  color="tertiary"
                  leftIcon={<PlusIcon className="h-5 w-auto" />}
                  size="sm"
                  variant="naked"
                >
                  Create workspace
                </Button>
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator className="my-2" />
            <Menu.Group>
              <Menu.SubMenu>
                <Menu.SubTrigger>
                  <span className="flex w-full items-center justify-between gap-1.5">
                    <span className="flex items-center gap-2">
                      <PreferencesIcon className="h-[1.15rem] w-auto" />
                      Appearance
                    </span>
                    <ArrowRightIcon className="h-4" />
                  </span>
                </Menu.SubTrigger>
                <Menu.SubItems className="rounded-xl pt-1.5 md:w-48">
                  <Menu.Group>
                    <Menu.Item
                      active={theme === "light"}
                      onSelect={() => {
                        setTheme("light");
                      }}
                    >
                      <SunIcon className="h-[1.15rem]" />
                      Light mode
                    </Menu.Item>
                    <Menu.Item
                      active={theme === "dark"}
                      onSelect={() => {
                        setTheme("dark");
                      }}
                    >
                      <MoonIcon className="h-[1.15rem]" />
                      Dark mode
                    </Menu.Item>
                    <Menu.Item
                      active={theme === "system"}
                      onSelect={() => {
                        setTheme("system");
                      }}
                    >
                      <SystemIcon className="h-[1.15rem]" />
                      System default
                    </Menu.Item>
                  </Menu.Group>
                </Menu.SubItems>
              </Menu.SubMenu>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <SettingsIcon className="h-[1.15rem]" />
                  Workspace settings
                </Link>
              </Menu.Item>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings/workspace/members"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <UsersAddIcon className="h-[1.3rem] w-auto" />
                  Invite members
                </Link>
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator className="my-2" />
            <Menu.Group>
              <Menu.Item onSelect={handleLogout}>
                <LogoutIcon className="h-5 w-auto text-danger" />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Flex className="mb-4" gap={2}>
        <Button
          className="rounded-[0.6rem] md:h-[2.5rem]"
          color="tertiary"
          fullWidth
          leftIcon={
            <NewStoryIcon className="h-5 w-auto text-gray dark:text-gray-300" />
          }
          onClick={() => {
            setIsOpen(!isOpen);
          }}
          variant="outline"
        >
          Create Story
          <Text
            as="span"
            className="ml-auto text-sm opacity-0 md:leading-[2.5rem]"
            color="muted"
          >
            ⌘N
          </Text>
        </Button>
        <Button
          asIcon
          className="rounded-[0.6rem] md:h-[2.5rem]"
          color="tertiary"
          leftIcon={<SearchIcon className="h-4" />}
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button>
      </Flex>
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
