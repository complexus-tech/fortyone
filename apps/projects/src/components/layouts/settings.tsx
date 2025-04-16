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
import { BodyContainer, MobileMenuButton } from "../shared";
import { NavLink } from "../ui";

export const SettingsLayout = ({ children }: { children: ReactNode }) => {
  const { userRole } = useUserRole();
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
    { title: "Delete account", href: "/settings/account/delete" },
    ...(myInvitations.length > 0
      ? [{ title: "Invitations", href: "/settings/invitations" }]
      : []),
  ];

  const workspaceItems = [
    ...(isAdmin
      ? [
          { title: "General", href: "/settings" },
          { title: "Members", href: "/settings/workspace/members" },
          { title: "Terminology", href: "/settings/workspace/terminology" },
          { title: "API tokens", href: "/settings/workspace/api" },
        ]
      : []),
  ];

  const featureItems = [
    ...(isAdmin || isMember
      ? [
          {
            title: getTermDisplay("objectiveTerm", {
              variant: "plural",
              capitalize: true,
            }),
            href: "/settings/workspace/objectives",
          },
          { title: "Labels", href: "/settings/workspace/labels" },
          { title: "Teams", href: "/settings/workspace/teams" },
          { title: "Create a team", href: "/settings/workspace/teams/create" },
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
    ...(isAdmin || isMember
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
              <ArrowLeft2Icon className="h-[1.1rem] w-auto opacity-50 transition group-hover:opacity-100" />
              Settings
            </button>
          </Flex>
        </Container>
        <Box className="overflow-x-auto border-y border-gray-100 pl-3 dark:border-dark-100">
          <Flex align="center" gap={2}>
            {mobileMenu.map(({ href, title }) => (
              <Link
                className={cn(
                  "h-16 shrink-0 border-b border-transparent px-3 leading-[4rem]",
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
        <Box className="h-[calc(100dvh-8rem)] overflow-y-auto pb-8 pt-6">
          <Container>{children}</Container>
        </Box>
      </Box>
      <Box className="hidden md:block">
        <ResizablePanel autoSaveId="settings:layout" direction="horizontal">
          <ResizablePanel.Panel
            className="bg-gray-50/60 dark:bg-black"
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
                  className="group flex items-center gap-3 text-lg font-medium"
                  onClick={goBack}
                  type="button"
                >
                  <ArrowLeft2Icon className="h-[1.1rem] w-auto opacity-50 transition group-hover:opacity-100" />
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
                          className="py-1.5"
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
            <Box className="h-screen overflow-y-auto">
              <Container className="max-w-[54rem] py-12">{children}</Container>
            </Box>
          </ResizablePanel.Panel>
        </ResizablePanel>
      </Box>
    </>
  );
};
