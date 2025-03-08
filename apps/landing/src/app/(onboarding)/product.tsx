/* eslint-disable no-nested-ternary -- ok to nest ternary */
"use client";

import { Box, Text, BlurImage, Container } from "ui";
import Link from "next/link";
import { FacebookIcon, InstagramIcon, LinkedinIcon, TwitterIcon } from "icons";
import { usePathname } from "next/navigation";
import { Blur } from "@/components/ui";

export const ProductImage = () => {
  const pathname = usePathname();

  return (
    <Box className="relative hidden md:block">
      <Blur className="absolute -top-96 left-1/2 right-1/2 z-[1] h-[300px] w-[300px] -translate-x-1/2 bg-warning/[0.07] md:h-[700px] md:w-[90vw]" />
      <BlurImage
        alt="Login"
        className="h-full w-full object-cover opacity-80"
        quality={100}
        src="/images/login.webp"
      />
      <Container className="absolute inset-0 z-10 flex flex-col items-center justify-end py-28">
        <Text align="center" className="mb-8 opacity-80" color="muted">
          Connect with us
        </Text>
        <Box className="3xl:gap-16 flex gap-8">
          <Link href="https://x.com/complexus_app" target="_blank">
            <TwitterIcon className="dark:hover:text-primary" />
          </Link>
          <Link
            href="https://www.linkedin.com/company/complexus-tech/"
            target="_blank"
          >
            <LinkedinIcon className="dark:hover:text-primary" />
          </Link>
          <Link
            href="https://www.instagram.com/complexus_tech/"
            target="_blank"
          >
            <InstagramIcon className="dark:hover:text-primary" />
          </Link>
          <Link href="https://www.facebook.com/complexus.tech" target="_blank">
            <FacebookIcon className="dark:hover:text-primary" />
          </Link>
        </Box>
        <Text className="mt-8 text-[0.95rem] opacity-90" color="muted">
          By{" "}
          {pathname?.includes("signup")
            ? "signing up"
            : pathname?.includes("login")
              ? "signing in"
              : "continuing"}
          , you agree to our{" "}
          <Link className="text-primary" href="/terms">
            Terms of Service
          </Link>{" "}
          and{" "}
          <Link className="text-primary" href="/privacy">
            Privacy Policy
          </Link>
        </Text>
      </Container>
    </Box>
  );
};
