"use client";
import { Box, Flex, Text } from "ui";
import { useEffect } from "react";
import { logOut } from "@/components/shared/sidebar/actions";
import { Logo } from "@/components/ui";
import { useAnalytics } from "@/hooks";

const clearClientStorage = () => {
  // Clear other client-side cookies
  document.cookie.split(";").forEach((c) => {
    document.cookie = c
      .replace(/^ +/, "")
      .replace(/=.*/, `=;expires=${new Date().toUTCString()};path=/`);
  });

  if ("caches" in window) {
    caches.keys().then((names) => {
      names.forEach((name) => {
        caches.delete(name);
      });
    });
  }
};

export default function Page() {
  const { analytics } = useAnalytics();

  useEffect(() => {
    const performLogout = async () => {
      try {
        await logOut();
        analytics.logout(true);
      } catch (error) {
        // console.error("Error during logout process:", error);
      } finally {
        clearClientStorage();
        // Redirect to main domain after logout to break out of subdomain context
        const mainDomain =
          process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app"
            ? "https://fortyone.app"
            : "/";
        window.location.href = `${mainDomain}?signedOut=true`;
      }
    };

    performLogout();
  }, [analytics]);
  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Box className="bg-primary aspect-square w-max animate-pulse rounded-full p-4">
          <Logo className="text-foreground h-8" asIcon />
        </Box>
        <Text className="mt-4" color="muted" fontWeight="medium">
          Logging out...
        </Text>
      </Flex>
    </Flex>
  );
}
