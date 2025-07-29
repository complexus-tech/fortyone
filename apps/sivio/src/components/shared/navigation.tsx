"use client";

import { Box, Container, Flex, NavLink, NavigationMenu, Text } from "ui";
import { SprintsIcon, ObjectiveIcon, StoryIcon, OKRIcon } from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import type { ReactNode } from "react";

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

export const Navigation = () => {
  const navLinks = [
    { title: "Pricing", href: "/pricing" },
    { title: "Contact", href: "/contact" },
    { title: "Blog", href: "/blog" },
    { title: "Docs", href: "https://docs.complexus.app" },
  ];

  const features = [
    {
      id: 1,
      name: "Stories",
      href: "/features/stories",
      description: "Manage and Track Tasks",
      icon: (
        <StoryIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 2,
      name: "Objectives",
      href: "/features/objectives",
      description: "Set and Achieve Goals",
      icon: (
        <ObjectiveIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 3,
      name: "OKRs",
      href: "/features/okrs",
      description: "Align and Achieve",
      icon: (
        <OKRIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 4,
      name: "Sprints",
      href: "/features/sprints",
      description: "Iterate and Deliver",
      icon: (
        <SprintsIcon className="relative h-6 w-auto shrink-0 md:top-1 md:h-4" />
      ),
    },
  ];

  const pathname = usePathname();

  return (
    <Box className="fixed left-0 top-2 z-10 w-screen md:top-5">
      <Container as="nav" className="md:w-max">
        <Box className="rounded-full">
          <Box className="z-10 flex h-[3.25rem] items-center justify-between gap-12 rounded-full border border-gray-100/20 bg-[#dddddd]/40 px-[0.35rem] font-medium backdrop-blur-xl dark:border-dark-100/50 dark:bg-dark-200/40">
            Logo
            <Flex align="center" className="hidden md:flex" gap={1}>
              <NavigationMenu>
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger
                      className={cn(
                        "rounded-full py-1.5 pl-3 pr-2.5 transition hover:bg-gray-100 dark:hover:bg-dark-200",
                        {
                          "bg-gray-100 dark:bg-dark-200":
                            pathname?.startsWith("/features"),
                        },
                      )}
                      hideArrow
                    >
                      Product
                    </NavigationMenu.Trigger>
                    <NavigationMenu.Content>
                      <Box className="grid w-max grid-cols-2 gap-2 p-2 pb-3 pr-2.5">
                        {features.map(
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
                    "rounded-full px-3 py-1.5 transition hover:bg-gray-100 hover:dark:bg-dark-200",
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
            </Flex>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
