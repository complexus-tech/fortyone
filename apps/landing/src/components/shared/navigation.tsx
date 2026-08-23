"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { ArrowDown2Icon } from "icons";
import { cn } from "lib";
import { Box, Button, Flex, NavLink } from "ui";
import { Logo, Container } from "@/components/ui";
import { APP_URL, SIGNUP_URL } from "@/lib/app-url";
import { featureLinks } from "@/lib/feature-links";
import { primaryUseCaseLinks } from "@/lib/use-case-links";
import { MobileNavigation } from "./mobile-navigation";
import {
  NavigationMenuIcon,
  type NavigationIconName,
  type NavigationIconTone,
} from "./navigation-menu-icon";
// import { RequestDemo } from "./request-demo";

type NavigationLinkItem = {
  href: string;
  title: string;
};

type NavigationMenuItem = NavigationLinkItem & {
  description: string;
  icon: NavigationIconName;
  tone: NavigationIconTone;
};

const resourceLinks: NavigationMenuItem[] = [
  {
    description: "Learn how FortyOne works",
    href: "https://docs.fortyone.app",
    icon: "documents",
    title: "Docs",
    tone: "blue",
  },
  {
    description: "Ideas for building focused teams",
    href: "/blog",
    icon: "blog",
    title: "Blog",
    tone: "lime",
  },
  {
    description: "See the product story in a few minutes",
    href: "https://pitch.fortyone.app",
    icon: "pitch",
    title: "Pitch",
    tone: "orange",
  },
];

const featureMenuDetails = {
  "ai-planning": {
    description: "Let Maya shape a realistic plan",
    icon: "ai-planning",
    tone: "lime",
  },
  calendar: {
    description: "See meetings and planned focus time",
    icon: "calendar",
    tone: "rose",
  },
  "customer-feedback": {
    description: "Turn customer insight into action",
    icon: "customer-feedback",
    tone: "blue",
  },
  documents: {
    description: "Keep context beside the work",
    icon: "documents",
    tone: "orange",
  },
  goals: {
    description: "Set objectives and track key results",
    icon: "goals",
    tone: "lime",
  },
  integrations: {
    description: "Connect the tools your team uses",
    icon: "integrations",
    tone: "blue",
  },
  roadmaps: {
    description: "Communicate what is coming next",
    icon: "roadmaps",
    tone: "aqua",
  },
  "strategy-map": {
    description: "Connect priorities across your company",
    icon: "strategy-map",
    tone: "lilac",
  },
  tasks: {
    description: "Plan and deliver the day-to-day work",
    icon: "tasks",
    tone: "aqua",
  },
} satisfies Record<
  (typeof featureLinks)[number]["slug"],
  {
    description: string;
    icon: NavigationIconName;
    tone: NavigationIconTone;
  }
>;

const featureMenuLinks: NavigationMenuItem[] = featureLinks.map(
  ({ href, label, slug }) => ({
    ...featureMenuDetails[slug],
    href,
    title: label,
  }),
);

const useCaseMenuDetails = {
  "customer-support": {
    description: "Turn customer signals into action",
    icon: "customer-support",
    tone: "rose",
  },
  developers: {
    description: "Plan, build, and ship with context",
    icon: "developers",
    tone: "lilac",
  },
  "field-crews": {
    description: "Keep distributed teams moving together",
    icon: "operations",
    tone: "orange",
  },
  government: {
    description: "Coordinate accountable public work",
    icon: "government",
    tone: "blue",
  },
  marketing: {
    description: "Connect campaigns to company goals",
    icon: "marketing",
    tone: "orange",
  },
  leadership: {
    description: "See priorities, progress, and risk clearly",
    icon: "goals",
    tone: "blue",
  },
  operations: {
    description: "Coordinate recurring cross-team work",
    icon: "operations",
    tone: "aqua",
  },
  product: {
    description: "Link discovery, strategy, and delivery",
    icon: "product",
    tone: "lime",
  },
} satisfies Record<
  (typeof primaryUseCaseLinks)[number]["slug"],
  {
    description: string;
    icon: NavigationIconName;
    tone: NavigationIconTone;
  }
>;

const primaryUseCaseMenuLinks: NavigationMenuItem[] = primaryUseCaseLinks.map(
  ({ href, label, slug }) => ({
    ...useCaseMenuDetails[slug],
    href,
    title: label,
  }),
);

const isExternalLink = (href: string) => href.startsWith("http");

const NAVIGATION_DOCK_ENTER_Y = 16;
const NAVIGATION_DOCK_EXIT_Y = 4;

const NavigationMenuLink = ({
  description,
  href,
  icon,
  title,
  tone,
}: NavigationMenuItem) => (
  <Link
    className="group hover:bg-accent/70 focus:bg-accent/70 focus-visible:bg-accent/70 dark:hover:bg-surface-prominent/75 dark:focus:bg-surface-prominent/75 dark:focus-visible:bg-surface-prominent/75 flex min-w-0 items-start gap-3 rounded-xl px-2.5 py-2.5 transition-colors outline-none select-none focus-visible:outline-none"
    href={href}
    prefetch={!isExternalLink(href)}
    rel={isExternalLink(href) ? "noreferrer" : undefined}
    target={isExternalLink(href) ? "_blank" : undefined}
  >
    <NavigationMenuIcon name={icon} tone={tone} />
    <span className="min-w-0 pt-0.5">
      <span className="block text-[0.93rem] leading-5 font-medium tracking-[-0.01em]">
        {title}
      </span>
      <span className="text-text-muted mt-0.5 block text-[0.78rem] leading-[1.08rem]">
        {description}
      </span>
    </span>
  </Link>
);

const NavigationDropdown = ({
  contentClassName,
  gridClassName,
  heading,
  isOpen,
  items,
  label,
  onOpenChange,
  value,
}: {
  contentClassName: string;
  gridClassName: string;
  heading?: string;
  isOpen: boolean;
  items: NavigationMenuItem[];
  label: string;
  onOpenChange: (value: string) => void;
  value: string;
}) => (
  <li
    className="group/dropdown relative"
    onBlur={(event) => {
      if (
        !(event.relatedTarget instanceof Node) ||
        !event.currentTarget.contains(event.relatedTarget)
      ) {
        onOpenChange("");
      }
    }}
    onMouseLeave={() => {
      onOpenChange("");
    }}
  >
    <button
      aria-controls={`${value}-navigation-menu`}
      aria-expanded={isOpen}
      className="hover:bg-state-hover focus-visible:bg-state-hover flex cursor-pointer items-center gap-1 rounded-md px-3 py-1.5 text-[0.95rem] transition outline-none select-none focus:outline-none focus-visible:outline-none"
      onClick={() => {
        onOpenChange(isOpen ? "" : value);
      }}
      type="button"
    >
      {label}
      <ArrowDown2Icon
        aria-hidden="true"
        className="h-3 w-auto transition-transform duration-200 group-focus-within/dropdown:rotate-180 group-hover/dropdown:rotate-180 motion-reduce:transition-none"
        strokeWidth={3}
      />
    </button>
    <div
      className={cn(
        "pointer-events-none invisible absolute top-full z-50 translate-y-1 pt-2 opacity-0 transition-[transform,opacity,visibility] delay-150 duration-[160ms] [transition-timing-function:var(--landing-ease-out)] group-focus-within/dropdown:pointer-events-auto group-focus-within/dropdown:visible group-focus-within/dropdown:translate-y-0 group-focus-within/dropdown:opacity-100 group-focus-within/dropdown:delay-0 group-hover/dropdown:pointer-events-auto group-hover/dropdown:visible group-hover/dropdown:translate-y-0 group-hover/dropdown:opacity-100 group-hover/dropdown:delay-0 motion-reduce:transition-none",
        isOpen && "pointer-events-auto visible translate-y-0 opacity-100",
        contentClassName,
      )}
      id={`${value}-navigation-menu`}
    >
      <Box
        className={cn(
          "border-border/60 bg-popover/98 text-popover-foreground shadow-shadow rounded-xl border p-3 shadow-[0_24px_60px_-24px_rgba(31,24,18,0.38)] backdrop-blur-xl dark:shadow-[0_24px_60px_-24px_rgba(0,0,0,0.78)]",
        )}
      >
        {heading ? (
          <div className="text-text-muted px-2.5 pt-1 pb-2 text-[0.7rem] font-medium tracking-[0.12em] uppercase">
            {heading}
          </div>
        ) : null}
        <div className={cn("grid gap-1", gridClassName)}>
          {items.map((item) => (
            <NavigationMenuLink key={item.href} {...item} />
          ))}
        </div>
      </Box>
    </div>
  </li>
);

const DesktopNavItem = ({ href, title }: NavigationLinkItem) => {
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

export const Navigation = () => {
  const [activeMenu, setActiveMenu] = useState("");
  const [isDocked, setIsDocked] = useState(false);
  const navigationRef = useRef<HTMLElement>(null);

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

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (
        event.key === "Escape" &&
        document.activeElement instanceof HTMLElement &&
        navigationRef.current?.contains(document.activeElement)
      ) {
        setActiveMenu("");
        document.activeElement.blur();
      }
    };

    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  return (
    <Box
      className="fixed inset-x-0 top-1 z-15"
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
          "landing-page-frame dark:bg-surface-elevated/80 absolute inset-x-0 top-3 h-full rounded-2xl bg-white shadow-[0_16px_44px_-18px_rgba(31,24,18,0.3)] backdrop-blur-xl transition-[transform,opacity] duration-[220ms] [transition-timing-function:var(--landing-ease-out)] motion-reduce:transition-none md:rounded-3xl dark:shadow-[0_16px_44px_-18px_rgba(0,0,0,0.7)]",
          isDocked
            ? "translate-y-0 scale-100 opacity-100"
            : "-translate-y-2 scale-[0.985] opacity-0",
        )}
      />
      <Container
        className={cn(
          "landing-page-frame relative flex items-center justify-between gap-8 px-5 py-3 transition-transform duration-[220ms] [transition-timing-function:var(--landing-ease-out)] motion-reduce:transition-none md:px-6",
          isDocked ? "translate-y-3" : "translate-y-0",
        )}
      >
        <Flex align="center" className="min-w-0 gap-8">
          <Logo />
          <nav
            aria-label="Main"
            className="landing-desktop-navigation relative z-10"
            ref={navigationRef}
          >
            <ul className="flex list-none items-center gap-1">
              <NavigationDropdown
                contentClassName="w-[46rem] max-w-[calc(100vw-3rem)]"
                gridClassName="grid-cols-3"
                heading="Product"
                isOpen={activeMenu === "features"}
                items={featureMenuLinks}
                label="Features"
                onOpenChange={setActiveMenu}
                value="features"
              />
              <NavigationDropdown
                contentClassName="w-[38rem] max-w-[calc(100vw-3rem)]"
                gridClassName="grid-cols-2"
                isOpen={activeMenu === "use-cases"}
                items={primaryUseCaseMenuLinks}
                label="Use Cases"
                onOpenChange={setActiveMenu}
                value="use-cases"
              />
              <li>
                <DesktopNavItem
                  href="/ai-project-manager"
                  title="AI Project Manager"
                />
              </li>
              <NavigationDropdown
                contentClassName="right-0 left-auto w-[22rem]"
                gridClassName="grid-cols-1"
                isOpen={activeMenu === "resources"}
                items={resourceLinks}
                label="Resources"
                onOpenChange={setActiveMenu}
                value="resources"
              />
              <li>
                <DesktopNavItem href="/pricing" title="Pricing" />
              </li>
            </ul>
          </nav>
        </Flex>
        <Flex align="center" className="ml-4 gap-2">
          {/* <RequestDemo /> */}
          <Button
            className="hidden bg-white/60 px-5 text-[0.93rem] backdrop-blur-sm md:flex dark:bg-transparent"
            color="invert"
            href={APP_URL}
            rounded="md"
            size="lg"
            variant="outline"
          >
            Log in
          </Button>
          <Button
            className="px-5 text-[0.93rem]"
            color="invert"
            href={SIGNUP_URL}
            rounded="md"
            size="lg"
          >
            Get started
          </Button>

          <MobileNavigation />
        </Flex>
      </Container>
    </Box>
  );
};
