"use client";

import { Box, Button, Flex, Menu, NavLink, NavigationMenu, Text } from "ui";
import { SprintsIcon, ObjectiveIcon, StoryIcon, OKRIcon } from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import type { ReactNode } from "react";
import { useState } from "react";
import { useSession } from "next-auth/react";
import { Logo, Container } from "@/components/ui";
import type { Workspace } from "@/types";
import { useWorkspaces } from "@/lib/hooks/workspaces";
import { useProfile } from "@/lib/hooks/profile";
import { MenuButton } from "./menu-button";
import { RequestDemo } from "./request-demo";

const MenuItem = ({
  name,
  description,
  icon,
  href,
}: {
  name: string;
  description: string;
  icon: ReactNode;
  href: string;
}) => (
  <Link
    className="flex w-[17rem] gap-2 rounded-[0.7rem] p-2 hover:bg-gray-50/80 hover:dark:bg-dark-100/50"
    href={href}
    prefetch
  >
    {icon}
    <Box>
      <Text className="dark:text-white">{name}</Text>
      <Text className="text-[0.85rem] opacity-70">{description}</Text>
    </Box>
  </Link>
);

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const Navigation = () => {
  const { data: session } = useSession();
  const navLinks = [
    { title: "Pricing", href: "/pricing" },
    { title: "Contact", href: "/contact" },
    { title: "Blog", href: "/blog" },
    { title: "Help Center", href: "https://docs.complexus.app" },
  ];

  const features = [
    {
      id: 1,
      name: "Stories",
      href: "/features/stories",
      description: "Turn ideas into clear, actionable work.",
      icon: (
        <StoryIcon className="relative h-6 shrink-0 text-dark dark:text-white md:top-1 md:h-4" />
      ),
    },
    {
      id: 2,
      name: "Objectives",
      href: "/features/objectives",
      description: "Focus teams on the outcomes that matter most.",
      icon: (
        <ObjectiveIcon className="relative h-6 shrink-0 text-dark dark:text-white md:top-1 md:h-4" />
      ),
    },
    {
      id: 3,
      name: "OKRs",
      href: "/features/okrs",
      description: "Align everyone on shared goals and see progress.",
      icon: (
        <OKRIcon className="relative h-6 shrink-0 text-dark dark:text-white md:top-1 md:h-4" />
      ),
    },
    {
      id: 4,
      name: "Sprints",
      href: "/features/sprints",
      description:
        "Build momentum with short cycles that turn plans into progress.",
      icon: (
        <SprintsIcon className="relative h-6 shrink-0 text-dark dark:text-white md:top-1 md:h-4" />
      ),
    },
  ];

  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const pathname = usePathname();
  const { data: workspaces = [] } = useWorkspaces();
  const { data: profile } = useProfile();

  const getNextUrl = () => {
    let workspace: Workspace | undefined;
    if (session) {
      if (workspaces.length === 0) {
        return "/onboarding/create";
      }
      if (profile?.lastUsedWorkspaceId) {
        workspace = workspaces.find(
          (w) => w.id === profile.lastUsedWorkspaceId,
        );
      } else {
        workspace = workspaces[0];
      }
      if (domain.includes("localhost")) {
        return `https://${workspace!.slug}.complexus.lc/my-work`;
      }
      return `https://${workspace!.slug}.${domain}/my-work`;
    }
    return "/login";
  };

  return (
    <Box className="fixed left-0 z-[15] w-screen border-b border-gray-100/70 bg-white/20 backdrop-blur-xl dark:border-dark-100/80 dark:bg-black/40">
      <Container className="flex h-16 items-center justify-between gap-12 font-medium">
        <Logo className="relative top-0.5 h-6 pl-1 text-dark dark:text-gray-50 md:h-[1.65rem] md:pl-0" />
        <Flex align="center" className="hidden md:flex" gap={1}>
          <NavigationMenu>
            <NavigationMenu.List>
              <NavigationMenu.Item>
                <NavigationMenu.Trigger
                  className={cn(
                    "px-3 py-1.5 opacity-90 transition hover:opacity-100",
                    {
                      "opacity-100 dark:text-white dark:opacity-100":
                        pathname?.startsWith("/features"),
                    },
                  )}
                  hideArrow
                >
                  Product
                </NavigationMenu.Trigger>
                <NavigationMenu.Content className="rounded-2xl border border-gray-100 bg-white p-1.5 dark:border-dark-50 dark:bg-black">
                  <Box className="grid w-max grid-cols-2 gap-2 rounded-xl border border-gray-100 p-2 dark:border-dark-50 dark:bg-dark/20">
                    {features.map(({ id, name, description, icon, href }) => (
                      <MenuItem
                        description={description}
                        href={href}
                        icon={icon}
                        key={id}
                        name={name}
                      />
                    ))}
                  </Box>
                </NavigationMenu.Content>
              </NavigationMenu.Item>
            </NavigationMenu.List>
          </NavigationMenu>
          {navLinks.map(({ title, href }) => (
            <NavLink
              className={cn(
                "flex items-center rounded-full px-3 opacity-90 transition hover:opacity-100",
                {
                  "opacity-100 dark:text-white dark:opacity-100":
                    pathname === href,
                },
              )}
              href={href}
              key={title}
              prefetch
            >
              {title}
            </NavLink>
          ))}
        </Flex>
        <Flex align="center" className="ml-4 gap-2">
          <RequestDemo />
          <Button
            className={cn("hidden px-5 text-[0.93rem] md:flex", {
              flex: session,
            })}
            color={session ? "invert" : "tertiary"}
            href={getNextUrl()}
            rounded="lg"
            variant={session ? "solid" : "naked"}
          >
            {session ? (
              <>
                {getNextUrl().includes("onboarding") ? (
                  "Create workspace"
                ) : (
                  <>
                    <span className="md:hidden">Open app</span>
                    <span className="hidden md:block">Open workspace</span>
                  </>
                )}
              </>
            ) : (
              "Login"
            )}
          </Button>
          {!session && (
            <Button
              className="px-5 text-[0.93rem]"
              color="invert"
              href="/signup"
              rounded="lg"
            >
              Sign up
            </Button>
          )}

          <Box className="flex md:hidden">
            <Menu onOpenChange={setIsMenuOpen} open={isMenuOpen}>
              <Menu.Button asChild>
                <button type="button">
                  <MenuButton open={isMenuOpen} />
                </button>
              </Menu.Button>
              <Menu.Items
                align="end"
                className="relative left-3.5 mt-4 w-[calc(100vw-2.5rem)] rounded-2xl py-4"
              >
                <Menu.Group className="px-4 py-2.5">
                  <Text color="muted">Product</Text>
                </Menu.Group>
                <Menu.Group>
                  {navLinks.map(({ title, href }) => (
                    <Menu.Item
                      className="block rounded-xl py-2.5"
                      key={title}
                      onClick={() => {
                        setIsMenuOpen(false);
                      }}
                    >
                      <NavLink className="flex text-xl" href={href} prefetch>
                        {title}
                      </NavLink>
                    </Menu.Item>
                  ))}
                </Menu.Group>
                <Menu.Separator />
                <Menu.Group className="px-4 py-2.5">
                  <Text color="muted">Features</Text>
                </Menu.Group>
                <Menu.Group>
                  {features.map(({ id, name, href }) => (
                    <Menu.Item
                      className="block rounded-xl py-2.5"
                      key={id}
                      onClick={() => {
                        setIsMenuOpen(false);
                      }}
                    >
                      <NavLink
                        className="flex items-center gap-2 text-xl"
                        href={href}
                      >
                        {name}
                      </NavLink>
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};
