"use client";

import Link from "next/link";
import { cn } from "lib";
import { Box, Button, Flex, NavigationMenu, NavLink } from "ui";
import { Logo, Container } from "@/components/ui";
import { APP_URL, SIGNUP_URL } from "@/lib/app-url";
import { featureLinks } from "@/lib/feature-links";
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
    title: "Pitch",
    href: "https://pitch.fortyone.app",
  },
];

const featureMenuLinks: NavigationMenuItem[] = featureLinks.map(
  ({ href, label }) => ({
    href,
    title: label,
  }),
);

const primaryUseCaseMenuLinks: NavigationMenuItem[] = primaryUseCaseLinks.map(
  ({ href, label }) => ({
    href,
    title: label,
  }),
);

const isExternalLink = (href: string) => href.startsWith("http");

const NavigationMenuLink = ({ href, title }: NavigationMenuItem) => (
  <NavigationMenu.Link asChild>
    <Link
      className="hover:bg-accent focus:bg-accent focus-visible:bg-accent dark:hover:bg-surface-prominent dark:focus:bg-surface-prominent dark:focus-visible:bg-surface-prominent flex w-full items-center rounded-md px-2.5 py-1.5 text-[0.95rem] leading-6 whitespace-nowrap transition-colors outline-none select-none focus-visible:outline-none"
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
      className="hover:bg-state-hover focus-visible:bg-state-hover data-[state=open]:bg-state-hover rounded-md px-3 py-1.5 transition outline-none focus:outline-none focus-visible:outline-none"
      hideArrow
    >
      {label}
    </NavigationMenu.Trigger>
    <NavigationMenu.Content
      className={cn("top-full z-50 mt-1.5", contentClassName)}
    >
      <Box className="border-border/70 bg-popover text-popover-foreground shadow-shadow rounded-lg border px-1.5 py-1.5 shadow-xl">
        {items.map((item) => (
          <NavigationMenuLink key={item.href} {...item} />
        ))}
      </Box>
    </NavigationMenu.Content>
  </NavigationMenu.Item>
);

const DesktopNavItem = ({ href, title }: NavigationMenuItem) => {
  return (
    <NavLink
      className="hover:bg-state-hover flex items-center rounded-md px-3 py-1.5 transition"
      href={href}
      prefetch
    >
      {title}
    </NavLink>
  );
};

export const Navigation = ({ hasSession }: { hasSession: boolean }) => {
  return (
    <Box className="bg-background/90 fixed left-0 z-15 w-screen py-3.5 backdrop-blur-xl">
      <Container className="flex items-center justify-between gap-12" full>
        <Logo />
        <NavigationMenu className="hidden md:flex" showViewport={false}>
          <NavigationMenu.List className="gap-4 space-x-0 lg:gap-6">
            <NavigationDropdown
              contentClassName="min-w-44"
              items={featureMenuLinks}
              label="Features"
            />
            <NavigationDropdown
              contentClassName="min-w-44"
              items={primaryUseCaseMenuLinks}
              label="Use Cases"
            />
            <NavigationMenu.Item>
              <DesktopNavItem
                href="/ai-project-manager"
                title="AI Project Manager"
              />
            </NavigationMenu.Item>
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
              color="invert"
              href={APP_URL}
              rounded="lg"
              size="lg"
              variant="outline"
            >
              Open app
            </Button>
          ) : (
            <>
              <Button
                className="hidden text-[0.93rem] md:flex"
                color="tertiary"
                href={APP_URL}
                rounded="lg"
                size="lg"
                variant="naked"
              >
                Login
              </Button>
              <Button
                className="text-[0.93rem]"
                color="invert"
                href={SIGNUP_URL}
                rounded="lg"
                size="lg"
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
