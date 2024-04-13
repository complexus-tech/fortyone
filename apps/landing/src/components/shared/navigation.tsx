"use client";

import { Box, Button, Flex, NavLink, NavigationMenu, Text } from "ui";
import {
  AnalyticsIcon,
  ArrowRightIcon,
  CodeIcon,
  DocsIcon,
  ObjectiveIcon,
  RoadmapIcon,
  StoryIcon,
} from "icons";
import Link from "next/link";
import { Logo, Container } from "@/components/ui";

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
    className="flex w-64 gap-2 rounded-lg p-2 hover:bg-primary/5"
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
    { title: "Integrations", href: "/" },
    { title: "Pricing", href: "/pricing" },
    { title: "About", href: "/about" },
  ];

  const product = [
    {
      id: 1,
      name: "Stories",
      description: "Track your work",
      icon: <StoryIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 2,
      name: "Objectives",
      description: "Set goals and track progress",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 3,
      name: "Roadmap",
      description: "Plan and visualize your work",
      icon: <RoadmapIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 4,
      name: "Reports",
      description: "Get analytics on your work",
      icon: <AnalyticsIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 5,
      name: "Integrations",
      description: "Connect third-party tools",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 6,
      name: "API",
      description: "Build custom workflows",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
    {
      id: 7,
      name: "Mobile",
      description: "Stay connected on the go",
      icon: <ObjectiveIcon className="relative top-1 h-4 w-auto shrink-0" />,
    },
  ];

  const resources = [
    {
      id: 1,
      name: "About us",
      description: "Learn more about complexus",
      icon: (
        <Logo
          asIcon
          className="relative top-1 h-[1.15rem] w-auto shrink-0 text-dark"
          fill="#dddddd"
        />
      ),
    },
    {
      id: 2,
      name: "Blog",
      description: "Read the latest articles",
      icon: <DocsIcon className="relative top-1 h-[1.15rem] w-auto shrink-0" />,
    },
    {
      id: 3,
      name: "Developers",
      description: "Explore API documentation",
      icon: <CodeIcon className="relative top-1 h-5 w-auto shrink-0" />,
    },
  ];

  return (
    <Box className="fixed left-0 top-4 z-10 w-screen">
      <Container as="nav">
        <Box className="rounded-full">
          <Box className="z-10 flex h-12 items-center justify-between rounded-full border border-gray-100/60 bg-white/60 px-2 backdrop-blur-lg dark:border-dark-200 dark:bg-black/40 md:h-16 md:px-6">
            <Logo className="relative top-0.5 z-10 h-5 text-secondary dark:text-gray-50 md:h-7" />
            <Flex align="center" className="hidden md:flex" gap={4}>
              <NavigationMenu>
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger>Product</NavigationMenu.Trigger>
                    <NavigationMenu.Content className="bg-warning/[0.02] pb-1">
                      <Box className="grid w-max grid-cols-2 gap-2 p-2">
                        {product.map(({ id, name, description, icon }) => (
                          <MenuItem
                            description={description}
                            href="/features"
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
                <NavLink href={href} key={title}>
                  {title}
                </NavLink>
              ))}
              <NavigationMenu align="end">
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger>Resources</NavigationMenu.Trigger>
                    <NavigationMenu.Content className="bg-warning/[0.02] pb-1">
                      <Box className="grid w-max grid-cols-2 gap-2 p-2">
                        {resources.map(({ id, name, description, icon }) => (
                          <MenuItem
                            description={description}
                            href="/features"
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
            </Flex>
            <Flex align="center" gap={3}>
              <Button
                className="px-3 text-sm"
                color="tertiary"
                rightIcon={
                  <ArrowRightIcon className="relative top-[0.5px] h-3 w-auto" />
                }
                rounded="full"
                size="sm"
              >
                Sign in
              </Button>
              <Button className="px-3 text-sm" rounded="full" size="sm">
                Get started
              </Button>
            </Flex>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
