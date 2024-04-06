"use client";

import { Box, Button, Flex, NavLink } from "ui";
import { ArrowRightIcon } from "icons";
import { Logo, Container } from "@/components/ui";

export const Navigation = () => {
  const navLinks = [
    { title: "Features", href: "/" },
    { title: "Integrations", href: "/" },
    { title: "Pricing", href: "/" },
    { title: "Company", href: "/" },
    { title: "Why Complexus", href: "/" },
    { title: "Contact", href: "/" },
    { title: "Resources", href: "/" },
  ];

  return (
    <Box className="fixed left-0 top-4 z-10 w-screen">
      <Container as="nav">
        <Box className="flex h-16 items-center justify-between rounded-full border border-dark-200 px-6 backdrop-blur-lg dark:bg-black/60">
          <Logo className="relative top-0.5 z-10 h-7 text-dark-100 dark:text-gray-50" />
          <Flex align="center" gap={4}>
            {navLinks.map(({ title, href }) => (
              <NavLink href={href} key={title}>
                {title}
              </NavLink>
            ))}
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
