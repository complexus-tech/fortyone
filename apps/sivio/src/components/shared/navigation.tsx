"use client";

import { Container, Flex, NavigationMenu } from "ui";
import Link from "next/link";
import { Logo } from "./logo";

export const Navigation = () => {
  const nav = [
    {
      name: "About AfricaGiving",
      items: [
        { name: "About Us", href: "/about" },
        { name: "Reporting", href: "/reporting" },
        { name: "Contact Us", href: "/contact" },
      ],
    },
    {
      name: "For Donors",
      items: [
        { name: "For Individuals", href: "/for-individuals" },
        { name: "For Corporates", href: "/for-corporates" },
        { name: "Ways to Give", href: "/ways-to-give" },
        { name: "Transparency", href: "/transparency" },
      ],
    },
  ];

  return (
    <Container as="nav" className="flex items-center justify-between py-2">
      <Logo />
      <Flex align="center" className="gap-10">
        {nav.map((item, index) => (
          <NavigationMenu key={index}>
            <NavigationMenu.List>
              <NavigationMenu.Item>
                <NavigationMenu.Trigger
                  className="text-2xl font-medium"
                  hideArrow
                >
                  {item.name}
                </NavigationMenu.Trigger>
                <NavigationMenu.Content className="px-4">
                  {item.items.map(({ name, href }) => (
                    <Link className="block text-lg" href={href} key={name}>
                      {name}
                    </Link>
                  ))}
                </NavigationMenu.Content>
              </NavigationMenu.Item>
            </NavigationMenu.List>
          </NavigationMenu>
        ))}
      </Flex>
    </Container>
  );
};
