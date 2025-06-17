"use client";

import { Box, Button, Flex, Menu, NavLink, NavigationMenu, Text } from "ui";
import {
  SprintsIcon,
  ObjectiveIcon,
  StoryIcon,
  OKRIcon,
  TwitterIcon,
  LinkedinIcon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import type { ReactNode } from "react";
import { useState } from "react";
import { useSession } from "next-auth/react";
import { Logo, Container } from "@/components/ui";
import type { Workspace } from "@/types";
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
    className="flex w-[17rem] gap-2 rounded-[0.6rem] p-2 hover:bg-gray-100 hover:dark:bg-dark-200"
    href={href}
  >
    {icon}
    <Box>
      <Text>{name}</Text>
      <Text className="text-[0.9rem]" color="muted">
        {description}
      </Text>
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
    { title: "Docs", href: "https://docs.complexus.app" },
  ];

  const product = [
    {
      id: 1,
      name: "Stories",
      href: "/product/stories",
      description: "Manage and Track Tasks",
      icon: (
        <StoryIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 2,
      name: "Objectives",
      href: "/product/objectives",
      description: "Set and Achieve Goals",
      icon: (
        <ObjectiveIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 3,
      name: "OKRs",
      href: "/product/okrs",
      description: "Align and Achieve",
      icon: (
        <OKRIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 4,
      name: "Sprints",
      href: "/product/sprints",
      description: "Iterate and Deliver",
      icon: (
        <SprintsIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
  ];

  const company = [
    {
      id: 1,
      href: "https://x.com/complexus_app",
      name: "X (Formerly Twitter)",
      description: "Follow us on X",
      icon: <TwitterIcon className="relative h-[1.15rem] shrink-0 md:top-1" />,
    },
    {
      id: 2,
      href: "https://linkedin.com/company/complexus-app",
      name: "LinkedIn",
      description: "Follow us on LinkedIn",
      icon: <LinkedinIcon className="relative shrink-0 md:top-1" />,
    },
  ];

  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const pathname = usePathname();

  const getNextUrl = () => {
    if (session) {
      let workspace: Workspace | undefined;
      if (session.activeWorkspace) {
        workspace = session.activeWorkspace;
      } else if (session.workspaces.length > 0) {
        workspace = session.workspaces[0];
      } else {
        return "/onboarding/create";
      }
      if (domain.includes("localhost")) {
        return `http://${workspace.slug}.complexus.lc/my-work`;
      }
      return `https://${workspace.slug}.${domain}/my-work`;
    }
    return "/login";
  };

  return (
    <Box className="fixed left-0 top-2 z-10 w-screen md:top-4">
      <Container as="nav" className="md:w-max">
        <Box className="rounded-full">
          <Box className="z-10 flex h-[3.75rem] items-center justify-between gap-6 rounded-[1.25rem] border border-gray-200/60 bg-white/80 px-2.5 backdrop-blur-lg dark:border-dark-100/50 dark:bg-dark-300/70">
            <Logo className="relative -left-3.5 top-0.5 z-10 h-5 text-dark dark:text-gray-50 md:h-[1.6rem]" />
            <Flex align="center" className="hidden md:flex" gap={2}>
              <NavigationMenu>
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger
                      className={cn(
                        "rounded-[0.6rem] py-1.5 pl-3 pr-2.5 transition hover:bg-gray-100 dark:hover:bg-dark-200",
                        {
                          "bg-gray-100 dark:bg-dark-200":
                            pathname?.startsWith("/product"),
                        },
                      )}
                      hideArrow
                    >
                      Product
                    </NavigationMenu.Trigger>
                    <NavigationMenu.Content>
                      <Box className="grid w-max grid-cols-2 gap-2 p-2 pb-3 pr-2.5">
                        {product.map(
                          ({ id, name, description, icon, href }) => (
                            <MenuItem
                              description={description}
                              href={href}
                              icon={icon}
                              key={id}
                              name={name}
                            />
                          ),
                        )}
                      </Box>
                    </NavigationMenu.Content>
                  </NavigationMenu.Item>
                </NavigationMenu.List>
              </NavigationMenu>
              {navLinks.map(({ title, href }) => (
                <NavLink
                  className={cn(
                    "rounded-[0.6rem] px-3 py-1.5 transition hover:bg-gray-100 hover:dark:bg-dark-200",
                    {
                      "bg-gray-100 dark:bg-dark-200": pathname === href,
                    },
                  )}
                  href={href}
                  key={title}
                >
                  {title}
                </NavLink>
              ))}
              <NavigationMenu align="end">
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger
                      className={cn(
                        "rounded-[0.6rem] py-1.5 pl-3 pr-2.5 transition hover:bg-gray-100 hover:dark:bg-dark-200",
                      )}
                      hideArrow
                    >
                      Company
                    </NavigationMenu.Trigger>
                    <NavigationMenu.Content className="relative pb-1">
                      <Box className="grid w-max grid-cols-1 gap-2 p-2">
                        {company.map(
                          ({ id, name, description, icon, href }) => (
                            <MenuItem
                              description={description}
                              href={href}
                              icon={icon}
                              key={id}
                              name={name}
                            />
                          ),
                        )}
                      </Box>
                    </NavigationMenu.Content>
                  </NavigationMenu.Item>
                </NavigationMenu.List>
              </NavigationMenu>
            </Flex>
            <Flex align="center" className="ml-6 gap-3">
              <RequestDemo />
              <Button
                className={cn("hidden px-5 text-[0.93rem] md:flex", {
                  flex: session,
                })}
                color={session ? "invert" : "tertiary"}
                href={getNextUrl()}
                rounded="lg"
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
                          <NavLink className="flex text-xl" href={href}>
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
                      {product.map(({ id, name, href }) => (
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
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
