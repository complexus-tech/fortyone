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
import { Logo } from "../ui/logo";
import { Container } from "../ui";

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
      "3xl:text-lg mb-3 block max-w-max opacity-80 transition-opacity duration-200 ease-in-out hover:text-primary hover:opacity-80 dark:opacity-60",
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
    <Box className="flex flex-col justify-between gap-y-8 border-b border-gray-100 pb-4 dark:border-dark-100 md:flex-row md:items-center md:gap-y-0">
      <Box className="3xl:gap-16 flex gap-8">
        <Link
          className="hover:text-primary"
          href="https://x.com/fortyoneapp"
          target="_blank"
        >
          <span className="sr-only">Twitter</span>
          <TwitterIcon className="text-dark dark:text-gray-200" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.linkedin.com/company/complexus-app/"
          target="_blank"
        >
          <span className="sr-only">LinkedIn</span>
          <LinkedinIcon className="text-dark dark:text-gray-200" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.instagram.com/complexus_tech/"
          target="_blank"
        >
          <span className="sr-only">Instagram</span>
          <InstagramIcon className="text-dark dark:text-gray-200" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.facebook.com/complexus.tech"
          target="_blank"
        >
          <span className="sr-only">Facebook</span>
          <FacebookIcon className="text-dark dark:text-gray-200" />
        </Link>
      </Box>
      <Box className="hidden items-center gap-5 opacity-70 md:flex">
        <Link className="3xl:text-lg text-sm" href="/privacy">
          Privacy Policy
        </Link>
        |
        <Text fontSize="sm">
          © {new Date().getFullYear()} FortyOne LLC &bull; All rights reserved.
        </Text>
      </Box>
    </Box>
  );
};

export const Footer = () => {
  const { theme, setTheme } = useTheme();
  const product = [
    {
      href: "/product/stories",
      title: "Stories",
    },
    {
      title: "Objectives",
      href: "/product/objectives",
    },
    {
      href: "/product/okrs",
      title: "OKRs",
    },

    {
      href: "/product/sprints",
      title: "Sprints",
    },
  ];
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
      title: "Help Center",
      href: "https://docs.fortyone.app",
    },
    {
      title: "Blog",
      href: "/blog",
    },
  ];
  return (
    <Box as="footer" className="relative">
      <Container>
        <Box className="mb-8 grid grid-cols-2 gap-x-6 gap-y-8 py-12 md:grid-cols-6 md:pt-20">
          <Box className="col-span-2">
            <Logo className="-left-1 h-8 md:-left-2 md:h-7" />
          </Box>
          <Box>
            <Text className="mb-4" fontSize="lg" fontWeight="semibold">
              Product
            </Text>
            {product.map(({ href, title }) => (
              <FooterLink href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
          <Box>
            <Text className="mb-4" fontSize="lg" fontWeight="semibold">
              Company
            </Text>
            {company.map(({ href, title }) => (
              <FooterLink href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
          <Box>
            <Text className="mb-4" fontSize="lg" fontWeight="semibold">
              Legal
            </Text>
            {legal.map(({ href, title }) => (
              <FooterLink href={href} key={href}>
                {title}
              </FooterLink>
            ))}
          </Box>
          <Box>
            <Text className="mb-4" fontSize="lg" fontWeight="semibold">
              Resources
            </Text>
            {resources.map(({ href, title }) => (
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
            FortyOne is a product of FortyOne LLC.
          </Text>
          <Flex className="flex gap-5">
            <Tooltip title="System">
              <button
                onClick={() => {
                  setTheme("system");
                }}
                type="button"
              >
                <SystemIcon
                  className={cn("h-4", {
                    "text-dark dark:text-white": theme === "system",
                  })}
                />
              </button>
            </Tooltip>
            <Tooltip title="Light">
              <button
                onClick={() => {
                  setTheme("light");
                }}
                type="button"
              >
                <SunIcon
                  className={cn("h-4", {
                    "text-dark dark:text-white": theme === "light",
                  })}
                />
              </button>
            </Tooltip>
            <Tooltip title="Dark">
              <button
                className={cn("", {
                  "text-dark dark:text-white": theme === "dark",
                })}
                onClick={() => {
                  setTheme("dark");
                }}
                type="button"
              >
                <MoonIcon
                  className={cn("h-4", {
                    "text-dark dark:text-white": theme === "dark",
                  })}
                />
              </button>
            </Tooltip>
          </Flex>
        </Flex>
      </Container>
    </Box>
  );
};
