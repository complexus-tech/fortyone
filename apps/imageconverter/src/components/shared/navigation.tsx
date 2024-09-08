"use client";

import { Box, Button, Flex, NavLink } from "ui";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import { Logo, Container } from "@/components/ui";

export const Navigation = () => {
  const navLinks = [
    { title: "Pricing", href: "/pricing" },
    { title: "About", href: "/about" },
    { title: "Contact", href: "/contact" },
  ];

  const pathname = usePathname();

  return (
    <Box className="fixed left-0 top-4 z-10 w-screen">
      <Container as="nav" className="md:w-max">
        <Box className="rounded-2xl">
          <Box className="z-10 flex h-[3.2rem] items-center justify-between rounded-full border border-gray-100/60 bg-white/60 px-2.5 backdrop-blur-lg md:px-2.5 dark:border-dark-200 dark:bg-black/40">
            <Logo className="mr-10" />
            <Flex align="center" className="hidden md:flex" gap={4}>
              {navLinks.map(({ title, href }) => (
                <NavLink
                  className={cn(
                    "rounded-3xl px-3 py-1.5 transition hover:bg-dark-300/80",
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
            </Flex>
            <Flex align="center" className="ml-2 gap-4 pr-1 md:pr-0">
              <Button
                className="relative px-4 text-[0.93rem] md:left-1"
                color="tertiary"
                rounded="full"
                size="sm"
              >
                Log in
              </Button>
              <Button className="relative" size="sm" rounded="full">
                Sign up
              </Button>
            </Flex>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
