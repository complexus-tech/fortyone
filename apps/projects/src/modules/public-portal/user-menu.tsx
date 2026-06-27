"use client";

import Link from "next/link";
import { useTheme } from "next-themes";
import {
  ArrowRightIcon,
  DashboardIcon,
  LogoutIcon,
  MoonIcon,
  SettingsIcon,
  SunIcon,
  SystemIcon,
} from "icons";
import { Avatar, Button, Flex, Menu, Text } from "ui";
import { logOut } from "@/components/shared/sidebar/actions";
import { clearAllStorage } from "@/components/shared/sidebar/utils";
import type { PublicPortalViewer } from "./types";

const getSignedOutUrl = () =>
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app"
    ? "https://fortyone.app?signedOut=true"
    : "/?signedOut=true";

const getThemeLabel = (theme?: string) => {
  if (theme === "light") {
    return "Day mode";
  }

  if (theme === "dark") {
    return "Night mode";
  }

  return "Sync with system";
};

const ThemeIcon = ({ theme }: { theme?: string }) => {
  if (theme === "light") {
    return <SunIcon className="h-[1.15rem]" />;
  }

  if (theme === "dark") {
    return <MoonIcon className="h-[1.15rem]" />;
  }

  return <SystemIcon className="h-[1.15rem]" />;
};

export const PublicPortalUserMenu = ({
  viewer,
}: {
  viewer: PublicPortalViewer;
}) => {
  const { theme, setTheme } = useTheme();

  const handleLogout = async () => {
    try {
      await logOut();
      clearAllStorage();
      window.location.href = getSignedOutUrl();
    } catch {
      clearAllStorage();
      window.location.href = getSignedOutUrl();
    }
  };

  return (
    <Menu>
      <Menu.Button>
        <Button
          aria-label="Open account menu"
          asIcon
          className="size-10 rounded-full p-0"
          color="tertiary"
          rounded="full"
          variant="naked"
        >
          <Avatar
            className="!size-9 text-sm font-semibold"
            name={viewer.name}
            rounded="full"
            size="md"
            src={viewer.avatarUrl}
          />
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="w-80 rounded-3xl pt-2" sideOffset={8}>
        <Menu.Group className="px-4 pt-2.5 pb-2">
          <Text className="line-clamp-1" fontWeight="semibold">
            {viewer.name}
          </Text>
          <Text className="line-clamp-1 text-[0.95rem]" color="muted">
            {viewer.email}
          </Text>
        </Menu.Group>
        <Menu.Separator className="mb-2" />
        <Menu.Group>
          <Menu.Item className="rounded-2xl">
            <Link
              className="flex w-full items-center gap-2"
              href={viewer.appHref}
            >
              <DashboardIcon className="h-[1.15rem]" />
              Open app
            </Link>
          </Menu.Item>
          <Menu.Item className="rounded-2xl">
            <Link
              className="flex w-full items-center gap-2"
              href={viewer.accountHref}
            >
              <SettingsIcon className="h-[1.15rem]" />
              Account settings
            </Link>
          </Menu.Item>
          <Menu.SubMenu>
            <Menu.SubTrigger className="rounded-2xl">
              <span className="flex w-full items-center justify-between gap-4">
                <span className="flex items-center gap-2">
                  <ThemeIcon theme={theme} />
                  Appearance
                </span>
                <Flex align="center" gap={1}>
                  <Text
                    className="hidden text-[0.95rem] first-letter:uppercase md:block"
                    color="muted"
                  >
                    {getThemeLabel(theme)}
                  </Text>
                  <ArrowRightIcon className="h-4" />
                </Flex>
              </span>
            </Menu.SubTrigger>
            <Menu.SubItems className="rounded-2xl pt-1.5 md:w-48">
              <Menu.Group>
                <Menu.Item
                  active={theme === "light"}
                  className="rounded-2xl"
                  onSelect={() => {
                    setTheme("light");
                  }}
                >
                  <SunIcon className="h-[1.15rem]" />
                  Day mode
                </Menu.Item>
                <Menu.Item
                  active={theme === "dark"}
                  className="rounded-2xl"
                  onSelect={() => {
                    setTheme("dark");
                  }}
                >
                  <MoonIcon className="h-[1.15rem]" />
                  Night mode
                </Menu.Item>
                <Menu.Item
                  active={theme === "system"}
                  className="rounded-2xl"
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
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          <Menu.Item
            className="rounded-2xl text-danger"
            onSelect={handleLogout}
          >
            <LogoutIcon className="text-danger h-5 w-auto" />
            Log out
          </Menu.Item>
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
