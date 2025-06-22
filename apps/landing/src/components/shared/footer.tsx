"use client";
import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Text, Tooltip } from "ui";
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
      "3xl:text-lg mb-4 block max-w-max opacity-80 transition-opacity duration-200 ease-in-out hover:text-primary hover:opacity-80 dark:opacity-60",
      className,
    )}
    href={href}
    target={href.startsWith("http") ? "_blank" : undefined}
  >
    {children}
  </Link>
);

const Copyright = () => {
  const { theme, setTheme } = useTheme();
  return (
    <Box className="flex flex-col justify-between gap-y-8 border-b border-gray-200 pb-4 dark:border-dark-200 md:flex-row md:items-center md:gap-y-0">
      <Box className="3xl:gap-16 flex gap-8">
        <Link
          className="hover:text-primary"
          href="https://x.com/complexus_app"
          target="_blank"
        >
          <span className="sr-only">Twitter</span>
          <TwitterIcon className="h-5 w-auto" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.linkedin.com/company/complexus-app/"
          target="_blank"
        >
          <span className="sr-only">LinkedIn</span>
          <LinkedinIcon className="h-5 w-auto" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.instagram.com/complexus_tech/"
          target="_blank"
        >
          <span className="sr-only">Instagram</span>
          <InstagramIcon className="h-5 w-auto" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.facebook.com/complexus.tech"
          target="_blank"
        >
          <span className="sr-only">Facebook</span>
          <FacebookIcon className="h-5 w-auto" />
        </Link>
      </Box>
      <Box className="hidden items-center gap-5 opacity-70 md:flex">
        <Link className="3xl:text-lg text-sm" href="/privacy">
          Privacy Policy
        </Link>
        |
        <Text fontSize="sm">
          © {new Date().getFullYear()} Complexus LLC &bull; All rights
          reserved.
        </Text>
        <div className="flex gap-0.5 rounded-full border border-gray-200 p-1 dark:border-dark-50">
          <Tooltip title="System">
            <button
              className={cn(
                "rounded-full border-0 bg-opacity-0 p-1 opacity-60 hover:opacity-100 dark:bg-opacity-0",
                {
                  "border border-gray-200 border-opacity-100 opacity-100 dark:border dark:border-gray-300/40":
                    theme === "system",
                },
              )}
              onClick={() => {
                setTheme("system");
              }}
              type="button"
            >
              <SystemIcon className="h-4 w-auto" />
            </button>
          </Tooltip>
          <Tooltip title="Light">
            <button
              className={cn(
                "rounded-full border-0 p-1 opacity-60 hover:opacity-100",
                {
                  "border border-gray-200 border-opacity-100 opacity-100 dark:border dark:border-gray-300/40":
                    theme === "light",
                },
              )}
              onClick={() => {
                setTheme("light");
              }}
              type="button"
            >
              <SunIcon className="h-4 w-auto" />
            </button>
          </Tooltip>
          <Tooltip title="Dark">
            <button
              className={cn(
                "rounded-full border-0 p-1 opacity-60 hover:opacity-100",
                {
                  "border border-gray-200 border-opacity-100 opacity-100 dark:border dark:border-gray-300/40":
                    theme === "dark",
                },
              )}
              onClick={() => {
                setTheme("dark");
              }}
              type="button"
            >
              <MoonIcon className="h-4 w-auto" />
            </button>
          </Tooltip>
        </div>
      </Box>
    </Box>
  );
};

export const Footer = () => {
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
      title: "Documentation",
      href: "https://docs.complexus.app",
    },
    {
      title: "Blog",
      href: "/blog",
    },
  ];
  return (
    <Box as="footer" className="relative">
      <Container className="grid grid-cols-2 gap-x-6 gap-y-8 pb-12 pt-12 md:grid-cols-6 md:pt-28">
        <Box className="col-span-2">
          <Logo className="-left-1 h-8 md:-left-4 md:h-7" />
          <Text className="mt-2 w-11/12">
            Project management software that adapts to your team&apos;s
            workflow, not the other way around.
          </Text>
        </Box>
        <Box>
          <Text
            className="mb-6 tracking-wide"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Product
          </Text>
          {product.map(({ href, title }) => (
            <FooterLink href={href} key={href}>
              {title}
            </FooterLink>
          ))}
        </Box>
        <Box>
          <Text
            className="mb-6 tracking-wide"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Company
          </Text>
          {company.map(({ href, title }) => (
            <FooterLink href={href} key={href}>
              {title}
            </FooterLink>
          ))}
        </Box>
        <Box>
          <Text
            className="mb-6 tracking-wide"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Legal
          </Text>
          {legal.map(({ href, title }) => (
            <FooterLink href={href} key={href}>
              {title}
            </FooterLink>
          ))}
        </Box>
        <Box>
          <Text
            className="mb-6 tracking-wide"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Resources
          </Text>
          {resources.map(({ href, title }) => (
            <FooterLink href={href} key={href}>
              {title}
            </FooterLink>
          ))}
        </Box>
      </Container>
      <Container className="pb-8 md:pb-16">
        <Copyright />
        <Text className="mt-6" color="muted" fontSize="sm">
          Complexus is a product of Complexus LLC.
        </Text>
      </Container>
    </Box>
  );
};
