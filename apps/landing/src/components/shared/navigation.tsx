"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
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

const NAVIGATION_DOCK_ENTER_Y = 16;
const NAVIGATION_DOCK_EXIT_Y = 4;

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
  const [isDocked, setIsDocked] = useState(false);

  useEffect(() => {
    const updateDockedState = () => {
      setIsDocked((wasDocked) =>
        wasDocked
          ? window.scrollY > NAVIGATION_DOCK_EXIT_Y
          : window.scrollY > NAVIGATION_DOCK_ENTER_Y,
      );
    };

    updateDockedState();
    window.addEventListener("scroll", updateDockedState, { passive: true });

    return () => {
      window.removeEventListener("scroll", updateDockedState);
    };
  }, []);

  return (
    <Box
      className="fixed inset-x-0 top-0 z-15"
      data-docked={isDocked ? "true" : "false"}
    >
      <Box
        aria-hidden="true"
        className={cn(
          "bg-background/90 absolute inset-0 backdrop-blur-xl transition-opacity duration-[160ms] motion-reduce:transition-none",
          isDocked ? "opacity-0" : "opacity-100",
        )}
      />
      <Box
        aria-hidden="true"
        className={cn(
          "border-border/60 bg-background/95 shadow-shadow absolute inset-x-2 top-3 h-full rounded-2xl border shadow-xl backdrop-blur-xl transition-[transform,opacity] duration-[220ms] [transition-timing-function:var(--landing-ease-out)] motion-reduce:transition-none md:inset-x-6 md:rounded-3xl",
          isDocked
            ? "translate-y-0 scale-100 opacity-100"
            : "-translate-y-2 scale-[0.985] opacity-0",
        )}
      />
      <Container
        className={cn(
          "relative flex items-center justify-between gap-12 py-3 transition-transform duration-[220ms] [transition-timing-function:var(--landing-ease-out)] motion-reduce:transition-none",
          isDocked ? "translate-y-3" : "translate-y-0",
        )}
        full
      >
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
