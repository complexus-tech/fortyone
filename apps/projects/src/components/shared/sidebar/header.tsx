"use client";
import { useState } from "react";
import { Avatar, Button, Flex, Menu, Text } from "ui";
import {
  ArrowDownIcon,
  CheckIcon,
  LogoutIcon,
  NewStoryIcon,
  PlusIcon,
  ObjectiveIcon,
  SettingsIcon,
  SprintsIcon,
  UsersAddIcon,
  EpicsIcon,
  SidebarExpandIcon,
  SearchIcon,
} from "icons";
import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { NewStoryDialog } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { logOut } from "./actions";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");

  const callbackUrl = `${pathname}?${searchParams.toString()}`;

  useHotkeys("c", () => {
    setIsOpen(true);
  });

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
                  name="Complexus Technologies"
                  rounded="md"
                  size="xs"
                  src="/complexus.png"
                />
              }
              rightIcon={
                <ArrowDownIcon className="relative top-[0.5px] h-3.5 w-auto" />
              }
              size="sm"
              variant="naked"
            >
              Complexus
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
            <Menu.Separator className="my-2" />
            <Menu.Group>
              <Menu.Item
                onClick={async () => {
                  await logOut(callbackUrl);
                }}
              >
                <LogoutIcon className="h-5 w-auto text-danger" />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Flex gap={2} className="mb-4">
        <Button
          className="rounded-[0.6rem] md:h-[2.5rem]"
          fullWidth
          color="tertiary"
          variant="outline"
          leftIcon={<NewStoryIcon className="h-5 w-auto" />}
          onClick={() => {
            setIsOpen(!isOpen);
          }}
        >
          Create Story
        </Button>
        <Button
          asIcon
          className="rounded-[0.6rem] md:h-[2.5rem]"
          color="tertiary"
          leftIcon={<SearchIcon className="h-4 w-auto" strokeWidth={3} />}
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
