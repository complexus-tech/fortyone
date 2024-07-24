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

        {/* <SidebarExpandIcon
          role="button"
          aria-label="Collapse"
          tabIndex={0}
          className="h-6 w-auto outline-none"
        /> */}

        {/* <Button
          align="center"
          className="px-[0.5rem] shadow"
          color="tertiary"
          leftIcon={
            <SearchIcon className="h-[0.95rem] w-auto" strokeWidth={3} />
          }
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button> */}
      </Flex>
      <Flex className="mb-3 w-full rounded-lg shadow">
        <Button
          className="h-9 rounded-r-none md:h-[2.65rem]"
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
              className="rounded-l-none border-l-0 px-[0.85rem] md:h-[2.65rem]"
              color="tertiary"
              leftIcon={<ArrowDownIcon className="h-[1.1rem] w-auto" />}
              variant="outline"
            >
              <span className="sr-only">More</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-64 pb-1">
            <Menu.Group className="gap-4 space-y-1">
              <Menu.Item>
                <NewStoryIcon className="h-5 w-auto" />
                Continue from Draft
              </Menu.Item>
              <Menu.Separator />
              <Menu.Item>
                <NewStoryIcon className="h-5 w-auto" />
                Create Story
              </Menu.Item>
              <Menu.Item>
                <EpicsIcon className="h-5 w-auto" />
                Create Epic
              </Menu.Item>
              <Menu.Item>
                <SprintsIcon className="h-5 w-auto" />
                Create Sprint
              </Menu.Item>
              <Menu.Item>
                <ObjectiveIcon className="h-5 w-auto" />
                Create Objective
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </>
  );
};
