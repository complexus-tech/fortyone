"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ArrowRightIcon,
  DashboardIcon,
  HistoryIcon,
  NewTabIcon,
  UserIcon,
  WorkspaceIcon,
} from "icons";
import { Avatar, Badge, Box, Button, Flex, Text } from "ui";
import { cn } from "lib";
import type { Session } from "@/auth";
import { getProjectsUrl } from "@/lib/env";

type NavItem = {
  href: string;
  label: string;
  icon: ReactNode;
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
            <Link className="flex items-center gap-2.5" href="/overview">
              <span className="bg-primary text-primary-foreground font-heading flex size-8 items-center justify-center rounded-xl text-[1.05rem] font-semibold">
                41
              </span>
              <span>
                <Text className="leading-none" fontWeight="semibold">
                  FortyOne
                </Text>
                <Text className="mt-0.5 text-[0.92rem]" color="muted">
                  Admin
                </Text>
              </span>
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
                      "flex h-9 items-center gap-2 rounded-xl px-2.5 transition duration-200 ease-linear",
                      active
                        ? "bg-state-selected text-foreground"
                        : "text-text-muted hover:bg-state-hover hover:text-foreground",
                    )}
                    href={item.href}
                    key={item.href}
                  >
                    {item.icon}
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>
          </Box>
        </Box>

        <Box className="relative z-1 px-3">
          <Box className="border-border bg-surface/80 shadow-shadow mb-3 rounded-xl border-[0.5px] p-3 shadow-lg">
            <Flex align="center" className="gap-2">
              <Avatar
                className="h-8"
                name={session.user.fullName || session.user.username}
                src={session.user.image}
              />
              <Box className="min-w-0">
                <Text className="line-clamp-1" fontWeight="semibold">
                  {session.user.fullName || session.user.username}
                </Text>
                <Text className="line-clamp-1 text-[0.92rem]" color="muted">
                  {session.user.email}
                </Text>
              </Box>
            </Flex>
            <Flex align="center" className="mt-3" justify="between">
              <Badge rounded="full" size="sm">
                Internal
              </Badge>
              <Button
                className="px-2"
                color="tertiary"
                href={projectsUrl}
                prefetch={false}
                size="xs"
                target="_blank"
                variant="naked"
              >
                Projects
                <NewTabIcon className="h-4" />
              </Button>
            </Flex>
          </Box>
        </Box>
      </Box>

      <Box className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <Box className="border-border/80 flex h-14 shrink-0 items-center justify-between border-b-[0.5px] px-4 md:hidden">
          <Link className="flex items-center gap-2" href="/overview">
            <span className="bg-primary text-primary-foreground font-heading flex size-8 items-center justify-center rounded-xl font-semibold">
              41
            </span>
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
