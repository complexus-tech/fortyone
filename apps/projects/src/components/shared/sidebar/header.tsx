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
  SearchIcon,
  SettingsIcon,
  SprintsIcon,
  UserIcon,
  UsersAddIcon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { NewStoryDialog } from "@/components/ui";
import { useLocalStorage } from "@/hooks";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");
  useHotkeys("c", () => {
    setIsOpen(true);
  });

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
              <Menu.Item>
                <LogoutIcon className="h-5 w-auto text-danger" />
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
              <Text
                className="text-[1.05rem]"
                color="muted"
                fontWeight="medium"
                textOverflow="truncate"
              >
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
                <LogoutIcon className="h-5 w-auto text-danger" />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Flex align="center" className="mb-4" gap={2} justify="between">
        <Flex className="w-full">
          <Button
            className="rounded-r-none shadow-sm"
            color="tertiary"
            fullWidth
            leftIcon={<NewStoryIcon className="h-5 w-auto" />}
            onClick={() => {
              setIsOpen(!isOpen);
            }}
            variant="outline"
          >
            Create Story
          </Button>
          <Menu>
            <Menu.Button>
              <Button
                align="center"
                className="rounded-l-none border-l-0 px-[0.65rem] shadow-sm"
                color="tertiary"
                leftIcon={<ArrowDownIcon className="h-[1.1rem] w-auto" />}
                variant="outline"
              >
                <span className="sr-only">More</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="w-56 pb-1">
              <Menu.Group className="gap-4 space-y-1">
                <Menu.Item>
                  <NewStoryIcon className="h-[1.1rem] w-auto" />
                  Create story
                </Menu.Item>
                <Menu.Item>
                  <ObjectiveIcon className="h-[1.1rem] w-auto" />
                  Create objective
                </Menu.Item>
                <Menu.Item>
                  <SprintsIcon className="h-[1.1rem] w-auto" />
                  Create sprint
                </Menu.Item>
                <Menu.Item>
                  <LogoutIcon className="h-[1.1rem] w-auto" />
                  Create story
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
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
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
