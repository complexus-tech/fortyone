"use client";
import type { Metadata } from "next";
import { Box, Flex, Text } from "ui";
import { useEffect } from "react";
import { logOut } from "@/components/shared/sidebar/actions";
import { ComplexusLogo } from "@/components/ui";
import { useAnalytics } from "@/hooks";

const clearClientStorage = () => {
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

export const metadata: Metadata = {
  title: "Logout",
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
        window.location.href = "https://www.complexus.app?signedOut=true";
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
        <Box className="aspect-square w-max animate-pulse rounded-full bg-primary p-4">
          <ComplexusLogo className="h-8 text-white" />
        </Box>
        <Text className="mt-4" color="muted" fontWeight="medium">
          Logging out...
        </Text>
      </Flex>
    </Flex>
  );
}
