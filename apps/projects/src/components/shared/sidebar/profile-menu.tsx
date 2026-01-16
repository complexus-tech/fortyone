/* eslint-disable no-nested-ternary -- ok for the theme icons */
import { Avatar, Badge, Box, Flex, Menu, Button, Text } from "ui";
import {
  LogoutIcon,
  SettingsIcon,
  SystemIcon,
  MoonIcon,
  SunIcon,
  ArrowRightIcon,
  InvitesIcon,
  ArrowRight2Icon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { useMyInvitations } from "@/modules/invitations/hooks/my-invitations";
import { useProfile } from "@/lib/hooks/profile";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useLocalStorage, useAnalytics } from "@/hooks";
import { logOut } from "@/components/shared/sidebar/actions";
import { clearAllStorage } from "./utils";

export const ProfileMenu = () => {
  const { data: myInvitations = [] } = useMyInvitations();
  const { data: profile } = useProfile();
  const { workspace } = useCurrentWorkspace();
  const { analytics } = useAnalytics();
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");

  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
      clearAllStorage();
      window.location.href = "/login?signedOut=true";
    } finally {
      clearAllStorage();
      window.location.href = "/login?signedOut=true";
    }
  };

  return (
    <Box className="border-border border-t-[0.5px] px-3 pt-4">
      <Menu>
        <Menu.Button>
          <Box className="relative">
            <Button
              className="justify-between px-2"
              color="tertiary"
              fullWidth
              variant="naked"
            >
              <Flex align="center" className="gap-2">
                <Avatar
                  className="relative h-7 text-sm"
                  name={profile?.fullName || profile?.username}
                  src={profile?.avatarUrl}
                  style={{
                    backgroundColor: workspace?.color,
                  }}
                />
                <Text className="line-clamp-1 text-left">
                  {profile?.fullName || profile?.username}{" "}
                </Text>
              </Flex>
              <ArrowRight2Icon className="shrink-0" />
            </Button>
            {myInvitations.length > 0 && (
              <Box className="bg-primary absolute top-0.5 right-1 z-2 size-2.5 animate-pulse rounded-full" />
            )}
          </Box>
        </Menu.Button>
        <Menu.Items align="end" className="ml-3 pt-0">
          <Menu.Group className="px-4 pt-2.5 pb-2">
            <Text className="line-clamp-1" color="muted">
              {profile?.email}
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
                prefetch
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
                    Appearance
                  </span>
                  <span className="flex items-center gap-1">
                    <Text
                      className="hidden text-[0.95rem] first-letter:uppercase md:block"
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
                  prefetch
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
              <LogoutIcon className="text-danger h-5 w-auto" />
              Log out
            </Menu.Item>
          </Menu.Group>
        </Menu.Items>
      </Menu>
    </Box>
  );
};
