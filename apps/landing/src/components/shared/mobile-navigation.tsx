"use client";

import Link from "next/link";
import { useState } from "react";
import { Box, Dialog, Flex } from "ui";
import { cn } from "lib";
import { featureLinks } from "@/lib/feature-links";
import { primaryUseCaseLinks } from "@/lib/use-case-links";

const resourceLinks = [
  { label: "Docs", href: "https://docs.fortyone.app" },
  { label: "Blog", href: "/blog" },
  { label: "Pitch", href: "https://pitch.fortyone.app" },
];

const NavMenuButton = ({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) => {
  return (
    <button
      aria-expanded={open}
      aria-label={open ? "Close navigation menu" : "Open navigation menu"}
      className="flex aspect-square h-10 items-center justify-center"
      onClick={() => {
        setOpen(!open);
      }}
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
};

export const MobileNavigation = () => {
  const [open, setOpen] = useState(false);

  const navItems = [
    {
      label: "Features",
      items: featureLinks,
    },
    {
      label: "Use Cases",
      items: primaryUseCaseLinks,
    },
    {
      label: "Resources",
      items: resourceLinks,
    },
    { label: "Pricing", href: "/pricing" },
  ];

  return (
    <>
      <div className="landing-mobile-navigation">
        <NavMenuButton open={open} setOpen={setOpen} />
      </div>

      <Dialog onOpenChange={setOpen} open={open}>
        <Dialog.Content
          className="bg-background dark:bg-background m-0 mt-16 w-full rounded-none border-0 outline-none sm:mt-18 dark:border-0"
          hideClose
          overlayClassName="bg-transparent dark:bg-transparent"
        >
          <Dialog.Header className="sr-only">
            <Dialog.Title className="sr-only">Menu</Dialog.Title>
          </Dialog.Header>
          <Dialog.Description className="sr-only">
            Menu dialog
          </Dialog.Description>
          <Dialog.Body className="flex h-[calc(100dvh-4rem)] max-h-screen flex-col justify-between overflow-y-auto overscroll-contain px-4 pt-4 pb-8 sm:h-[calc(100dvh-4.5rem)] sm:pt-5">
            <Box>
              <Flex className="gap-4 pl-2 sm:gap-6" direction="column">
                {navItems.map(({ label, href, items }) => {
                  if (items) {
                    return (
                      <div key={label}>
                        <div className="mb-2 text-base font-medium sm:mb-3 sm:text-lg">
                          {label}
                        </div>
                        <Flex
                          className="gap-1 pl-3 sm:gap-3 sm:pl-4"
                          direction="column"
                        >
                          {items.map(({ label: itemLabel, href: itemHref }) => (
                            <Link
                              className="hover:bg-accent dark:hover:bg-surface-prominent -mx-2 flex min-h-10 items-center rounded px-2 py-1 text-[0.9375rem] opacity-80 transition-colors hover:opacity-100 sm:min-h-0 sm:rounded-md sm:py-1.5 sm:text-base"
                              href={itemHref}
                              key={itemLabel}
                              onClick={() => {
                                setOpen(false);
                              }}
                            >
                              {itemLabel}
                            </Link>
                          ))}
                        </Flex>
                      </div>
                    );
                  }

                  return href ? (
                    <Link
                      className="hover:bg-accent dark:hover:bg-surface-prominent -mx-2 flex min-h-10 items-center rounded px-2 py-1 text-base font-medium transition-colors sm:min-h-0 sm:rounded-md sm:py-1.5 sm:text-lg"
                      href={href}
                      key={label}
                      onClick={() => {
                        setOpen(false);
                      }}
                    >
                      {label}
                    </Link>
                  ) : null;
                })}
              </Flex>
            </Box>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
