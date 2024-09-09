"use client";
import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Button, Flex, Text, Tooltip } from "ui";
import Link from "next/link";
import {
  FacebookIcon,
  InstagramIcon,
  LinkedInIcon,
  MoonIcon,
  SettingsIcon,
  SunIcon,
  SystemIcon,
  TwitterIcon,
} from "icons";
import { Logo } from "../ui/logo";
import { Container } from "../ui";
import { useTheme } from "next-themes";

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
  >
    {children}
  </Link>
);

const Copyright = () => {
  const { theme, setTheme } = useTheme();

  return (
    <Box className="flex flex-col justify-between gap-y-8 border-b border-gray-100 pb-4 dark:border-dark-200 md:flex-row md:items-center md:gap-y-0">
      <Box className="3xl:gap-16 flex gap-8">
        <Link
          className="hover:text-primary"
          href="https://twitter.com/complexus_tech"
          target="_blank"
        >
          <TwitterIcon className="h-5 w-auto" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.linkedin.com/company/complexus-tech/"
          target="_blank"
        >
          <LinkedInIcon className="h-5 w-auto" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.instagram.com/complexus_tech/"
          target="_blank"
        >
          <InstagramIcon className="h-5 w-auto" />
        </Link>
        <Link
          className="hover:text-primary"
          href="https://www.facebook.com/complexus.tech"
          target="_blank"
        >
          <FacebookIcon className="h-5 w-auto" />
        </Link>
      </Box>
      <div className="flex gap-0.5 rounded-full border border-gray-200 p-0.5 dark:border-dark-50">
        <Tooltip title="System">
          <button
            className={cn(
              "rounded-full border-0 bg-opacity-0 p-1 opacity-60 hover:opacity-100 dark:bg-opacity-0",
              {
                "border border-gray-200 border-opacity-100 opacity-100 dark:border dark:border-gray-300/40":
                  theme === "system",
              },
            )}
            onClick={() => setTheme("system")}
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
            onClick={() => setTheme("light")}
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
            onClick={() => setTheme("dark")}
          >
            <MoonIcon className="h-4 w-auto" />
          </button>
        </Tooltip>
      </div>
    </Box>
  );
};

export const Footer = () => {
  const product = [
    {
      href: "/product#changelog",
      title: "Changelog",
    },
    {
      title: "API",
      href: "/product#api",
    },
  ];
  const company = [
    {
      title: "About us",
      href: "/about",
    },
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
      title: "Blog",
      href: "/blog",
    },
  ];
  return (
    <Box
      as="footer"
      className="relative border-t border-gray-100 bg-white dark:border-dark-200 dark:bg-dark"
    >
      <Container className="grid grid-cols-2 gap-x-6 gap-y-8 pb-16 pt-12 md:grid-cols-6 md:pt-20">
        <Box className="col-span-2">
          <Logo className="-left-1 h-7" />
          <Text color="muted" className="ml-2.5 mt-2 w-10/12">
            Unlimited image conversion: formats, colors, and sizes at your
            fingertips.
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
      <Container className="pb-10">
        <Copyright />
        <Flex className="mt-4" justify="between" align="center">
          <Text color="muted" fontSize="sm">
            Complexus is a product of{" "}
            <a
              className="underline underline-offset-[3px] hover:text-primary"
              href="http://complexus.tech"
              rel="noopener noreferrer"
              target="_blank"
            >
              Complexus Technologies.
            </a>
          </Text>
          <Box className="hidden items-center gap-5 opacity-70 md:flex">
            <Link className="text-sm" href="/privacy">
              Privacy Policy
            </Link>
            |
            <Text fontSize="sm">
              © {new Date().getFullYear()} Complexus Technologies &bull; All
              rights reserved.
            </Text>
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};
