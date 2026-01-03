"use client";
import { ArrowLeft2Icon, UserIcon, WorkflowIcon, WorkspaceIcon } from "icons";
import type { ReactNode } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Badge, Box, Container, Flex, ResizablePanel, Text, Tooltip } from "ui";
import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { cn } from "lib";
import { useLocalStorage, useUserRole, useTerminology } from "@/hooks";
import { useMyInvitations } from "@/modules/invitations/hooks/my-invitations";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { BodyContainer, MobileMenuButton } from "../shared";
import { NavLink } from "../ui";
import { Commands } from "../shared/commands";

export const SettingsLayout = ({ children }: { children: ReactNode }) => {
  const { userRole } = useUserRole();
  const { hasFeature } = useSubscriptionFeatures();
  const [prevPage, setPrevPage] = useLocalStorage("pathBeforeSettings", "");
  const router = useRouter();
  const pathname = usePathname();
  const { data: myInvitations = [] } = useMyInvitations();
  const { getTermDisplay } = useTerminology();

  const goBack = () => {
    router.push(prevPage || "/my-work");
    setPrevPage("");
  };

  useHotkeys("esc", () => {
    goBack();
  });

  const isAdmin = userRole === "admin";
  const isMember = userRole === "member";

  const accountItems = [
    { title: "Profile", href: "/settings/account" },
    { title: "Preferences", href: "/settings/account/preferences" },
    { title: "Notifications", href: "/settings/account/notifications" },
    { title: "Security", href: "/settings/account/security" },
    ...(myInvitations.length > 0
      ? [{ title: "Invitations", href: "/settings/invitations" }]
      : []),
  ];

  const workspaceItems = [
    ...(isAdmin
      ? [
          { title: "General", href: "/settings" },
          { title: "Members", href: "/settings/workspace/members" },
          { title: "Billing & plans", href: "/settings/workspace/billing" },
          ...(hasFeature("customTerminology")
            ? [
                {
                  title: "Terminology",
                  href: "/settings/workspace/terminology",
                },
              ]
            : []),
          // { title: "Integrations", href: "/settings/workspace/integrations" },
          // { title: "API tokens", href: "/settings/workspace/api" },
        ]
      : []),
  ];

  const featureItems = [
    ...(isAdmin || isMember
      ? [
          { title: "Labels", href: "/settings/workspace/labels" },
          {
            title: getTermDisplay("objectiveTerm", {
              variant: "plural",
              capitalize: true,
            }),
            href: "/settings/workspace/objectives",
          },
          { title: "Teams", href: "/settings/workspace/teams" },
        ]
      : []),
  ];

  const navigation = [
    {
      category: "Account",
      icon: <UserIcon className="h-[1.15rem]" />,
      items: accountItems,
    },
    ...(isAdmin
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

  const mobileMenu = navigation
    .map(({ items }) => {
      return {
        items: items.map(({ href, title }) => ({ href, title })),
      };
    })
    .flatMap(({ items }) => items);

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
        <Box className="border-border overflow-x-auto border-y pl-3">
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
        <Box className="h-[calc(100dvh-8rem)] overflow-y-auto pt-6 pb-8">
          <Container>{children}</Container>
        </Box>
      </Box>
      <Box className="hidden md:block">
        <ResizablePanel autoSaveId="settings:layout" direction="horizontal">
          <ResizablePanel.Panel
            className="from-sidebar to-sidebar/50 bg-linear-to-br"
            defaultSize={15}
            maxSize={20}
            minSize={12}
          >
            <Box className="flex h-16 items-center px-4">
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
            <BodyContainer className="px-4">
              <Flex className="mt-6" direction="column" gap={4}>
                {navigation.map(({ category, items, icon }) => (
                  <Box className="mb-3" key={category}>
                    <Flex align="center" className="mb-2" gap={4}>
                      {icon}
                      <Text color="muted">{category}</Text>
                    </Flex>
                    <Flex className="ml-8" direction="column" gap={1}>
                      {items.map(({ href, title }) => (
                        <NavLink
                          active={pathname === href}
                          className="relative -left-1 py-1.5"
                          href={href}
                          key={href}
                        >
                          {title}
                        </NavLink>
                      ))}
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </BodyContainer>
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel defaultSize={85}>
            <Box className="h-dvh overflow-y-auto">
              <Container
                className={cn("max-w-216 py-12", {
                  "max-w-[80rem]": pathname.includes("billing"),
                })}
              >
                {children}
              </Container>
            </Box>
          </ResizablePanel.Panel>
        </ResizablePanel>
      </Box>
      <Commands />
    </>
  );
};
