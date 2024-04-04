"use client";

import { Box, Button, Flex, NavLink } from "ui";
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
    { title: "Sign in", href: "/" },
  ];

  return (
    <Box
      as="nav"
      className="fixed left-0 top-0 z-10 h-16 w-screen border-b backdrop-blur-lg dark:border-primary/5 dark:bg-black/5"
    >
      <Container className="relative z-[1] flex h-full items-center justify-between">
        <Logo className="relative top-0.5 z-10 h-7 text-dark-100 dark:text-gray-50" />
        <Flex align="center" gap={4}>
          {navLinks.map(({ title, href }) => (
            <NavLink href={href} key={title}>
              {title}
            </NavLink>
          ))}
          <Button className="px-3 text-sm" rounded="full" size="sm">
            Get started
          </Button>
        </Flex>
      </Container>
    </Box>
  );
};
