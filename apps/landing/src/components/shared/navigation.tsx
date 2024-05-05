"use client";

import { Box, Button, Flex, Menu, NavLink, NavigationMenu, Text } from "ui";
import {
  AnalyticsIcon,
  ArrowRightIcon,
  ChatIcon,
  CodeIcon,
  DocsIcon,
  EpicsIcon,
  MenuIcon,
  SprintsIcon,
  ObjectiveIcon,
  RoadmapIcon,
  StoryIcon,
  WhiteboardIcon,
} from "icons";
import Link from "next/link";
import { Logo, Container } from "@/components/ui";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import { MenuButton } from "./menu-button";
import { useState } from "react";

const MenuItem = ({
  name,
  description,
  icon,
  href,
}: {
  name: string;
  description: string;
  icon: JSX.Element;
  href: string;
}) => (
  <Link
    className="flex w-[17rem] gap-2 rounded-lg p-2 hover:bg-dark-100/40"
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

export const Navigation = () => {
  const navLinks = [
    { title: "Pricing", href: "/pricing" },
    { title: "About", href: "/about" },
    { title: "Contact", href: "/contact" },
  ];

  const product = [
    {
      id: 1,
      name: "Stories",
      href: "/product#stories",
      description: "Track your work",
      icon: <StoryIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 2,
      name: "Objectives",
      href: "/product#objectives",
      description: "Set goals and track progress",
      icon: <ObjectiveIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 3,
      name: "Roadmaps",
      href: "/product#roadmaps",
      description: "Plan and visualize your work",
      icon: <RoadmapIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 4,
      name: "Sprints",
      href: "/product#sprints",
      description: "Organize work into sprints",
      icon: <SprintsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 5,
      name: "Epics",
      href: "/product#epics",
      description: "Break down work into epics",
      icon: <EpicsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 6,
      name: "Documents",
      href: "/product#documents",
      description: "Store and share files",
      icon: <DocsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 7,
      name: "Reporting",
      href: "/product#reports",
      description: "Analyze and report on progress",
      icon: <AnalyticsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 8,
      name: "Discussions",
      href: "/product#discussions",
      description: "Discuss work with your team",
      icon: <ChatIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 9,
      name: "Whiteboards",
      href: "/product#whiteboards",
      description: "Sketch and brainstorm ideas",
      icon: (
        <WhiteboardIcon className="relative h-4 w-auto shrink-0 md:top-1" />
      ),
    },
  ];

  const resources = [
    {
      id: 2,
      href: "/blog",
      name: "Blog",
      description: "Read the latest articles",
      icon: (
        <DocsIcon className="relative h-[1.15rem] w-auto shrink-0 md:top-1" />
      ),
    },
    {
      id: 3,
      href: "/developers",
      name: "Developers",
      description: "Explore API documentation",
      icon: <CodeIcon className="relative h-5 w-auto shrink-0 md:top-1" />,
    },
  ];

  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const toggleMenu = () => setIsMenuOpen((open) => !open);

  const pathname = usePathname();

  return (
    <Box className="fixed left-0 top-4 z-10 w-screen">
      <Container as="nav" className="md:w-max">
        <Box className="rounded-full">
          <Box className="z-10 flex h-[3.75rem] items-center justify-between rounded-full border border-gray-100/60 bg-white/60 px-2.5 backdrop-blur-lg dark:border-dark-200 dark:bg-black/40 md:h-16 md:px-3 md:pl-5">
            <Logo className="relative top-0.5 z-10 h-5 text-secondary dark:text-gray-50 md:h-[1.6rem]" />
            <Flex align="center" className="hidden md:flex" gap={2}>
              <NavigationMenu>
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger
                      className={cn(
                        "rounded-3xl px-3 py-1.5 transition hover:bg-dark-300/80",
                        {
                          "bg-dark-300/80": pathname.startsWith("/product"),
                        },
                      )}
                    >
                      Product
                    </NavigationMenu.Trigger>
                    <NavigationMenu.Content className="bg-warning/[0.02] pb-1">
                      <Box className="grid w-max grid-cols-2 gap-2 p-2">
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
                  href={href}
                  className={cn(
                    "rounded-3xl px-3 py-1.5 transition hover:bg-dark-300/80",
                    {
                      "bg-dark-300/80": pathname === href,
                    },
                  )}
                  key={title}
                >
                  {title}
                </NavLink>
              ))}
              {/* <NavigationMenu align="end" className="mx-2">
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger>Resources</NavigationMenu.Trigger>
                    <NavigationMenu.Content className="bg-warning/[0.02] pb-1">
                      <Box className="grid w-max grid-cols-1 gap-2 p-2">
                        {resources.map(
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
              </NavigationMenu> */}
            </Flex>
            <Flex align="center" className="ml-2 gap-2 pr-1 md:pr-0">
              <Button
                className="relative text-[0.93rem] md:left-1"
                rounded="full"
                href="https://forms.gle/NmG4XFS5GhvRjUxu6"
              >
                Join waitlist
              </Button>
              <Menu open={isMenuOpen} onOpenChange={setIsMenuOpen}>
                <Menu.Button asChild>
                  <button className="flex md:hidden">
                    <MenuButton open={isMenuOpen} />
                  </button>
                </Menu.Button>
                <Menu.Items
                  align="end"
                  className="relative left-4 mt-4 w-[calc(100vw-2.5rem)] rounded-3xl py-4"
                >
                  <Menu.Group>
                    {navLinks.map(({ title, href }) => (
                      <Menu.Item
                        key={title}
                        className="block rounded-lg py-2"
                        onClick={() => {
                          setIsMenuOpen(false);
                        }}
                      >
                        <NavLink className="flex" href={href}>
                          {title}
                        </NavLink>
                      </Menu.Item>
                    ))}
                  </Menu.Group>
                  <Menu.Separator />
                  <Menu.Group>
                    {product.map(({ id, name, href, icon }) => (
                      <Menu.Item
                        key={id}
                        className="block rounded-lg py-2"
                        onClick={() => {
                          setIsMenuOpen(false);
                        }}
                      >
                        <NavLink
                          className="flex items-center gap-2"
                          href={href}
                        >
                          {icon}
                          {name}
                        </NavLink>
                      </Menu.Item>
                    ))}
                  </Menu.Group>
                </Menu.Items>
              </Menu>
            </Flex>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
