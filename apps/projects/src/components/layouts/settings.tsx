"use client";
import { ArrowLeft2Icon, UserIcon, WorkflowIcon, WorkspaceIcon } from "icons";
import type { ReactNode } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Badge, Box, Container, Flex, Text, Tooltip } from "ui";
import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { cn } from "lib";
import {
  useLocalStorage,
  useUserRole,
  useTerminology,
  useWorkspacePath,
} from "@/hooks";
import { useMyInvitations } from "@/modules/invitations/hooks/my-invitations";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { Commands } from "@/shell/commands/commands";
import { MobileMenuButton } from "../shared/mobile-menu";
import { NavLink } from "../ui";

export const SettingsLayout = ({ children }: { children: ReactNode }) => {
  const { userRole } = useUserRole();
  const { hasFeature } = useSubscriptionFeatures();
  const [prevPage, setPrevPage] = useLocalStorage("pathBeforeSettings", "");
  const router = useRouter();
  const pathname = usePathname();
  const { data: myInvitations = [] } = useMyInvitations();
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();

  const goBack = () => {
    router.push(prevPage || withWorkspace("/my-work"));
    setPrevPage("");
  };

  useHotkeys("esc", () => {
    goBack();
  });

  const isAdmin = userRole === "admin";
  const isMember = userRole === "member";
  const canUseIntegrations = isAdmin || isMember || userRole === "guest";

  const accountItems = [
    { title: "Profile", href: withWorkspace("/settings/account") },
    { title: "Calendar", href: withWorkspace("/settings/account/calendar") },
    {
      title: "Preferences",
      href: withWorkspace("/settings/account/preferences"),
    },
    {
      title: "Notifications",
      href: withWorkspace("/settings/account/notifications"),
    },
    { title: "Security", href: withWorkspace("/settings/account/security") },
    ...(myInvitations.length > 0
      ? [{ title: "Invitations", href: withWorkspace("/settings/invitations") }]
      : []),
  ];

  const workspaceItems = [
    ...(isAdmin
      ? [
          { title: "General", href: withWorkspace("/settings") },
          {
            title: "Members",
            href: withWorkspace("/settings/workspace/members"),
          },
          {
            title: "Billing & plans",
            href: withWorkspace("/settings/workspace/billing"),
          },
          ...(hasFeature("customTerminology")
            ? [
                {
                  title: "Terminology",
                  href: withWorkspace("/settings/workspace/terminology"),
                },
              ]
            : []),
        ]
      : []),
    {
      title: "API",
      href: withWorkspace("/settings/workspace/api"),
    },
    ...(canUseIntegrations
      ? [
          {
            title: "Integrations",
            href: withWorkspace("/settings/integrations"),
          },
        ]
      : []),
  ];

  const featureItems = [
    ...(isAdmin || isMember
      ? [
          {
            title: "Labels",
            href: withWorkspace("/settings/workspace/labels"),
          },
          {
            title: getTermDisplay("objectiveTerm", {
              variant: "plural",
              capitalize: true,
            }),
            href: withWorkspace("/settings/workspace/objectives"),
          },
          { title: "Teams", href: withWorkspace("/settings/workspace/teams") },
          {
            title: "Feedback",
            href: withWorkspace("/settings/workspace/feedback"),
          },
        ]
      : []),
  ];

  const navigation = [
    {
      category: "Account",
      icon: <UserIcon className="h-[1.15rem]" />,
      items: accountItems,
    },
    ...(workspaceItems.length > 0
      ? [
          {
            category: "Workspace",
            icon: <WorkspaceIcon />,
            items: workspaceItems,
          },
        ]
      : []),
    ...(isAdmin
      ? [
          {
            category: "Features",
            icon: <WorkflowIcon />,
            items: featureItems,
          },
        ]
      : []),
  ];

  const isSettingsItemActive = (href: string, title: string) =>
    pathname === href ||
    (title === "Integrations" && pathname.startsWith(`${href}/`));

  const mobileMenu = navigation.flatMap(({ items }) =>
    items.map(({ href, title }) => ({ href, title })),
  );

  return (
    <>
      <Box className="md:hidden">
        <Container>
          <Flex align="center" className="h-16" gap={2}>
            <MobileMenuButton />
            <button
              className="group flex items-center gap-1 font-medium"
              onClick={goBack}
              type="button"
            >
              <ArrowLeft2Icon strokeWidth={3} />
              Settings
            </button>
          </Flex>
        </Container>
        <Box className="border-border overflow-x-auto border-y-[0.5px] pl-3">
          <Flex align="center" gap={2}>
            {mobileMenu.map(({ href, title }) => (
              <Link
                className={cn(
                  "h-16 shrink-0 border-b border-transparent px-3 leading-16",
                  {
                    "border-primary text-primary": pathname === href,
                  },
                )}
                href={href}
                key={href}
                prefetch
              >
                {title}
              </Link>
            ))}
          </Flex>
        </Box>
        <Box className="settings-card-borders h-[calc(100dvh-8rem)] overflow-y-auto pt-6 pb-8">
          <Container>{children}</Container>
        </Box>
      </Box>
      <Box className="hidden h-dvh md:flex" data-settings-shell>
        <Box className="flex w-(--sidebar-width) shrink-0 flex-col">
          <Box className="flex h-(--app-shell-header-height) shrink-0 items-center px-4">
            <Tooltip
              title={
                <span className="flex items-center gap-1">
                  Close Settings
                  <Badge color="tertiary" rounded="sm" size="sm">
                    Esc
                  </Badge>
                </span>
              }
            >
              <button
                className="group flex items-center gap-1.5 text-lg font-medium"
                onClick={goBack}
                type="button"
              >
                <ArrowLeft2Icon strokeWidth={2.8} />
                Settings
              </button>
            </Tooltip>
          </Box>
          <Box className="min-h-0 flex-1 overflow-y-auto px-4 pb-4">
            <Flex className="mt-6" direction="column" gap={4}>
              {navigation.map(({ category, items, icon }) => {
                const isCategoryActive = items.some(({ href, title }) =>
                  isSettingsItemActive(href, title),
                );

                return (
                  <Box className="mb-3" key={category}>
                    <Flex
                      align="center"
                      className={cn(
                        "text-text-muted mb-2 transition-colors [&_svg]:text-current",
                        isCategoryActive && "text-primary",
                      )}
                      gap={4}
                    >
                      <span className="shrink-0">{icon}</span>
                      <Text>{category}</Text>
                    </Flex>
                    <Flex className="ml-8" direction="column" gap={1}>
                      {items.map(({ href, title }) => {
                        const isActive = isSettingsItemActive(href, title);

                        return (
                          <NavLink
                            active={isActive}
                            aria-current={isActive ? "page" : undefined}
                            className={cn(
                              "hover:bg-primary/5 hover:text-primary relative -left-1 py-1.5",
                              isActive && "bg-primary/5 text-primary",
                            )}
                            href={href}
                            key={href}
                          >
                            {title}
                          </NavLink>
                        );
                      })}
                    </Flex>
                  </Box>
                );
              })}
            </Flex>
          </Box>
        </Box>
        <Box className="h-dvh min-w-0 flex-1 pt-(--app-content-inset) pr-(--app-content-inset) pb-(--app-content-inset) pl-2">
          <Box
            className="border-border/80 bg-surface-muted/40 dark:bg-surface-muted/20 settings-card-borders h-full min-w-0 overflow-y-auto rounded-xl border-[0.5px]"
            data-settings-content-canvas
          >
            <Container
              className={cn("max-w-216 pt-16 pb-12", {
                "max-w-[80rem]": pathname.includes("billing"),
              })}
            >
              {children}
            </Container>
          </Box>
        </Box>
      </Box>
      <Commands />
    </>
  );
};
