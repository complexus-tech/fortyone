"use client";

import { ReloadIcon, WarningIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";

type AdminErrorStateProps = {
  digest?: string;
  onRetry?: () => void;
};

export const AdminErrorState = ({ digest, onRetry }: AdminErrorStateProps) => {
  return (
    <Box className="flex min-h-full items-center justify-center p-6">
      <Box className="border-border bg-surface shadow-shadow w-full max-w-xl rounded-lg border-[0.5px] p-6 shadow-lg">
        <Flex align="start" className="gap-4">
          <Box className="bg-warning/15 text-warning flex size-11 shrink-0 items-center justify-center rounded-lg">
            <WarningIcon className="h-5 w-auto" />
          </Box>
          <Box className="min-w-0 flex-1">
            <Text as="h1" className="text-xl" fontWeight="semibold">
              Admin data could not load
            </Text>
            <Text className="mt-2 text-[0.98rem]" color="muted">
              One of the protected Admin API requests failed while rendering
              this page. The endpoint and status are now logged in Vercel.
            </Text>
            {digest ? (
              <Text className="mt-3 text-[0.82rem]" color="muted">
                Error digest: {digest}
              </Text>
            ) : null}
            <Flex align="center" className="mt-5 gap-2">
              {onRetry ? (
                <Button
                  color="tertiary"
                  leftIcon={<ReloadIcon className="h-4 w-auto" />}
                  onClick={onRetry}
                  type="button"
                >
                  Reload
                </Button>
              ) : null}
              <Button color="tertiary" href="/overview">
                Overview
              </Button>
            </Flex>
          </Box>
        </Flex>
      </Box>
    </Box>
  );
};
