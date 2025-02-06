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
import { useSession } from "next-auth/react";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useWorkspace } from "@/lib/hooks/workspace";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useUserRole } from "@/hooks/role";
import { logOut } from "./actions";

export const Header = () => {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { data: session } = useSession();
  const { theme, setTheme } = useTheme();
  const { data: workspace } = useWorkspace();
  const [isOpen, setIsOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");
  const { userRole } = useUserRole();

  const workspaces = session?.workspaces || [];

  const callbackUrl = `${pathname}?${searchParams.toString()}`;

  useHotkeys("shift+n", () => {
    setIsOpen(true);
  });

  useHotkeys("shift+o", () => {
    setIsObjectivesOpen(true);
  });

  useHotkeys("shift+s", () => {
    setIsSprintsOpen(true);
  });

  const handleLogout = async () => {
    await logOut(callbackUrl);
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
              {workspaces.map(({ id, name, isActive }) => (
                <Menu.Item className="justify-between" key={id}>
                  <span className="flex items-center gap-2">
                    <Avatar className="h-6 text-xs" name={name} rounded="md" />
                    <span className="inline-block max-w-[20ch] truncate">
                      {name}
                    </span>
                  </span>
                  {isActive ? (
                    <CheckIcon className="shrink-0" strokeWidth={2.1} />
                  ) : null}
                </Menu.Item>
              ))}

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
                  href={
                    userRole === "admin" ? "/settings" : "/settings/account"
                  }
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <SettingsIcon className="h-[1.15rem]" />
                  {userRole === "admin" ? "Workspace settings" : "Settings"}
                </Link>
              </Menu.Item>
              {userRole === "admin" && (
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
              )}
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
          leftIcon={<NewStoryIcon />}
          onClick={() => {
            setIsOpen(!isOpen);
          }}
          variant="outline"
        >
          Create Story
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
      <NewSprintDialog isOpen={isSprintsOpen} setIsOpen={setIsSprintsOpen} />
      <NewObjectiveDialog
        isOpen={isObjectivesOpen}
        setIsOpen={setIsObjectivesOpen}
      />
    </>
  );
};
