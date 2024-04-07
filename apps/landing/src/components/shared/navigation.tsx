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
import { Logo, Container } from "@/components/ui";

const MenuItem = ({
  name,
  description,
  icon,
}: {
  name: string;
  description: string;
  icon: JSX.Element;
}) => (
  <Box className="flex w-64 gap-2 rounded-lg p-2 hover:bg-dark-200/80">
    {icon}
    <Box>
      <Text>{name}</Text>
      <Text className="text-[0.9rem]" color="muted">
        {description}
      </Text>
    </Box>
  </Box>
);

export const Navigation = () => {
  const navLinks = [
    { title: "Integrations", href: "/" },
    { title: "Pricing", href: "/" },
    { title: "Why Complexus", href: "/" },
    { title: "Contact", href: "/" },
  ];

  const features = [
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
          className="relative top-1 h-[1.1rem] w-auto shrink-0 text-dark"
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
        <Box className="z-10 flex h-16 items-center justify-between rounded-full border border-dark-200 px-6 backdrop-blur-lg dark:bg-black/40">
          <Logo className="relative top-0.5 z-10 h-7 text-dark-100 dark:text-gray-50" />
          <Flex align="center" gap={4}>
            <NavigationMenu>
              <NavigationMenu.List>
                <NavigationMenu.Item>
                  <NavigationMenu.Trigger>Features</NavigationMenu.Trigger>
                  <NavigationMenu.Content className="pb-1">
                    <Box className="grid w-max grid-cols-2 gap-2 p-2">
                      {features.map(({ id, name, description, icon }) => (
                        <MenuItem
                          description={description}
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
                  <NavigationMenu.Content className="pb-1">
                    <Box className="grid w-max grid-cols-2 gap-2 p-2">
                      {resources.map(({ id, name, description, icon }) => (
                        <MenuItem
                          description={description}
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
      </Container>
    </Box>
  );
};
