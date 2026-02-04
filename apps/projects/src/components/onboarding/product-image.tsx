/* eslint-disable no-nested-ternary -- ok to nest ternary */
"use client";

import { Box, Text, BlurImage, Container } from "ui";
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
        src="/images/mesh.webp"
      />
      <Container className="absolute inset-0 z-10 flex flex-col justify-between pt-12 pb-16 md:px-16 dark:text-black">
        <Link href="/">
          <Logo className="h-8 dark:text-black" />
        </Link>
        <Box>
          <Text
            as="h2"
            className="mb-4 font-mono font-semibold tracking-wider uppercase antialiased dark:text-black"
          >
            [Built for builders]
          </Text>
          <Text
            as="h3"
            className="mb-10 text-5xl leading-[1.2] font-semibold dark:text-black"
          >
            Plan, track, deliver with the project management tool your team will
            love.
          </Text>
          <Box className="3xl:gap-16 flex gap-8">
            <Link
              aria-label="Follow FortyOne on X"
              href="https://x.com/fortyoneapp"
              rel="noreferrer"
              target="_blank"
            >
              <TwitterIcon className="text-foreground h-6 dark:text-black dark:hover:text-white" />
            </Link>
            <Link
              aria-label="Follow Complexus on LinkedIn"
              href="https://www.linkedin.com/company/complexus-tech/"
              rel="noreferrer"
              target="_blank"
            >
              <LinkedinIcon className="text-foreground h-6 dark:text-black dark:hover:text-white" />
            </Link>
            <Link
              aria-label="Follow Complexus on Instagram"
              href="https://www.instagram.com/complexus_tech/"
              rel="noreferrer"
              target="_blank"
            >
              <InstagramIcon className="text-foreground h-6 dark:text-black dark:hover:text-white" />
            </Link>
            <Link
              aria-label="Follow Complexus on Facebook"
              href="https://www.facebook.com/complexus.tech"
              rel="noreferrer"
              target="_blank"
            >
              <FacebookIcon className="text-foreground h-6 dark:text-black dark:hover:text-white" />
            </Link>
          </Box>
          <Text className="mt-10 text-lg dark:text-black">
            By{" "}
            {pathname?.includes("signup")
              ? "signing up"
              : pathname?.includes("login")
                ? "signing in"
                : "continuing"}
            , you agree to our{" "}
            <Link
              className="text-black underline"
              href="https://www.fortyone.app/terms"
            >
              Terms of Service
            </Link>{" "}
            and{" "}
            <Link
              className="text-black underline"
              href="https://www.fortyone.app/privacy"
            >
              Privacy Policy
            </Link>
          </Text>
        </Box>
      </Container>
    </Box>
  );
};
