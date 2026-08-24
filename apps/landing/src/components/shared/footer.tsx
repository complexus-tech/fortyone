"use client";

import type { ComponentType, ReactNode } from "react";
import { useSyncExternalStore } from "react";
import {
  FacebookIcon,
  InstagramIcon,
  LinkedinIcon,
  MoonIcon,
  SunIcon,
  SystemIcon,
  TwitterIcon,
} from "icons";
import { cn } from "lib";
import Link from "next/link";
import { useTheme } from "next-themes";
import { Tooltip } from "ui";
import { comparisons } from "@/lib/comparisons";
import { featureLinks } from "@/lib/feature-links";
import { useCaseLinks } from "@/lib/use-case-links";
import { Logo } from "../ui/logo";

const COPYRIGHT_YEAR = 2026;

const themeOptions = [
  { id: "light", label: "Light", icon: SunIcon },
  { id: "system", label: "System", icon: SystemIcon },
  { id: "dark", label: "Dark", icon: MoonIcon },
] as const;

const subscribeToHydration = () => () => undefined;
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;

const caseLinks = useCaseLinks.map(({ href, label }) => ({
  href,
  title: label,
}));

const footerFeatureLinks = featureLinks.map(({ href, label }) => ({
  href,
  title: label,
}));

const company = [
  { title: "Pricing", href: "/pricing" },
  { title: "Contact", href: "/contact" },
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
  { title: "Privacy Policy", href: "/privacy" },
  { title: "Terms of Service", href: "/terms" },
];

const resources = [
  { title: "Developers", href: "/developers" },
  { title: "Docs", href: "https://docs.fortyone.app" },
  { title: "Blog", href: "/blog" },
  { title: "Pitch", href: "https://pitch.fortyone.app" },
];

const integrations = [
  { title: "Google Calendar", href: "/integrations/google-calendar" },
  { title: "Slack", href: "/integrations/slack" },
  { title: "GitHub", href: "/integrations/github" },
];

const socialLinks = [
  {
    href: "https://x.com/fortyoneapp",
    label: "X",
    icon: TwitterIcon,
  },
  {
    href: "https://www.linkedin.com/company/complexus-app/",
    label: "LinkedIn",
    icon: LinkedinIcon,
  },
  {
    href: "https://www.instagram.com/complexus_tech/",
    label: "Instagram",
    icon: InstagramIcon,
  },
  {
    href: "https://www.facebook.com/complexus.tech",
    label: "Facebook",
    icon: FacebookIcon,
  },
] as const;

const FooterLink = ({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) => {
  const isExternal = href.startsWith("http");

  return (
    <Link
      className="text-text-muted hover:text-primary focus-visible:text-primary block max-w-max py-1 text-[0.9375rem] leading-6 transition-colors duration-200 ease-out focus-visible:outline-none"
      href={href}
      prefetch={false}
      rel={isExternal ? "noreferrer" : undefined}
      target={isExternal ? "_blank" : undefined}
    >
      {children}
    </Link>
  );
};

const FooterGroup = ({
  title,
  links,
}: {
  title: string;
  links: readonly { href: string; title: string }[];
}) => (
  <section>
    <h2 className="font-heading text-foreground mb-3 text-base font-semibold tracking-tight">
      {title}
    </h2>
    <nav aria-label={`${title} links`}>
      {links.map(({ href, title: linkTitle }) => (
        <FooterLink href={href} key={href}>
          {linkTitle}
        </FooterLink>
      ))}
    </nav>
  </section>
);

const SocialLink = ({
  href,
  icon: Icon,
  label,
}: {
  href: string;
  icon: ComponentType<{ "aria-hidden"?: boolean; className?: string }>;
  label: string;
}) => (
  <Tooltip title={label}>
    <Link
      aria-label={label}
      className="border-border bg-state-hover text-foreground hover:border-primary/50 hover:bg-primary/10 hover:text-primary focus-visible:outline-primary grid size-11 place-items-center rounded-xl border transition-[border-color,background-color,color,transform] duration-200 ease-out hover:-translate-y-0.5 focus-visible:outline-2 focus-visible:outline-offset-2"
      href={href}
      rel="noreferrer"
      target="_blank"
    >
      <Icon aria-hidden className="h-5 w-auto text-current" />
    </Link>
  </Tooltip>
);

export const Footer = () => {
  const { theme, setTheme } = useTheme();
  const hasMounted = useSyncExternalStore(
    subscribeToHydration,
    getClientSnapshot,
    getServerSnapshot,
  );

  const activeTheme = hasMounted ? theme ?? "system" : undefined;

  return (
    <footer className="bg-background py-2 sm:py-3 md:py-6">
      <div className="landing-hero-shell landing-page-frame overflow-hidden rounded-[3rem] md:rounded-[4rem]">
        <div className="px-6 py-10 sm:px-10 sm:py-12 lg:px-14 lg:py-16 xl:px-20 xl:py-20">
          <div className="grid grid-cols-2 gap-x-6 gap-y-14 lg:grid-cols-[minmax(18rem,1.45fr)_repeat(3,minmax(0,1fr))] lg:gap-x-10 xl:gap-x-16">
            <div className="col-span-2 flex flex-col justify-between gap-10 lg:col-span-1 lg:min-h-[34rem]">
              <p className="font-heading max-w-md text-[2.6rem] leading-[1.02] font-semibold tracking-[-0.045em] text-balance sm:text-5xl lg:text-[3.4rem]">
                Keep what matters moving.
              </p>
              <Logo className="text-foreground -left-1 h-7 sm:h-8" />
            </div>

            <div className="space-y-10">
              <FooterGroup links={footerFeatureLinks} title="Features" />
              <FooterGroup links={company} title="Company" />
            </div>

            <div className="space-y-10">
              <FooterGroup links={caseLinks} title="Use cases" />
              <FooterGroup links={integrations} title="Integrations" />
            </div>

            <div className="col-span-2 grid grid-cols-2 gap-x-6 gap-y-10 lg:col-span-1 lg:block lg:space-y-10">
              <FooterGroup links={resources} title="Resources" />
              <FooterGroup links={comparisonLinks} title="Compare" />
              <FooterGroup links={legal} title="Legal" />
            </div>
          </div>

          <div className="border-border mt-14 border-t pt-8 lg:mt-16 lg:pt-10">
            <div className="grid gap-8 lg:grid-cols-2 lg:items-end">
              <div>
                <p className="text-foreground mb-3 text-sm font-semibold">
                  Follow FortyOne
                </p>
                <div className="flex flex-wrap gap-2.5">
                  {socialLinks.map((socialLink) => (
                    <SocialLink {...socialLink} key={socialLink.href} />
                  ))}
                </div>
              </div>

              <div className="lg:justify-self-end">
                <p className="text-foreground mb-3 text-sm font-semibold lg:text-right">
                  Appearance
                </p>
                <div
                  aria-label="Color theme"
                  className="bg-state-hover flex w-max items-center gap-0.5 rounded-full p-1"
                  role="group"
                >
                  {themeOptions.map(({ icon: Icon, id, label }) => (
                    <Tooltip key={id} title={label}>
                      <button
                        aria-label={`Use ${label.toLowerCase()} theme`}
                        aria-pressed={activeTheme === id}
                        className={cn(
                          "text-text-muted hover:text-foreground focus-visible:outline-primary grid h-7 w-8 place-items-center rounded-full transition-[width,color,background-color,transform] duration-200 ease-out focus-visible:outline-2 focus-visible:outline-offset-1 active:scale-[0.94]",
                          {
                            "bg-state-active text-foreground w-10":
                              activeTheme === id,
                          },
                        )}
                        onClick={() => {
                          setTheme(id);
                        }}
                        type="button"
                      >
                        <Icon aria-hidden className="text-current" />
                      </button>
                    </Tooltip>
                  ))}
                </div>
              </div>
            </div>

            <div className="border-border text-text-muted mt-8 flex flex-wrap items-center justify-between gap-4 border-t pt-6 text-sm">
              <p>
                Built with heart in{" "}
                <span
                  aria-label="Zimbabwe"
                  className="text-foreground opacity-100"
                  role="img"
                >
                  🇿🇼
                </span>{" "}
                by{" "}
                <a
                  className="decoration-border-strong hover:text-primary underline decoration-dotted underline-offset-4 transition-colors"
                  href="https://complexus.tech"
                  rel="noreferrer"
                  target="_blank"
                >
                  Complexus
                </a>
              </p>
              <p>© {COPYRIGHT_YEAR} Complexus LLC · All rights reserved.</p>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
};
