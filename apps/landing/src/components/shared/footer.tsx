"use client";
import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Flex, Text, Tooltip } from "ui";
import Link from "next/link";
import {
  FacebookIcon,
  InstagramIcon,
  LinkedinIcon,
  MoonIcon,
  SunIcon,
  SystemIcon,
  TwitterIcon,
} from "icons";
import { useTheme } from "next-themes";
import { comparisons } from "@/lib/comparisons";
import { featureLinks } from "@/lib/feature-links";
import { useCaseLinks } from "@/lib/use-case-links";
import { Logo } from "../ui/logo";
import { Container } from "../ui/container";

const COPYRIGHT_YEAR = 2026;

const themeOptions = [
  { id: "system", label: "System", icon: SystemIcon },
  { id: "light", label: "Light", icon: SunIcon },
  { id: "dark", label: "Dark", icon: MoonIcon },
] as const;

const caseLinks = useCaseLinks.map(({ href, label }) => ({
  href,
  title: label,
}));

const footerFeatureLinks = featureLinks.map(({ href, label }) => ({
  href,
  title: label,
}));

const company = [
  {
    title: "Pricing",
    href: "/pricing",
  },
  {
    title: "Contact",
    href: "/contact",
  },
];

const comparisonFooterOrder = ["asana", "jira", "clickup", "trello", "monday"];

const comparisonLinks = comparisonFooterOrder
  .map((slug) => comparisons.find((comparison) => comparison.slug === slug))
  .filter((comparison) => comparison !== undefined)
  .map(({ competitor, slug }) => ({
    href: `/compare/${slug}`,
    title: competitor,
  }));

const legal = [
  {
    title: "Privacy Policy",
    href: "/privacy",
  },
  {
    title: "Terms of Service",
    href: "/terms",
  },
];

const resources = [
  {
    title: "AI Project Manager",
    href: "/ai-project-manager",
  },
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

const FooterLink = ({
  href,
  children,
  className = "",
}: {
  href: string;
  children: ReactNode;
  className?: string;
}) => (
  <Link
    className={cn(
      "hover:text-primary mb-3 block max-w-max text-[0.9375rem] transition-colors duration-200 ease-in-out",
      className,
    )}
    href={href}
    target={href.startsWith("http") ? "_blank" : undefined}
  >
    {children}
  </Link>
);

const Copyright = () => {
  return (
    <Box className="border-border flex flex-col justify-between gap-y-8 border-b pb-4 md:flex-row md:items-center md:gap-y-0">
      <Box className="3xl:gap-16 flex gap-8">
        <Link
          className="hover:text-primary"
          href="https://x.com/fortyoneapp"
          target="_blank"
        >
          <span className="sr-only">Twitter</span>
          <TwitterIcon className="text-foreground" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.linkedin.com/company/complexus-app/"
          target="_blank"
        >
          <span className="sr-only">LinkedIn</span>
          <LinkedinIcon className="text-foreground" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.instagram.com/complexus_tech/"
          target="_blank"
        >
          <span className="sr-only">Instagram</span>
          <InstagramIcon className="text-foreground" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.facebook.com/complexus.tech"
          target="_blank"
        >
          <span className="sr-only">Facebook</span>
          <FacebookIcon className="text-foreground" />
        </Link>
      </Box>
      <Box className="hidden items-center gap-5 opacity-70 md:flex">
        <Link className="3xl:text-lg text-sm" href="/privacy">
          Privacy Policy
        </Link>
        |
        <Text fontSize="sm">
          © {COPYRIGHT_YEAR} Complexus LLC &bull; All rights reserved.
        </Text>
      </Box>
    </Box>
  );
};

export const Footer = () => {
  const { theme, setTheme } = useTheme();
  return (
    <Box as="footer" className="bg-surface-muted dark:bg-surface relative">
      <Container>
        <Box className="mb-8 grid grid-cols-2 gap-x-6 gap-y-8 py-12 md:grid-cols-6 md:pt-20">
          <Box className="hidden md:block">
            <Logo className="-left-1 h-8 md:-left-2 md:h-7" />
          </Box>
          <Box>
            <Text className="text-text-muted mb-4">Features</Text>
            {footerFeatureLinks.map(({ href, title }) => (
              <FooterLink className="whitespace-nowrap" href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
          <Box>
            <Text className="text-text-muted mb-4">Use cases</Text>
            {caseLinks.map(({ href, title }) => (
              <FooterLink className="whitespace-nowrap" href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
          <Box>
            <Text className="text-text-muted mb-4">Company</Text>
            {company.map(({ href, title }) => (
              <FooterLink href={href} key={href}>
                {title}
              </FooterLink>
            ))}
            <Box className="mt-8">
              <Text className="text-text-muted mb-4">Compare</Text>
              {comparisonLinks.map(({ href, title }) => (
                <FooterLink href={href} key={href}>
                  {title}
                </FooterLink>
              ))}
            </Box>
          </Box>
          <Box>
            <Text className="text-text-muted mb-4">Resources</Text>
            {resources.map(({ href, title }) => (
              <FooterLink href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
          <Box>
            <Text className="text-text-muted mb-4">Legal</Text>
            {legal.map(({ href, title }) => (
              <FooterLink href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
        </Box>
      </Container>
      <Container className="pb-8 md:pb-16">
        <Copyright />
        <Flex className="mt-6" justify="between">
          <Text color="muted" fontSize="sm">
            Product of{" "}
            <a className="text-inherit underline" href="https://complexus.tech">
              Complexus
            </a>
            .
          </Text>
          <div
            aria-label="Color theme"
            className="bg-state-hover text-text-muted flex shrink-0 items-center gap-0.5 rounded-full p-1"
            role="group"
          >
            {themeOptions.map(({ icon: Icon, id, label }) => (
              <Tooltip key={id} title={label}>
                <button
                  aria-label={`Use ${label.toLowerCase()} theme`}
                  aria-pressed={theme === id}
                  className={cn(
                    "hover:text-foreground focus-visible:outline-foreground relative grid h-7 w-8 place-items-center rounded-full transition-[width,color,background-color,transform] duration-200 ease-out focus-visible:outline-2 focus-visible:outline-offset-1 active:scale-[0.94]",
                    {
                      "bg-state-active text-foreground w-10": theme === id,
                    },
                  )}
                  onClick={() => {
                    setTheme(id);
                  }}
                  type="button"
                >
                  <Icon aria-hidden="true" className="size-[15px]" />
                </button>
              </Tooltip>
            ))}
          </div>
        </Flex>
      </Container>
    </Box>
  );
};
