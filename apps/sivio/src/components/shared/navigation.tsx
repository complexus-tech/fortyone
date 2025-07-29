"use client";

import { Flex, NavigationMenu } from "ui";
import Link from "next/link";
import { Fragment } from "react";
import { Container } from "../ui";
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
    {
      name: "For Organisations",
      items: [
        { name: "How To Join", href: "/how-to-join" },
        { name: "Update Profile", href: "/update-profile" },
        { name: "Resources", href: "/resources" },
        { name: "Transparency", href: "/org-transparency" },
      ],
    },
    {
      name: "Stories",
      items: [
        { name: "Impact Stories", href: "/impact-stories" },
        { name: "Submit a Story", href: "/submit-story" },
      ],
    },
    {
      name: "News and Events",
      href: "/news",
    },
  ];

  return (
    <Container as="nav" className="flex items-center justify-between py-2">
      <Logo />
      <Flex align="center" className="gap-10">
        {nav.map((item, index) => (
          <Fragment key={index}>
            {item?.items ? (
              <NavigationMenu>
                <NavigationMenu.List>
                  <NavigationMenu.Item>
                    <NavigationMenu.Trigger
                      className="text-2xl font-medium"
                      hideArrow
                    >
                      {item.name}
                    </NavigationMenu.Trigger>
                    <NavigationMenu.Content className="min-w-64 space-y-6 px-5 py-4">
                      {item.items.map(({ name, href }) => (
                        <Link
                          className="block text-lg hover:text-primary"
                          href={href}
                          key={name}
                        >
                          {name}
                        </Link>
                      ))}
                    </NavigationMenu.Content>
                  </NavigationMenu.Item>
                </NavigationMenu.List>
              </NavigationMenu>
            ) : (
              <Link className="block text-2xl font-medium" href={item.href}>
                {item.name}
              </Link>
            )}
          </Fragment>
        ))}
      </Flex>
    </Container>
  );
};
