import type { ComponentPropsWithoutRef, ReactNode } from "react";
import Link from "next/link";
import {
  BellIcon,
  RequestsIcon,
  RoadmapIcon,
  SearchIcon,
  UpdatesIcon,
} from "icons";
import { Avatar, Box, Button, Flex, Text } from "ui";
import { cn, getReadableTextColor } from "lib";
import type { PublicPortal, PublicPortalTab, PublicPortalViewer } from "./types";
import { PublicPortalUserMenu } from "./user-menu";
import { getPortalPath } from "./utils";

const navItems = [
  { icon: RequestsIcon, label: "Requests", tab: "requests" },
  { icon: RoadmapIcon, label: "Roadmap", tab: "roadmap" },
  { icon: UpdatesIcon, label: "Updates", tab: "updates" },
] satisfies {
  icon: (props: ComponentPropsWithoutRef<"svg">) => ReactNode;
  label: string;
  tab: PublicPortalTab;
}[];

export const PublicPortalShell = ({
  activeTab,
  children,
  portal,
  viewer,
}: {
  activeTab: PublicPortalTab;
  children: ReactNode;
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => (
  <Box className="bg-background min-h-dvh overflow-y-auto">
    <Box className="border-border/60 bg-background/95 sticky top-0 z-20 border-b-[0.5px] backdrop-blur">
      <Box className="relative mx-auto flex h-16 w-full max-w-[78rem] items-center justify-between px-4 md:px-6">
        <Flex align="center" className="min-w-0 flex-1" gap={3}>
          <Avatar
            className="!size-9 text-base font-bold shadow-sm"
            name={portal.workspace.name}
            rounded="full"
            size="md"
            src={portal.workspace.avatarUrl}
            style={{
              backgroundColor: portal.workspace.color,
              color: getReadableTextColor(portal.workspace.color),
            }}
          />
          <Text className="line-clamp-1 text-base" fontWeight="semibold">
            {portal.workspace.name}
          </Text>
        </Flex>
        <Box className="absolute left-1/2 hidden -translate-x-1/2 md:block">
          <nav className="bg-surface border-border/70 shadow-shadow/30 ml-2 hidden rounded-full border-[0.5px] p-1 shadow-sm md:flex">
            {navItems.map((item) => {
              const isActive = item.tab === activeTab;
              const Icon = item.icon;
              return (
                <Link
                  className={cn(
                    "text-text-muted hover:text-foreground flex items-center gap-2 rounded-full px-3.5 py-1.5 text-[0.95rem] transition",
                    {
                      "border-border bg-state-selected/50 text-foreground shadow-xs dark:bg-state-selected border-[0.5px]":
                        isActive,
                    },
                  )}
                  href={getPortalPath(portal, item.tab)}
                  key={item.tab}
                >
                  <Icon className="h-4 text-current" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </Box>
        <Flex align="center" className="min-w-0 flex-1 justify-end" gap={2}>
          <Box className="border-border bg-surface text-text-muted hidden h-10 w-56 items-center gap-2 rounded-full border-[0.5px] px-3.5 md:flex">
            <SearchIcon className="h-4 shrink-0" />
            <span>Search...</span>
            <span className="border-border ml-auto rounded border px-1.5 text-[0.78rem]">
              /
            </span>
          </Box>
          {viewer ? (
            <>
              <Button
                asIcon
                className="hidden md:flex"
                color="tertiary"
                href={viewer.notificationsHref}
                rounded="full"
                size="md"
                variant="naked"
              >
                <BellIcon className="h-5" />
                <span className="sr-only">Notifications</span>
              </Button>
              <PublicPortalUserMenu viewer={viewer} />
            </>
          ) : (
            <Button
              className="h-10 px-4"
              color="invert"
              href="/"
              rounded="full"
              size="md"
            >
              Login/signup
            </Button>
          )}
        </Flex>
      </Box>
      <nav className="border-border/60 bg-surface-muted/40 flex border-t-[0.5px] p-1.5 md:hidden">
        {navItems.map((item) => {
          const isActive = item.tab === activeTab;
          const Icon = item.icon;
          return (
            <Link
              className={cn(
                "text-text-muted flex flex-1 items-center justify-center gap-2 rounded-full py-2.5 text-center text-[0.95rem]",
                {
                  "border-border bg-state-selected/50 text-foreground dark:bg-state-selected border-[0.5px]":
                    isActive,
                },
              )}
              href={getPortalPath(portal, item.tab)}
              key={item.tab}
            >
              <Icon className="h-4 text-current" />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </Box>
    {children}
  </Box>
);
