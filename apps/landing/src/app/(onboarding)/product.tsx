/* eslint-disable no-nested-ternary -- ok to nest ternary */
"use client";

import { Box, Text, BlurImage, Container, Flex } from "ui";
import Link from "next/link";
import { FacebookIcon, InstagramIcon, LinkedinIcon, TwitterIcon } from "icons";
import { usePathname } from "next/navigation";
import { Logo } from "@/components/ui";

export const ProductImage = () => {
  const pathname = usePathname();

  return (
    <Box className="relative hidden p-4 md:block">
      <BlurImage
        alt="Login"
        className="h-full w-full rounded-2xl object-cover"
        quality={100}
        src="/images/mesh.webp"
      />
      <Container className="absolute inset-0 z-10 flex flex-col justify-between pb-16 pt-12 dark:text-black md:px-16">
        <Logo className="h-7 dark:text-black" />
        <Box>
          <Text className="mb-4 font-mono font-semibold uppercase dark:text-black">
            Built for builders
          </Text>
          <Text
            as="h3"
            className="mb-10 text-5xl font-semibold leading-[1.2] dark:text-black"
          >
            Plan, track, deliver with the project management tool your team will
            love.
          </Text>
          <Box className="3xl:gap-16 flex gap-8">
            <Link href="https://x.com/fortyoneapp" target="_blank">
              <TwitterIcon className="text-dark dark:text-black dark:hover:text-primary" />
            </Link>
            <Link
              href="https://www.linkedin.com/company/complexus-tech/"
              target="_blank"
            >
              <LinkedinIcon className="text-dark dark:text-black dark:hover:text-primary" />
            </Link>
            <Link
              href="https://www.instagram.com/complexus_tech/"
              target="_blank"
            >
              <InstagramIcon className="text-dark dark:text-black dark:hover:text-primary" />
            </Link>
            <Link
              href="https://www.facebook.com/complexus.tech"
              target="_blank"
            >
              <FacebookIcon className="text-dark dark:text-black dark:hover:text-primary" />
            </Link>
          </Box>
          <Text className="mt-10 text-[0.95rem] dark:text-black">
            By{" "}
            {pathname?.includes("signup")
              ? "signing up"
              : pathname?.includes("login")
                ? "signing in"
                : "continuing"}
            , you agree to our{" "}
            <Link className="text-dark underline dark:text-black" href="/terms">
              Terms of Service
            </Link>{" "}
            and{" "}
            <Link
              className="text-dark underline dark:text-black"
              href="/privacy"
            >
              Privacy Policy
            </Link>
          </Text>
        </Box>
      </Container>
    </Box>
  );
};
