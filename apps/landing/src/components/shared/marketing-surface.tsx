import type { ReactNode } from "react";
import Link from "next/link";
import { ArrowUpRightIcon } from "icons";
import type {
  NavigationIconName,
  NavigationIconTone,
} from "./navigation-menu-icon";
import { NavigationMenuIcon } from "./navigation-menu-icon";

export function MarketingHero({
  id,
  eyebrow,
  title,
  description,
  children,
}: {
  id: string;
  eyebrow: string;
  title: ReactNode;
  description: ReactNode;
  children?: ReactNode;
}) {
  return (
    <section aria-labelledby={id} className="pt-24">
      <div className="dark landing-light-contrast landing-footer-gradient landing-hero-shell landing-page-frame text-foreground rounded-2xl px-6 py-14 text-center sm:rounded-[3rem] md:rounded-[4rem] md:py-18">
        <p className="text-text-muted text-sm">{eyebrow}</p>
        <h1
          className="mx-auto mt-6 max-w-[22ch] text-5xl font-medium text-balance md:text-[3.5rem]"
          id={id}
        >
          {title}
        </h1>
        <p className="text-text-description mx-auto mt-6 max-w-2xl text-pretty">
          {description}
        </p>
        {children}
      </div>
    </section>
  );
}

export function MarketingResourceCard({
  title,
  description,
  href,
  icon,
  tone,
}: {
  title: string;
  description: string;
  href: string;
  icon: NavigationIconName;
  tone: NavigationIconTone;
}) {
  return (
    <Link
      className="group bg-surface-muted/60 hover:bg-surface-muted focus-visible:outline-ring flex h-full flex-col rounded-2xl p-6 transition-colors focus-visible:outline-2 focus-visible:outline-offset-4 md:rounded-[2rem] md:p-7"
      href={href}
    >
      <div className="flex items-center justify-between">
        <NavigationMenuIcon className="size-11" name={icon} tone={tone} />
        <span
          aria-hidden="true"
          className="bg-background/80 grid size-9 place-items-center rounded-lg"
        >
          <ArrowUpRightIcon className="text-foreground size-4 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5 motion-reduce:transform-none" />
        </span>
      </div>
      <h3 className="mt-5 text-2xl font-medium">{title}</h3>
      <p className="text-text-description mt-3 text-pretty">{description}</p>
    </Link>
  );
}
