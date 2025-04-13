"use client";
import posthog from "posthog-js";
import { ArrowLeftIcon } from "icons";
import { Box, Button, Text } from "ui";
import { useEffect } from "react";
import { ComplexusLogo } from "@/components/ui";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    posthog.captureException(error);
  }, [error]);
  return (
    <html lang="en">
      <body>
        <Box className="flex h-screen items-center justify-center">
          <Box className="flex flex-col items-center">
            <ComplexusLogo className="h-8 text-white" />
            <Text className="mb-6 mt-10" fontSize="3xl">
              There was an error
            </Text>
            <Text className="mb-6 max-w-md text-center" color="muted">
              Oops! something went wrong. Please try again.
            </Text>
            <Button
              className="gap-1 pl-2"
              color="tertiary"
              leftIcon={<ArrowLeftIcon className="h-[1.05rem] w-auto" />}
              onClick={() => {
                reset();
                window.location.reload();
              }}
            >
              Reload page
            </Button>
          </Box>
        </Box>
      </body>
    </html>
  );
}
