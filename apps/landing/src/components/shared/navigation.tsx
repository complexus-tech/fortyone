"use client";

import { Box, Button, Flex, NavLink, NavigationMenu, Text } from "ui";
import {
  AnalyticsIcon,
  ArrowRightIcon,
  CodeIcon,
  DocsIcon,
  MenuIcon,
  ObjectiveIcon,
  RoadmapIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { Logo, Container } from "@/components/ui";
import { usePathname } from "next/navigation";
import { cn } from "lib";

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
    className="flex w-64 gap-2 rounded-lg p-2 hover:bg-dark-100/40"
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
      icon: <StoryIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 2,
      name: "Objectives",
      href: "/product#objectives",
      description: "Set goals and track progress",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 3,
      name: "Roadmaps",
      href: "/product#roadmaps",
      description: "Plan and visualize your work",
      icon: <RoadmapIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 4,
      name: "Reports",
      href: "/product#reports",
      description: "Get analytics on your work",
      icon: <AnalyticsIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 5,
      name: "Integrations",
      href: "/product#integrations",
      description: "Connect third-party tools",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 6,
      name: "API",
      href: "/product#api",
      description: "Build custom workflows",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 7,
      name: "Mobile",
      href: "/product#mobile",
      description: "Stay connected on the go",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
  ];

  const resources = [
    {
      id: 2,
      href: "/blog",
      name: "Blog",
      description: "Read the latest articles",
      icon: <DocsIcon className="relative top-1 h-[1.15rem] w-auto shrink-0" />,
    },
    {
      id: 3,
      href: "/developers",
      name: "Developers",
      description: "Explore API documentation",
      icon: <CodeIcon className="relative top-1 h-5 w-auto shrink-0" />,
    },
  ];

  const pathname = usePathname();

  return (
    <Box className="fixed left-0 top-4 z-10 w-screen">
      <Container as="nav" className="md:w-max">
        <Box className="rounded-full">
          <Box className="z-10 flex h-[3.3rem] items-center justify-between rounded-full border border-gray-100/60 bg-white/60 px-2.5 backdrop-blur-lg dark:border-dark-200 dark:bg-black/40 md:h-16 md:px-3 md:pl-5">
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
              <NavigationMenu align="end" className="mx-2">
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
            <Flex align="center" className="ml-2 gap-4 pr-1 md:gap-2 md:pr-0">
              <Button
                className="hidden px-3 text-[0.93rem] md:inline-block"
                color="tertiary"
                rightIcon={
                  <ArrowRightIcon className="relative top-[0.5px] h-3 w-auto" />
                }
                rounded="full"
              >
                Sign in
              </Button>
              <Button className="text-[0.93rem]" rounded="full">
                Get started
              </Button>
              <MenuIcon className="h-6 w-auto" />
            </Flex>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
