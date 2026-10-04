"use client";
import {
  ArrowLeft2Icon,
  SettingsIcon,
  UserIcon,
  WorkflowIcon,
  WorkspaceIcon,
} from "icons";
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
import { DEFAULT_WORKSPACE_PATH } from "@/shared/routing/workspace";
import { Commands } from "@/shell/commands/commands";
import { MobileMenuButton } from "../shared/mobile-menu";
import { NavLink } from "../ui";
import {
  buildSettingsNavigation,
  isSettingsItemActive,
  type SettingsNavigationCategory,
} from "./settings-navigation";

const categoryIcons: Record<SettingsNavigationCategory["category"], ReactNode> =
  {
    Account: <UserIcon className="h-[1.15rem]" />,
    Workspace: <WorkspaceIcon />,
    Administration: <SettingsIcon />,
    Features: <WorkflowIcon />,
  };

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
    router.push(prevPage || withWorkspace(DEFAULT_WORKSPACE_PATH));
    setPrevPage("");
  };

  useHotkeys("esc", (event) => {
    if (event.defaultPrevented) return;
    goBack();
  });

  const navigation = buildSettingsNavigation({
    userRole,
    hasCustomTerminology: hasFeature("customTerminology"),
    hasInvitations: myInvitations.length > 0,
    objectiveTitle: getTermDisplay("objectiveTerm", {
      variant: "plural",
      capitalize: true,
    }),
    withWorkspace,
  });

  const mobileMenu = navigation.flatMap(({ items }) => items);

  return (
    <>
      <Box className="flex h-dvh flex-col md:flex-row" data-settings-shell>
        <Box className="shrink-0 md:hidden">
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
              {mobileMenu.map((item) => {
                const { href, title } = item;
                const isActive = isSettingsItemActive(pathname, item);

                return (
                  <Link
                    className={cn(
                      "h-16 shrink-0 border-b border-transparent px-3 leading-16",
                      {
                        "border-primary text-primary": isActive,
                      },
                    )}
                    href={href}
                    key={href}
                    prefetch
                  >
                    {title}
                  </Link>
                );
              })}
            </Flex>
          </Box>
        </Box>
        <Box className="hidden w-(--sidebar-width) shrink-0 flex-col md:flex">
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
              {navigation.map(({ category, items }) => {
                const isCategoryActive = items.some((item) =>
                  isSettingsItemActive(pathname, item),
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
                      <span className="shrink-0">
                        {categoryIcons[category]}
                      </span>
                      <Text>{category}</Text>
                    </Flex>
                    <Flex className="ml-8" direction="column" gap={1}>
                      {items.map((item) => {
                        const { href, title } = item;
                        const isActive = isSettingsItemActive(pathname, item);

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
        <Box className="min-h-0 min-w-0 flex-1 md:pt-(--app-content-inset) md:pl-2">
          <Box
            className="app-content-canvas-gradient settings-card-borders md:border-border/80 md:bg-surface-muted/60 dark:md:bg-surface-muted/40 h-full min-w-0 overflow-y-auto max-md:bg-none md:rounded-tl-2xl md:border-t-[0.5px] md:border-l-[0.5px]"
            data-settings-content-canvas
          >
            <Container
              className={cn("pt-6 pb-8 md:max-w-216 md:pt-16 md:pb-12", {
                "md:max-w-[80rem]":
                  pathname.includes("billing") ||
                  pathname.endsWith("/settings/workspace/security"),
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
