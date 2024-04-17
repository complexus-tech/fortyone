"use client";
import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Text } from "ui";
import Link from "next/link";
import { FacebookIcon, InstagramIcon, LinkedInIcon, TwitterIcon } from "icons";
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
  >
    {children}
  </Link>
);

const Copyright = () => (
  <Box className="flex flex-col justify-between gap-y-8 border-b border-gray-200 pb-4 dark:border-dark-200 md:flex-row md:items-center md:gap-y-0">
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
    <Box className="hidden items-center gap-5 opacity-70 md:flex">
      <Link className="3xl:text-lg text-sm" href="/">
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
      href: "/product#roadmaps",
      title: "Roadmaps",
    },

    {
      href: "/product#sprints",
      title: "Sprints",
    },
    {
      href: "/product#epics",
      title: "Epics",
    },
    {
      href: "/product#documents",
      title: "Documents",
    },
    {
      href: "/product#reporting",
      title: "Reporting",
    },
    {
      href: "/product#discussions",
      title: "Discussions",
    },
    {
      href: "/product#whiteboards",
      title: "Whiteboards",
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
      href: "/legal/privacy",
    },
    {
      title: "Terms of Service",
      href: "/legal/terms",
    },
  ];

  const resources = [
    {
      title: "Help Center",
      href: "/resources/help",
    },
    {
      title: "Developers",
      href: "/resources/developers",
    },
    {
      title: "Changelog",
      href: "/resources/changelog",
    },
    {
      title: "Guides",
      href: "/resources/guides",
    },

    {
      title: "Status",
      href: "/resources/status",
    },
  ];
  return (
    <Box as="footer" className="relative bg-white dark:bg-black">
      <Container className="grid grid-cols-2 gap-x-6 gap-y-8 pb-12 pt-12 md:grid-cols-6 md:pt-20">
        <Box className="col-span-2">
          <Logo className="-left-1 h-7" />
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
        <Text className="mt-4" color="muted" fontSize="sm">
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
