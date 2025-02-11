"use client";

import { Box, Button, Flex, Menu, NavLink, NavigationMenu, Text } from "ui";
import {
  AnalyticsIcon,
  ChatIcon,
  DocsIcon,
  EpicsIcon,
  SprintsIcon,
  ObjectiveIcon,
  RoadmapIcon,
  StoryIcon,
  WhiteboardIcon,
  CodeIcon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import { useState } from "react";
import { Logo, Container } from "@/components/ui";
import { MenuButton } from "./menu-button";

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
      description: "Manage and Track Tasks",
      icon: <StoryIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 2,
      name: "Objectives",
      href: "/product#objectives",
      description: "Set and Achieve Goals",
      icon: <ObjectiveIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 3,
      name: "Roadmaps",
      href: "/product#roadmaps",
      description: "Plan Your Objectives' Journey",
      icon: <RoadmapIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 4,
      name: "Sprints",
      href: "/product#sprints",
      description: "Iterate and Deliver",
      icon: <SprintsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 5,
      name: "Epics",
      href: "/product#epics",
      description: "Oversee Major Initiatives",
      icon: <EpicsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 6,
      name: "Documents",
      href: "/product#documents",
      description: "Store and Collaborate",
      icon: <DocsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 7,
      name: "Reporting",
      href: "/product#reports",
      description: "Gain Insights and Analyze",
      icon: <AnalyticsIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 8,
      name: "Messaging",
      href: "/product#messaging",
      description: "Communicate and Collaborate",
      icon: <ChatIcon className="relative h-4 w-auto shrink-0 md:top-1" />,
    },
    {
      id: 9,
      name: "Whiteboards",
      href: "/product#whiteboards",
      description: "Sketch and Brainstorm ideas",
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

  const pathname = usePathname();

  return (
    <Box className="fixed left-0 top-6 z-10 w-screen">
      <Container as="nav" className="md:w-max">
        <Box className="rounded-full">
          <Box className="z-10 flex h-[3.7rem] items-center justify-between rounded-full border border-gray-100/60 bg-white/60 px-2.5 backdrop-blur-lg dark:border-dark-200 dark:bg-black/40">
            <Logo className="relative -left-3.5 top-0.5 z-10 h-5 text-secondary dark:text-gray-50 md:h-[1.6rem]" />
            <Flex align="center" className="hidden md:flex" gap={2}>
              <NavigationMenu>
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger
                      className={cn(
                        "rounded-3xl px-2.5 py-1.5 transition hover:bg-dark-300/80",
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
                  className={cn(
                    "rounded-3xl px-2.5 py-1.5 transition hover:bg-dark-300/80",
                    {
                      "bg-dark-300/80": pathname === href,
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
              </NavigationMenu>
            </Flex>
            <Flex align="center" className="ml-6 gap-2">
              <Button
                className="px-4 text-[0.93rem]"
                color="tertiary"
                rounded="full"
              >
                Sign up
              </Button>
              <Button
                className="px-4 text-[0.93rem] "
                href="https://forms.gle/NmG4XFS5GhvRjUxu6"
                rounded="full"
              >
                Log in
              </Button>

              <Box className="flex md:hidden">
                <Menu onOpenChange={setIsMenuOpen} open={isMenuOpen}>
                  <Menu.Button asChild>
                    <button type="button">
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
                          className="block rounded-lg py-2"
                          key={title}
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
                          className="block rounded-lg py-2"
                          key={id}
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
              </Box>
            </Flex>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
