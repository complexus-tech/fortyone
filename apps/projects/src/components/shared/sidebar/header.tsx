/* eslint-disable no-nested-ternary -- ok for the theme icons */
"use client";
import { useState } from "react";
import { Avatar, Badge, Box, Button, Flex, Menu, Text } from "ui";
import {
  ArrowDownIcon,
  CheckIcon,
  LogoutIcon,
  NewStoryIcon,
  PlusIcon,
  SettingsIcon,
  UsersAddIcon,
  SearchIcon,
  SystemIcon,
  MoonIcon,
  SunIcon,
  ArrowRightIcon,
  InvitesIcon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { useSession } from "next-auth/react";
import nProgress from "nprogress";
import { useTheme } from "next-themes";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import { useAnalytics, useLocalStorage } from "@/hooks";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useUserRole } from "@/hooks/role";
import { useCurrentWorkspace, useWorkspaces } from "@/lib/hooks/workspaces";
import { useProfile } from "@/lib/hooks/profile";
import { useMyInvitations } from "@/modules/invitations/hooks/my-invitations";
import { changeWorkspace, logOut } from "./actions";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

const clearAllStorage = () => {
  document.cookie.split(";").forEach((c) => {
    document.cookie = c
      .replace(/^ +/, "")
      .replace(/=.*/, `=;expires=${new Date().toUTCString()};path=/`);
  });

  if ("caches" in window) {
    caches.keys().then((names) => {
      names.forEach((name) => {
        caches.delete(name);
      });
    });
  }
};

export const Header = () => {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const { data: session } = useSession();
  const { data: profile } = useProfile();
  const { analytics } = useAnalytics();
  const [isOpen, setIsOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");
  const { userRole } = useUserRole();
  const { data: workspaces = [] } = useWorkspaces();
  const { data: myInvitations = [] } = useMyInvitations();
  const { workspace } = useCurrentWorkspace();

  useHotkeys("shift+n", () => {
    if (userRole !== "guest") {
      setIsOpen(true);
    }
  });

  useHotkeys("shift+o", () => {
    if (userRole !== "guest") {
      setIsObjectivesOpen(true);
    }
  });

  useHotkeys("shift+s", () => {
    if (userRole !== "guest") {
      setIsSprintsOpen(true);
    }
  });

  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
    } finally {
      clearAllStorage();
      window.location.href = "https://www.complexus.app";
    }
  };

  const handleChangeWorkspace = async (workspaceId: string, slug: string) => {
    try {
      nProgress.start();
      await changeWorkspace(workspaceId);
      if (domain.includes("localhost")) {
        window.location.href = `http://${slug}.${domain}/my-work`;
      } else {
        window.location.href = `https://${slug}.${domain}/my-work`;
      }
    } catch (error) {
      if (domain.includes("localhost")) {
        window.location.href = `http://${slug}.${domain}/my-work`;
      } else {
        window.location.href = `https://${slug}.${domain}/my-work`;
      }
    }
  };

  const handleCreateWorkspace = () => {
    if (domain.includes("localhost")) {
      window.location.href = `http://${domain}/onboarding/create`;
    } else {
      window.location.href = `https://${domain}/onboarding/create`;
    }
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
                  className="h-[1.6rem] text-sm"
                  name={workspace?.name}
                  rounded="md"
                  style={{
                    backgroundColor: workspace?.color,
                  }}
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
          <Menu.Items align="start" className="min-w-64 pt-0">
            <Menu.Group className="space-y-1 pt-1.5">
              {workspaces.map(({ id, name, color, slug, userRole }) => (
                <Menu.Item
                  className="justify-between gap-6"
                  key={id}
                  onSelect={() => handleChangeWorkspace(id, slug)}
                >
                  <span className="flex items-center gap-2">
                    <Avatar
                      className="h-[1.6rem] text-xs font-semibold tracking-wide"
                      name={name}
                      rounded="md"
                      style={{
                        backgroundColor: color,
                      }}
                    />
                    <span className="inline-block max-w-[20ch] truncate">
                      {name}
                    </span>
                    <Badge
                      className="bg-white px-1 text-[80%] font-medium uppercase"
                      color="tertiary"
                    >
                      {userRole}
                    </Badge>
                  </span>
                  {id === workspace?.id ? (
                    <CheckIcon className="shrink-0" strokeWidth={2.1} />
                  ) : null}
                </Menu.Item>
              ))}
            </Menu.Group>
            <Menu.Separator className="my-2" />
            <Menu.Group>
              <Menu.Item className="pl-3" onSelect={handleCreateWorkspace}>
                <PlusIcon />
                Create workspace
              </Menu.Item>
              {userRole === "admin" && (
                <>
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
                      Invite & manage members
                    </Link>
                  </Menu.Item>
                </>
              )}
            </Menu.Group>
            <Menu.Separator className="my-2" />
            <Menu.Group>
              <Menu.Item className="text-danger" onSelect={handleLogout}>
                <LogoutIcon className="h-5 w-auto text-danger dark:text-danger" />
                Log out
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <Menu>
          <Menu.Button>
            <Box className="relative">
              <Button
                asIcon
                color="tertiary"
                leftIcon={
                  <Avatar
                    className="relative h-[1.6rem] text-sm"
                    name={profile?.fullName || profile?.username}
                    src={profile?.avatarUrl}
                    style={{
                      backgroundColor: workspace?.color,
                    }}
                  />
                }
                size="sm"
                variant="naked"
              >
                <span className="sr-only">Profile</span>
              </Button>
              {myInvitations.length > 0 && (
                <Box className="absolute right-1 top-0.5 z-[2] size-2 animate-pulse rounded-full bg-primary" />
              )}
            </Box>
          </Menu.Button>
          <Menu.Items align="start" className="ml-4 pt-0">
            <Menu.Group className="px-4 pb-2 pt-2.5">
              <Text className="line-clamp-1" color="muted">
                {session?.user?.email}
              </Text>
            </Menu.Group>
            <Menu.Separator className="mb-2" />
            <Menu.Group>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings/account"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <SettingsIcon />
                  Account settings
                </Link>
              </Menu.Item>
              <Menu.SubMenu>
                <Menu.SubTrigger>
                  <span className="flex w-full items-center justify-between gap-4">
                    <span className="flex items-center gap-2">
                      {theme === "system" ? (
                        <SystemIcon className="h-[1.15rem]" />
                      ) : theme === "light" ? (
                        <SunIcon className="h-[1.15rem]" />
                      ) : (
                        <MoonIcon className="h-[1.15rem]" />
                      )}
                      Color theme
                    </span>
                    <span className="flex items-center gap-1">
                      <Text
                        className="text-[0.95rem] first-letter:uppercase"
                        color="muted"
                      >
                        {theme === "system"
                          ? "Sync with system"
                          : theme === "light"
                            ? "Day mode"
                            : "Night mode"}
                      </Text>
                      <ArrowRightIcon className="h-4" />
                    </span>
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
                      Day mode
                    </Menu.Item>
                    <Menu.Item
                      active={theme === "dark"}
                      onSelect={() => {
                        setTheme("dark");
                      }}
                    >
                      <MoonIcon className="h-[1.15rem]" />
                      Night mode
                    </Menu.Item>
                    <Menu.Item
                      active={theme === "system"}
                      onSelect={() => {
                        setTheme("system");
                      }}
                    >
                      <SystemIcon className="h-[1.15rem]" />
                      Sync with system
                    </Menu.Item>
                  </Menu.Group>
                </Menu.SubItems>
              </Menu.SubMenu>
              {myInvitations.length > 0 && (
                <Menu.Item>
                  <Link
                    className="flex w-full items-center justify-between gap-2"
                    href="/settings/invitations"
                    onClick={() => {
                      setPathBeforeSettings(pathname);
                    }}
                  >
                    <Flex gap={2}>
                      <InvitesIcon
                        className="relative top-px"
                        strokeWidth={2.5}
                      />
                      My invitations
                    </Flex>
                    <Badge rounded="full" size="sm">
                      {myInvitations.length}
                    </Badge>
                  </Link>
                </Menu.Item>
              )}
            </Menu.Group>
            <Menu.Separator className="my-2" />
            <Menu.Group>
              <Menu.Item className="text-danger" onSelect={handleLogout}>
                <LogoutIcon className="h-5 w-auto text-danger dark:text-danger" />
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
          disabled={userRole === "guest"}
          fullWidth
          leftIcon={<NewStoryIcon />}
          onClick={() => {
            if (userRole !== "guest") {
              setIsOpen(!isOpen);
            }
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
