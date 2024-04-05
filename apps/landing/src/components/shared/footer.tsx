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
  className = "opacity-60",
}: {
  href: string;
  children: ReactNode;
  className?: string;
}) => (
  <Link
    className={cn("3xl:text-lg mb-6 block max-w-max", className)}
    href={href}
  >
    {children}
  </Link>
);

const Copyright = () => (
  <Box className="col-span-6">
    <Box className=" flex flex-col justify-between gap-y-8 border-b border-dark-200 pb-8 md:col-span-4 md:flex-row md:items-center md:gap-y-0">
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
  </Box>
);

export const Footer = () => {
  const quickLinks = [
    {
      href: "/projects",
      title: "Projects",
    },
    {
      href: "/services",
      title: "Services",
    },
    {
      href: "/about",
      title: "About",
    },
    {
      href: "/contact",
      title: "Contact",
    },
  ];
  const servicesLinks = [
    {
      href: "/services#websites",
      title: "Websites",
    },
    {
      href: "/services#web-apps",
      title: "Web Apps",
    },
    {
      href: "/services#mobile-apps",
      title: "Mobile Apps",
    },
    {
      href: "/services#ui-ux",
      title: "UI/UX Design",
    },

    {
      href: "/services#consultancy",
      title: "IT Consultancy",
    },
  ];
  return (
    <Box as="footer" className="relative bg-black">
      <Container className="grid grid-cols-6 gap-y-8 py-20">
        <Box className="col-span-6 md:col-span-1">
          <Logo className="h-7" />
          <Text color="muted">Complexus</Text>
        </Box>
        <Box className="col-span-3 mt-8 md:col-span-1 md:mt-0">
          <Text
            className="3xl:text-xl 3xl:tracking-widest mb-8 tracking-wider"
            fontWeight="medium"
            transform="uppercase"
          >
            Explore
          </Text>
          {quickLinks.map(({ href, title }) => (
            <FooterLink href={href} key={href}>
              {title}
            </FooterLink>
          ))}
        </Box>
        <Box className="col-span-3 mt-8 md:col-span-1 md:mt-0">
          <Text
            className="3xl:text-xl 3xl:tracking-widest mb-8 tracking-wider"
            fontWeight="medium"
            transform="uppercase"
          >
            Services
          </Text>
          {servicesLinks.map(({ href, title }) => (
            <FooterLink href={href} key={href}>
              {title}
            </FooterLink>
          ))}
        </Box>
        <Box className="col-span-6 md:col-span-1">
          <Text
            className="3xl:text-xl 3xl:tracking-widest mb-8 tracking-wider"
            fontWeight="medium"
            transform="uppercase"
          >
            Get in touch
          </Text>
          <FooterLink
            className="text-primary"
            href="mailto:hello@complexus.tech"
          >
            hello@complexus.tech
          </FooterLink>
          <FooterLink className="text-primary" href="tel:+263776686870">
            (+263) 776-686-870
          </FooterLink>
        </Box>
        test
        <div className="col-span-2" />
        <Copyright />
        <Text color="muted" fontSize="sm">
          Complexus is a product of Complexus Technologies. All rights reserved.
        </Text>
      </Container>
    </Box>
  );
};
