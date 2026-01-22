"use client";

import Link from "next/link";
import { useState } from "react";
import { Box, Dialog, Flex } from "ui";
import { cn } from "lib";

const NavMenuButton = ({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) => {
  return (
    <button
      className="flex aspect-square h-10 items-center justify-center"
      onClick={() => {
        setOpen(!open);
      }}
      type="button"
    >
      <span>
        <span
          className={cn(
            "bg-dark mb-[0.4rem] block h-px w-5 transition duration-300 ease-in-out dark:bg-white",
            {
              "mb-0 rotate-45": open,
            },
          )}
        />
        <span
          className={cn(
            "bg-dark block h-px w-5 transition duration-300 ease-in-out dark:bg-white",
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
      label: "Product",
      items: [
        { label: "Tasks", href: "/features/tasks" },
        { label: "Objectives", href: "/features/objectives" },
        { label: "OKRs", href: "/features/okrs" },
        { label: "Sprints", href: "/features/sprints" },
      ],
    },
    { label: "Pricing", href: "/pricing" },
    { label: "Contact", href: "/contact" },
    { label: "Blog", href: "/blog" },
    { label: "Help Center", href: "https://docs.fortyone.app" },
    { label: "Pitch", href: "https://pitch.fortyone.app" },
  ];

  return (
    <>
      <div className="md:hidden">
        <NavMenuButton open={open} setOpen={setOpen} />
      </div>

      <Dialog onOpenChange={setOpen} open={open}>
        <Dialog.Content
          className="bg-surface/90 m-0 mt-16 w-full rounded-none border-0 outline-none dark:border-0"
          hideClose
          overlayClassName="bg-transparent dark:bg-transparent"
        >
          <Dialog.Header className="sr-only">
            <Dialog.Title className="sr-only">Menu</Dialog.Title>
          </Dialog.Header>
          <Dialog.Description className="sr-only">
            Menu dialog
          </Dialog.Description>
          <Dialog.Body className="flex h-[calc(100vh-4rem)] max-h-screen flex-col justify-between px-4 py-10">
            <Box>
              <Flex className="pl-2" direction="column" gap={7}>
                {navItems.map(({ label, href, items }) => {
                  if (items) {
                    return (
                      <div key={label}>
                        <div className="mb-4 text-3xl">{label}</div>
                        <Flex className="pl-5" direction="column" gap={5}>
                          {items.map(({ label: itemLabel, href: itemHref }) => (
                            <Link
                              className="text-2xl opacity-80"
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
                      className="text-3xl"
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
