"use client";
import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Text } from "ui";
import Link from "next/link";
import { FacebookIcon, InstagramIcon, LinkedinIcon, TwitterIcon } from "icons";
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
      "3xl:text-lg mb-6 block max-w-max opacity-80 transition-opacity duration-200 ease-in-out hover:text-primary hover:opacity-80 dark:opacity-60",
      className,
    )}
    href={href}
  >
    {children}
  </Link>
);

const Copyright = () => (
  <Box className="flex flex-col justify-between gap-y-8 border-b border-gray-200 pb-4 dark:border-dark-200 md:flex-row md:items-center md:gap-y-0">
    <Box className="3xl:gap-16 flex gap-8">
      <Link
        className="hover:text-primary"
        href="https://x.com/complexus_app"
        target="_blank"
      >
        <TwitterIcon className="h-5 w-auto" />
      </Link>
      <Link
        className="hover:text-primary"
        href="https://www.linkedin.com/company/complexus-tech/"
        target="_blank"
      >
        <LinkedinIcon className="h-5 w-auto" />
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
    <Box className="hidden items-center gap-5 opacity-70 md:flex">
      <Link className="3xl:text-lg text-sm" href="/privacy">
        Privacy Policy
      </Link>
      |
      <Text fontSize="sm">
        © {new Date().getFullYear()} Complexus Technologies &bull; All rights
        reserved.
      </Text>
    </Box>
  </Box>
);

export const Footer = () => {
  const product = [
    {
      href: "/product#stories",
      title: "Stories",
    },
    {
      title: "Objectives",
      href: "/product#objectives",
    },
    {
      href: "/product#okrs",
      title: "OKRs",
    },

    {
      href: "/product#sprints",
      title: "Sprints",
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
      title: "Developers",
      href: "/developers",
    },
    {
      title: "Blog",
      href: "/blog",
    },
  ];
  return (
    <Box as="footer" className="relative bg-white dark:bg-black">
      <Container className="grid grid-cols-2 gap-x-6 gap-y-8 pb-12 pt-12 md:grid-cols-6 md:pt-28">
        <Box className="col-span-2">
          <Logo className="-left-4 h-7" />
          <Text className="mt-2">
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
      <Container className="pb-16">
        <Copyright />
        <Text className="mt-6" color="muted" fontSize="sm">
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
      </Container>
    </Box>
  );
};
