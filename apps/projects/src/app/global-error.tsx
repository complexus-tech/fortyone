"use client";
import * as Sentry from "@sentry/nextjs";
import type Error from "next/error";
import { ArrowLeft2Icon, ReloadIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { useEffect } from "react";
import { Instrument_Sans as InstrumentSans } from "next/font/google";
import { FortyOneLogo } from "@/components/ui";
import "../styles/global.css";

const font = InstrumentSans({
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    Sentry.captureException(error);
  }, [error]);
  return (
    <html className={font.className} lang="en">
      <body>
        <Box className="flex h-screen items-center justify-center">
          <Box className="flex flex-col items-center">
            <FortyOneLogo className="h-8 text-white" />
            <Text className="mb-6 mt-10" fontSize="3xl">
              There was an error
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              Oops! something went wrong. Please try again.
            </Text>
            <Flex gap={2} justify="center">
              <Button
                className="gap-1 pl-2"
                color="tertiary"
                leftIcon={<ReloadIcon className="h-[1.05rem] w-auto" />}
                onClick={() => {
                  reset();
                  window.location.reload();
                }}
              >
                Reload page
              </Button>
              <Button
                className="gap-1 pl-2"
                color="tertiary"
                href="https://www.fortyone.app/login"
                leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
              >
                Go to home
              </Button>
            </Flex>
          </Box>
        </Box>
      </body>
    </html>
  );
}
