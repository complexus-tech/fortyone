import type { ComponentPropsWithoutRef, ReactNode } from "react";
import Link from "next/link";
import { RequestsIcon, RoadmapIcon, UpdatesIcon } from "icons";
import { Avatar, Box, Button, Flex, Text } from "ui";
import { cn, getReadableTextColor } from "lib";
import { getLoginUrl } from "@/utils/callback-url";
import type {
  PublicPortal,
  PublicPortalTab,
  PublicPortalViewer,
} from "./types";
import { PublicPortalUserMenu } from "./user-menu";
import { PublicPortalNotifications } from "./notifications-popover";
import { getFeedbackSignupPath } from "./feedback-setup";
import { getPortalCallbackUrl, getPortalPath } from "./utils";

const navItems = [
  { icon: RequestsIcon, label: "Feedback", tab: "feedback" },
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
  loginCallbackUrl,
  portal,
  viewer,
}: {
  activeTab?: PublicPortalTab;
  children: ReactNode;
  loginCallbackUrl?: string;
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => (
  <Box className="bg-background flex h-dvh flex-col overflow-hidden">
    <Box className="border-border/60 bg-background sticky top-0 z-20 shrink-0 border-b">
      <Box className="relative mx-auto flex h-16 w-full max-w-[78rem] items-center justify-between px-4 md:px-6">
        <Link
          aria-label={`${portal.workspace.name} feedback`}
          className="min-w-0 flex-1 transition-opacity hover:opacity-80"
          href={getPortalPath(portal, "feedback")}
        >
          <Flex align="center" gap={3}>
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
        </Link>
        <Box className="absolute left-1/2 hidden -translate-x-1/2 md:block">
          <nav className="bg-surface-muted/85 ml-2 hidden rounded-xl p-1 md:flex">
            {navItems.map((item) => {
              const isActive = item.tab === activeTab;
              const Icon = item.icon;
              return (
                <Link
                  className={cn(
                    "text-text-muted hover:text-foreground flex items-center gap-2 rounded-xl border border-transparent px-3.5 py-1.5 text-[0.95rem] transition",
                    {
                      "border-border bg-surface-elevated text-foreground":
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
          {viewer ? (
            <>
              <PublicPortalNotifications portal={portal} />
              <PublicPortalUserMenu viewer={viewer} />
            </>
          ) : (
            <Button
              className="h-10 px-4"
              color="invert"
              href={getLoginUrl(
                loginCallbackUrl ??
                  getPortalCallbackUrl(portal, activeTab ?? "feedback"),
              )}
              size="md"
            >
              Login/signup
            </Button>
          )}
        </Flex>
      </Box>
      <nav className="border-border/60 bg-background flex border-t p-1.5 md:hidden">
        {navItems.map((item) => {
          const isActive = item.tab === activeTab;
          const Icon = item.icon;
          return (
            <Link
              className={cn(
                "text-text-muted flex flex-1 items-center justify-center gap-2 rounded-xl border border-transparent py-2.5 text-center text-[0.95rem]",
                {
                  "border-border bg-surface-elevated text-foreground": isActive,
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
    <Box className="bg-background min-h-0 flex-1 overflow-y-auto">
      {children}
    </Box>
    <Button
      className="bg-surface-elevated/90 shadow-shadow fixed right-4 bottom-4 z-30 h-10 border-[0.5px] px-3 shadow-lg backdrop-blur md:right-6 md:bottom-6"
      color="tertiary"
      href={viewer?.feedbackSetupHref ?? getFeedbackSignupPath()}
      size="sm"
      variant="outline"
    >
      Create your own board
    </Button>
  </Box>
);
