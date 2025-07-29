"use client";

import { Container, NavigationMenu } from "ui";
import { SprintsIcon, ObjectiveIcon, StoryIcon, OKRIcon } from "icons";
import Link from "next/link";
import type { ReactNode } from "react";

const MenuItem = ({
  name,

  href,
}: {
  name: string;
  description: string;
  icon: ReactNode;
  href: string;
}) => (
  <Link className="block text-lg" href={href}>
    {name}
  </Link>
);

export const Navigation = () => {
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

  return (
    <Container as="nav">
      <NavigationMenu>
        <NavigationMenu.List>
          <NavigationMenu.Item>
            <NavigationMenu.Trigger className="text-2xl font-medium" hideArrow>
              Product
            </NavigationMenu.Trigger>
            <NavigationMenu.Content className="px-4">
              {features.map(({ id, name, description, icon, href }) => (
                <MenuItem
                  description={description}
                  href={href}
                  icon={icon}
                  key={id}
                  name={name}
                />
              ))}
            </NavigationMenu.Content>
          </NavigationMenu.Item>
        </NavigationMenu.List>
      </NavigationMenu>
    </Container>
  );
};
