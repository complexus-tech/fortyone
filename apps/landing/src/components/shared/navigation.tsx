"use client";

import { Box, Button, Flex, NavLink, Text } from "ui";
import { SprintsIcon, ObjectiveIcon, StoryIcon, OKRIcon } from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import type { ReactNode } from "react";
import { Logo, Container } from "@/components/ui";
import { MobileNavigation } from "./mobile-navigation";

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
    className="hover:bg-state-hover flex w-68 gap-2 rounded-[0.7rem] p-2"
    href={href}
    prefetch
  >
    {icon}
    <Box>
      <Text className="dark:text-white">{name}</Text>
      <Text className="text-[0.85rem] opacity-70">{description}</Text>
    </Box>
  </Link>
);

export const Navigation = () => {
  const navLinks = [
    { title: "Pricing", href: "/pricing" },
    { title: "Contact", href: "/contact" },
    { title: "Blog", href: "/blog" },
    { title: "Help Center", href: "https://docs.fortyone.app" },
    { title: "Pitch", href: "https://pitch.fortyone.app" },
  ];

  const features = [
    {
      id: 1,
      name: "Tasks",
      href: "/features/tasks",
      description: "Turn ideas into clear, actionable work.",
      icon: (
        <StoryIcon className="text-foreground relative h-6 shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 2,
      name: "Objectives",
      href: "/features/objectives",
      description: "Focus teams on the outcomes that matter most.",
      icon: (
        <ObjectiveIcon className="text-foreground relative h-6 shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 3,
      name: "OKRs",
      href: "/features/okrs",
      description: "Align everyone on shared goals and see progress.",
      icon: (
        <OKRIcon className="text-foreground relative h-6 shrink-0 md:top-1 md:h-4" />
      ),
    },
    {
      id: 4,
      name: "Sprints",
      href: "/features/sprints",
      description:
        "Build momentum with short cycles that turn plans into progress.",
      icon: (
        <SprintsIcon className="text-foreground relative h-6 shrink-0 md:top-1 md:h-4" />
      ),
    },
  ];

  const pathname = usePathname();
  const getNextUrl = () => "https://cloud.fortyone.app/";

  return (
    <Box className="border-border/70 d/80 fixed left-0 z-15 w-screen border-b bg-white/20 backdrop-blur-xl dark:bg-black/40">
      <Container className="flex h-16 items-center justify-between gap-12">
        <Logo />
        <Flex align="center" className="hidden md:flex" gap={1}>
          {navLinks.map(({ title, href }) => (
            <NavLink
              className={cn(
                "flex items-center rounded-full px-3 opacity-90 transition hover:opacity-100",
                {
                  "opacity-100 dark:text-white dark:opacity-100":
                    pathname === href,
                },
              )}
              href={href}
              key={title}
              prefetch
            >
              {title}
            </NavLink>
          ))}
        </Flex>
        <Flex align="center" className="ml-4 gap-2">
          {/* <RequestDemo /> */}
          <Button
            className="hidden px-5 text-[0.93rem] md:flex"
            color="tertiary"
            href={getNextUrl()}
            rounded="lg"
            variant="naked"
          >
            Login
          </Button>
          <Button
            className="px-5 text-[0.93rem]"
            color="invert"
            href="https://cloud.fortyone.app/signup"
            rounded="lg"
          >
            Sign up
          </Button>

          <MobileNavigation />
        </Flex>
      </Container>
    </Box>
  );
};
