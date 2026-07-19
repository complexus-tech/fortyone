import { UpdatesIcon } from "icons";
import { Box, Flex, Text } from "ui";
import { PublicPortalShell } from "./portal-shell";
import type { PublicPortal, PublicPortalViewer } from "./types";

export const PublicPortalUpdatesPage = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => (
  <PublicPortalShell activeTab="updates" portal={portal} viewer={viewer}>
    <Box className="mx-auto w-full max-w-4xl px-4 py-8 md:px-6">
      <Flex align="center" className="mb-8" gap={2}>
        <UpdatesIcon className="h-5" />
        <Text className="text-xl" fontWeight="semibold">
          Updates
        </Text>
      </Flex>
      <Box className="space-y-4">
        {portal.updates.map((update) => (
          <Box
            className="border-border bg-surface rounded-xl border-[0.5px] p-5"
            key={update.id}
          >
            <Flex align="center" justify="between">
              <Text className="text-[1.1rem]" fontWeight="semibold">
                {update.title}
              </Text>
              <Text color="muted">{update.publishedAtLabel}</Text>
            </Flex>
            <Text className="mt-3 max-w-2xl leading-6" color="muted">
              {update.body}
            </Text>
          </Box>
        ))}
      </Box>
    </Box>
  </PublicPortalShell>
);
