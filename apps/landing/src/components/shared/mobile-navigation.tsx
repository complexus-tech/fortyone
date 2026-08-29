"use client";

import type { ComponentPropsWithoutRef } from "react";
import Link from "next/link";
import { forwardRef, useId, useState } from "react";
import { ArrowDown2Icon } from "icons";
import { Popover } from "ui";
import { cn } from "lib";
import { featureLinks } from "@/lib/feature-links";
import { primaryUseCaseLinks } from "@/lib/use-case-links";

const resourceLinks = [
  { label: "Docs", href: "https://docs.fortyone.app" },
  {
    label: "API Reference",
    href: "https://docs.fortyone.app/api-reference",
  },
  { label: "Blog", href: "/blog" },
  { label: "Pitch", href: "https://pitch.fortyone.app" },
];

const mobileNavigationGroups = [
  {
    label: "Features",
    items: featureLinks,
    value: "features",
  },
  {
    label: "Use Cases",
    items: primaryUseCaseLinks,
    value: "use-cases",
  },
  {
    label: "Resources",
    items: resourceLinks,
    value: "resources",
  },
] as const;

const isExternalLink = (href: string) => href.startsWith("http");

type NavMenuButtonProps = Omit<
  ComponentPropsWithoutRef<"button">,
  "children"
> & {
  menuId: string;
  open: boolean;
};

const NavMenuButton = forwardRef<HTMLButtonElement, NavMenuButtonProps>(
  ({ className, menuId, open, ...props }, ref) => {
    return (
      <button
        {...props}
        aria-controls={menuId}
        aria-expanded={open}
        aria-label={open ? "Close navigation menu" : "Open navigation menu"}
        className={cn(
          "focus-visible:outline-primary flex aspect-square h-10 items-center justify-center rounded-md outline-none focus-visible:outline-2 focus-visible:outline-offset-2",
          className,
        )}
        ref={ref}
        type="button"
      >
        <span>
          <span
            className={cn(
              "bg-foreground mb-[0.4rem] block h-px w-5 transition duration-300 ease-in-out",
              {
                "mb-0 rotate-45": open,
              },
            )}
          />
          <span
            className={cn(
              "bg-foreground block h-px w-5 transition duration-300 ease-in-out",
              {
                "-translate-y-[0.05rem] -rotate-45": open,
              },
            )}
          />
        </span>
      </button>
    );
  },
);

NavMenuButton.displayName = "NavMenuButton";

export const MobileNavigation = () => {
  const instanceId = useId().replaceAll(":", "");
  const [expandedGroup, setExpandedGroup] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const menuId = `${instanceId}-mobile-navigation`;

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen);

    if (!nextOpen) setExpandedGroup(null);
  };

  const closeMenu = () => {
    handleOpenChange(false);
  };

  return (
    <div className="landing-mobile-navigation">
      <Popover onOpenChange={handleOpenChange} open={open}>
        <Popover.Trigger asChild>
          <NavMenuButton menuId={menuId} open={open} />
        </Popover.Trigger>

        <Popover.Content
          align="end"
          aria-label="Mobile navigation"
          className="landing-hero-shell border-border/50 dark:border-border/60 m-0 max-h-[calc(100dvh-6rem)] w-[calc(100vw-1rem)] max-w-none overflow-hidden rounded-2xl border bg-transparent p-0 shadow-[0_24px_64px_-24px_rgba(31,24,18,0.48)] outline-none sm:w-[calc(100vw-1.5rem)] sm:rounded-[3rem] md:w-[calc(100vw-3rem)] md:rounded-[4rem] lg:hidden dark:bg-transparent dark:shadow-[0_24px_64px_-24px_rgba(0,0,0,0.82)]"
          collisionPadding={8}
          id={menuId}
          side="bottom"
          sideOffset={12}
        >
          <div className="max-h-[calc(100dvh-6rem)] overflow-y-auto overscroll-contain px-5 py-4 sm:px-7 sm:py-6">
            <nav aria-label="Mobile">
              <ul className="divide-border/70 divide-y">
                {mobileNavigationGroups.map(({ items, label, value }) => {
                  const isExpanded = expandedGroup === value;
                  const buttonId = `${menuId}-${value}-trigger`;
                  const panelId = `${menuId}-${value}-panel`;

                  return (
                    <li key={value}>
                      <button
                        aria-controls={panelId}
                        aria-expanded={isExpanded}
                        className="focus-visible:outline-primary flex min-h-12 w-full items-center justify-between gap-4 rounded-md px-2 text-left text-base font-semibold outline-none focus-visible:outline-2 focus-visible:outline-offset-[-2px]"
                        id={buttonId}
                        onClick={() => {
                          setExpandedGroup(isExpanded ? null : value);
                        }}
                        type="button"
                      >
                        {label}
                        <ArrowDown2Icon
                          aria-hidden="true"
                          className={cn(
                            "text-text-muted h-4 w-auto transition-transform duration-200 motion-reduce:transition-none",
                            isExpanded && "rotate-180",
                          )}
                          strokeWidth={2.5}
                        />
                      </button>

                      {isExpanded ? (
                        <div
                          aria-labelledby={buttonId}
                          className="grid gap-0.5 px-2 pb-3"
                          id={panelId}
                          role="region"
                        >
                          {items.map(({ label: itemLabel, href: itemHref }) => (
                            <Link
                              className="text-text-muted hover:bg-background/55 hover:text-foreground focus-visible:outline-primary dark:hover:bg-surface-prominent/65 flex min-h-9 items-center rounded-md px-3 text-[0.9375rem] transition-colors outline-none focus-visible:outline-2 focus-visible:outline-offset-[-2px]"
                              href={itemHref}
                              key={itemLabel}
                              onClick={closeMenu}
                              prefetch={!isExternalLink(itemHref)}
                              rel={
                                isExternalLink(itemHref)
                                  ? "noreferrer"
                                  : undefined
                              }
                              target={
                                isExternalLink(itemHref) ? "_blank" : undefined
                              }
                            >
                              {itemLabel}
                            </Link>
                          ))}
                        </div>
                      ) : null}
                    </li>
                  );
                })}
                <li>
                  <Link
                    className="focus-visible:outline-primary flex min-h-12 items-center rounded-md px-2 text-base font-semibold outline-none focus-visible:outline-2 focus-visible:outline-offset-[-2px]"
                    href="/pricing"
                    onClick={closeMenu}
                    prefetch
                  >
                    Pricing
                  </Link>
                </li>
              </ul>
            </nav>
          </div>
        </Popover.Content>
      </Popover>
    </div>
  );
};
