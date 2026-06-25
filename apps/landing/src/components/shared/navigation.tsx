"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import { Box, Button, Flex, NavigationMenu, NavLink } from "ui";
import { Logo, Container } from "@/components/ui";
import { APP_URL, SIGNUP_URL } from "@/lib/app-url";
import { primaryUseCaseLinks } from "@/lib/use-case-links";
import { MobileNavigation } from "./mobile-navigation";
// import { RequestDemo } from "./request-demo";

type NavigationMenuItem = {
  href: string;
  title: string;
};

const resourceLinks: NavigationMenuItem[] = [
  {
    title: "Docs",
    href: "https://docs.fortyone.app",
  },
  {
    title: "Blog",
    href: "/blog",
  },
  {
    title: "GitHub",
    href: "https://github.com/complexus-tech/fortyone",
  },
  {
    title: "Pitch",
    href: "https://pitch.fortyone.app",
  },
];

const useCaseMenuLinks: NavigationMenuItem[] = primaryUseCaseLinks.map(
  ({ href, label }) => ({
    href,
    title: label,
  }),
);

const isExternalLink = (href: string) => href.startsWith("http");

const NavigationMenuLink = ({ href, title }: NavigationMenuItem) => (
  <NavigationMenu.Link asChild>
    <Link
      className="hover:bg-accent focus:bg-accent focus-visible:bg-accent flex w-full items-center rounded-lg px-2 py-1.5 text-[0.95rem] leading-6 whitespace-nowrap transition-colors outline-none select-none focus-visible:outline-none"
      href={href}
      prefetch={!isExternalLink(href)}
      rel={isExternalLink(href) ? "noreferrer" : undefined}
      target={isExternalLink(href) ? "_blank" : undefined}
    >
      {title}
    </Link>
  </NavigationMenu.Link>
);

const NavigationDropdown = ({
  contentClassName,
  items,
  label,
}: {
  contentClassName: string;
  items: NavigationMenuItem[];
  label: string;
}) => (
  <NavigationMenu.Item className="relative">
    <NavigationMenu.Trigger
      className="rounded-full px-3 opacity-90 transition outline-none hover:opacity-100 focus:outline-none focus-visible:outline-none data-[state=open]:opacity-100"
      hideArrow
    >
      {label}
    </NavigationMenu.Trigger>
    <NavigationMenu.Content
      className={cn("top-full z-50 mt-1.5", contentClassName)}
    >
      <Box className="border-border shadow-shadow dark:bg-surface-elevated rounded-xl border bg-white px-1.5 py-1.5 shadow-xl">
        {items.map((item) => (
          <NavigationMenuLink key={item.href} {...item} />
        ))}
      </Box>
    </NavigationMenu.Content>
  </NavigationMenu.Item>
);

const DesktopNavItem = ({ href, title }: NavigationMenuItem) => {
  const pathname = usePathname();

  return (
    <NavLink
      className={cn(
        "flex items-center rounded-full px-3 opacity-90 transition hover:opacity-100",
        {
          "opacity-100 dark:text-white dark:opacity-100": pathname === href,
        },
      )}
      href={href}
      prefetch
    >
      {title}
    </NavLink>
  );
};

export const Navigation = ({ hasSession }: { hasSession: boolean }) => {
  return (
    <Box className="border-border/40 fixed left-0 z-15 w-screen border-b bg-white/60 backdrop-blur-xl dark:bg-black/50">
      <Container className="flex h-16 items-center justify-between gap-12">
        <Logo />
        <NavigationMenu className="hidden md:flex" showViewport={false}>
          <NavigationMenu.List className="gap-7 space-x-0 lg:gap-10">
            <NavigationDropdown
              contentClassName="min-w-44"
              items={useCaseMenuLinks}
              label="Use Cases"
            />
            <NavigationDropdown
              contentClassName="min-w-40"
              items={resourceLinks}
              label="Resources"
            />
            <NavigationMenu.Item>
              <DesktopNavItem href="/pricing" title="Pricing" />
            </NavigationMenu.Item>
          </NavigationMenu.List>
        </NavigationMenu>
        <Flex align="center" className="ml-4 gap-2">
          {/* <RequestDemo /> */}
          {hasSession ? (
            <Button
              className="px-5 text-[0.93rem]"
              color="invert"
              href={APP_URL}
              rounded="lg"
            >
              Open app
            </Button>
          ) : (
            <>
              <Button
                className="hidden px-5 text-[0.93rem] md:flex"
                color="tertiary"
                href={APP_URL}
                rounded="full"
                variant="naked"
              >
                Login
              </Button>
              <Button
                className="px-5 text-[0.93rem]"
                color="invert"
                href={SIGNUP_URL}
                rounded="full"
              >
                Sign up
              </Button>
            </>
          )}

          <MobileNavigation />
        </Flex>
      </Container>
    </Box>
  );
};
