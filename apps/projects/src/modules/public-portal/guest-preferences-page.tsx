"use client";

import { useState } from "react";
import Link from "next/link";
import { Box, Button, Flex, Switch, Text } from "ui";
import { toast } from "sonner";
import type { FeedbackPreferences } from "./actions";
import { updateFeedbackPreferencesAction } from "./actions";
import { PublicPortalShell } from "./portal-shell";
import type { PublicPortal, PublicPortalParticipant } from "./types";
import { getRequestPathBySlug } from "./utils";

export const PublicPortalGuestPreferencesPage = ({
  initialPreferences,
  participant,
  portal,
}: {
  initialPreferences: FeedbackPreferences | null;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
}) => {
  const [preferences, setPreferences] = useState(initialPreferences);
  const [pendingKey, setPendingKey] = useState<string | null>(null);

  const updatePreferences = async ({
    key,
    portalEmailsEnabled,
    items,
  }: {
    key: string;
    portalEmailsEnabled?: boolean;
    items?: { itemId: string; following: boolean }[];
  }) => {
    setPendingKey(key);
    const response = await updateFeedbackPreferencesAction({
      portalSlug: portal.slug,
      portalEmailsEnabled,
      items,
    });
    setPendingKey(null);
    if (!response.data) {
      toast.error("Unable to update email preferences", {
        description: response.error?.message ?? "Please try again.",
      });
      return;
    }
    setPreferences(response.data);
    toast.success("Email preferences saved");
  };

  return (
    <PublicPortalShell participant={participant} portal={portal}>
      <Box className="mx-auto w-full max-w-2xl px-4 py-10 md:px-6 md:py-14">
        <Text as="h1" className="text-2xl" fontWeight="semibold">
          Feedback email preferences
        </Text>
        <Text className="mt-2 max-w-xl leading-6" color="muted">
          Choose which public feedback updates this portal may email you. These
          settings apply only to {portal.workspace.name}.
        </Text>

        {preferences ? (
          <Box className="mt-8 space-y-5">
            <Flex
              align="center"
              className="border-border bg-surface rounded-xl border-[0.5px] p-5"
              gap={4}
              justify="between"
            >
              <Box>
                <Text fontWeight="semibold">Portal emails</Text>
                <Text className="mt-1 text-sm leading-5" color="muted">
                  Allow meaningful status, reply, and published-update emails
                  from this portal.
                </Text>
              </Box>
              <Switch
                aria-label="Portal emails"
                checked={preferences.portalEmailsEnabled}
                disabled={pendingKey !== null}
                onCheckedChange={(checked) => {
                  void updatePreferences({
                    key: "portal",
                    portalEmailsEnabled: checked,
                  });
                }}
              />
            </Flex>

            {preferences.items.length > 0 ? (
              <Box className="border-border bg-surface overflow-hidden rounded-xl border-[0.5px]">
                <Box className="border-border/70 border-b px-5 py-4">
                  <Text fontWeight="semibold">Followed feedback</Text>
                  <Text className="mt-1 text-sm" color="muted">
                    Control emails for individual feedback items.
                  </Text>
                </Box>
                {preferences.items.map((item) => (
                  <Flex
                    align="center"
                    className="border-border/50 border-b px-5 py-4 last:border-b-0"
                    gap={4}
                    justify="between"
                    key={item.itemId}
                  >
                    <Link
                      className="min-w-0 flex-1 hover:underline"
                      href={getRequestPathBySlug(portal, item.itemSlug)}
                    >
                      <Text className="truncate" fontWeight="medium">
                        {item.title}
                      </Text>
                    </Link>
                    <Switch
                      aria-label={`Follow ${item.title}`}
                      checked={item.following}
                      disabled={pendingKey !== null}
                      onCheckedChange={(checked) => {
                        void updatePreferences({
                          key: item.itemId,
                          items: [{ itemId: item.itemId, following: checked }],
                        });
                      }}
                    />
                  </Flex>
                ))}
              </Box>
            ) : null}

            {preferences.portalEmailsEnabled ? (
              <Button
                color="danger"
                disabled={pendingKey !== null}
                onClick={() => {
                  void updatePreferences({
                    key: "unsubscribe-all",
                    portalEmailsEnabled: false,
                  });
                }}
                variant="naked"
              >
                Unsubscribe from all portal emails
              </Button>
            ) : (
              <Box
                className="border-border bg-surface-muted/40 rounded-xl border p-4"
                role="status"
              >
                <Text fontWeight="medium">All portal emails are paused</Text>
                <Text className="mt-1 text-sm" color="muted">
                  You can turn them back on at any time while this preference
                  session remains active.
                </Text>
              </Box>
            )}
          </Box>
        ) : (
          <Box className="border-border bg-surface mt-8 rounded-xl border p-6">
            <Text fontWeight="semibold">This preference link has expired</Text>
            <Text className="mt-2 leading-6" color="muted">
              Open a newer feedback email and use its manage-preferences link,
              or verify your email on this portal again.
            </Text>
          </Box>
        )}
      </Box>
    </PublicPortalShell>
  );
};
