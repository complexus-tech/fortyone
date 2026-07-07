"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import {
  ArrowRightIcon,
  ArrowRight2Icon,
  DashboardIcon,
  HistoryIcon,
  MoonIcon,
  NewTabIcon,
  SunIcon,
  SystemIcon,
  UserIcon,
  WorkspaceIcon,
} from "icons";
import { Avatar, Badge, Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import type { Session } from "@/auth";
import { FortyOneLogo } from "@/components/fortyone-logo";
import { getProjectsUrl } from "@/lib/env";

type NavItem = {
  href: string;
  label: string;
  icon: ReactNode;
};

const getThemeLabel = (theme?: string) => {
  if (theme === "light") {
    return "Day mode";
  }

  if (theme === "dark") {
    return "Night mode";
  }

  return "Sync with system";
};

const getThemeIcon = (theme?: string) => {
  if (theme === "light") {
    return <SunIcon className="h-[1.15rem]" />;
  }

  if (theme === "dark") {
    return <MoonIcon className="h-[1.15rem]" />;
  }

  return <SystemIcon className="h-[1.15rem]" />;
};

const navItems: NavItem[] = [
  {
    href: "/overview",
    label: "Overview",
    icon: <DashboardIcon />,
  },
  {
    href: "/workspaces",
    label: "Workspaces",
    icon: <WorkspaceIcon />,
  },
  {
    href: "/users",
    label: "Users",
    icon: <UserIcon />,
  },
  {
    href: "/audit",
    label: "Audit log",
    icon: <HistoryIcon />,
  },
];

export const AdminShell = ({
  children,
  session,
}: {
  children: ReactNode;
  session: Session;
}) => {
  const pathname = usePathname();
  const projectsUrl = getProjectsUrl();

  return (
    <Flex className="h-dvh w-screen overflow-hidden">
      <Box className="from-sidebar to-sidebar/50 border-border/80 relative hidden h-dvh w-(--sidebar-width) shrink-0 flex-col justify-between border-r-[0.5px] bg-linear-to-br pb-6 md:flex">
        <Box className="relative z-1 px-4">
          <Box className="border-border/70 flex h-16 items-center border-b-[0.5px]">
            <Link className="flex items-center gap-3" href="/overview">
              <FortyOneLogo className="h-6" />
              <Text className="text-[0.95rem]" color="muted">
                Admin
              </Text>
            </Link>
          </Box>

          <Box className="mt-5">
            <Text
              className="mb-2 px-2 text-[0.82rem] tracking-[0.08em]"
              color="muted"
              transform="uppercase"
            >
              Platform
            </Text>
            <nav className="space-y-1">
              {navItems.map((item) => {
                const active =
                  pathname === item.href ||
                  pathname.startsWith(`${item.href}/`);

                return (
                  <Link
                    className={cn(
                      "group text-foreground/60 hover:bg-accent flex items-center gap-2 rounded-lg px-2 py-[0.4rem] transition-colors duration-200 outline-none",
                      active ? "bg-accent text-foreground" : "text-text-muted",
                    )}
                    href={item.href}
                    key={item.href}
                  >
                    <span className="shrink-0">{item.icon}</span>
                    <span className="line-clamp-1">{item.label}</span>
                  </Link>
                );
              })}
            </nav>
          </Box>
        </Box>

        <AdminProfileMenu projectsUrl={projectsUrl} session={session} />
      </Box>

      <Box className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <Box className="border-border/80 flex h-14 shrink-0 items-center justify-between border-b-[0.5px] px-4 md:hidden">
          <Link className="flex items-center gap-2" href="/overview">
            <FortyOneLogo asIcon className="h-8" />
            <Text fontWeight="semibold">Admin</Text>
          </Link>
          <Button
            asIcon
            color="tertiary"
            href={projectsUrl}
            prefetch={false}
            size="sm"
            target="_blank"
            variant="naked"
          >
            <ArrowRightIcon />
          </Button>
        </Box>

        <main className="min-h-0 flex-1 overflow-y-auto">{children}</main>
      </Box>
    </Flex>
  );
};

const AdminProfileMenu = ({
  projectsUrl,
  session,
}: {
  projectsUrl: string;
  session: Session;
}) => {
  const { setTheme, theme } = useTheme();
  const displayName = session.user.fullName || session.user.username;

  return (
    <Box className="relative z-1 px-3">
      <Box className="border-border border-t-[0.5px] pt-4">
        <Menu>
          <Menu.Button>
            <Button
              align="between"
              className="px-2"
              color="tertiary"
              fullWidth
              variant="naked"
            >
              <Flex align="center" className="min-w-0 gap-2">
                <Avatar
                  className="relative h-7 text-sm"
                  name={displayName}
                  src={session.user.image}
                />
                <Box className="min-w-0 text-left">
                  <Text className="line-clamp-1" fontWeight="semibold">
                    {displayName}
                  </Text>
                  <Text className="line-clamp-1 text-[0.95rem]" color="muted">
                    {session.user.email}
                  </Text>
                </Box>
              </Flex>
              <ArrowRight2Icon className="shrink-0" />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="ml-3 w-64 pt-0">
            <Menu.Group className="px-4 pt-2.5 pb-2">
              <Flex align="center" className="gap-2">
                <Avatar
                  className="h-8"
                  name={displayName}
                  src={session.user.image}
                />
                <Box className="min-w-0">
                  <Text className="line-clamp-1" fontWeight="semibold">
                    {displayName}
                  </Text>
                  <Text className="line-clamp-1 text-[0.95rem]" color="muted">
                    {session.user.email}
                  </Text>
                </Box>
              </Flex>
              <Badge className="mt-3" color="tertiary">
                Internal
              </Badge>
            </Menu.Group>
            <Menu.Separator className="mb-2" />
            <Menu.Group>
              <Menu.SubMenu>
                <Menu.SubTrigger>
                  <span className="flex w-full items-center justify-between gap-4">
                    <span className="flex items-center gap-2">
                      {getThemeIcon(theme)}
                      Appearance
                    </span>
                    <span className="flex items-center gap-1">
                      <Text
                        className="hidden text-[0.95rem] md:block"
                        color="muted"
                      >
                        {getThemeLabel(theme)}
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
              <Menu.Item>
                <a
                  className="flex w-full items-center gap-2"
                  href={projectsUrl}
                  rel="noreferrer"
                  target="_blank"
                >
                  <NewTabIcon className="h-[1.15rem]" />
                  Projects
                </a>
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Box>
    </Box>
  );
};
