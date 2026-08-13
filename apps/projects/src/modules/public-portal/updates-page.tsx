"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRightIcon, UpdatesIcon } from "icons";
import { Box, Flex, Text } from "ui";
import { markFeedbackUpdatesSeenAction } from "./actions";
import { PublicPortalShell } from "./portal-shell";
import { isGuestParticipant } from "./participant";
import { RequestStatusPill } from "./request-card";
import type {
  PublicPortal,
  PublicPortalParticipant,
  PublicPortalUpdate,
} from "./types";
import { getRequestPathBySlug, getUpdatePathBySlug } from "./utils";

export const PublicPortalUpdatesPage = ({
  participant,
  portal,
  updates,
}: {
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  updates: PublicPortalUpdate[];
}) => {
  const [updatesSeen, setUpdatesSeen] = useState(false);
  const shellParticipant =
    updatesSeen && isGuestParticipant(participant)
      ? { ...participant, unreadUpdateCount: 0 }
      : participant;

  useEffect(() => {
    if (
      !isGuestParticipant(participant) ||
      participant.unreadUpdateCount === 0
    ) {
      return;
    }

    void markFeedbackUpdatesSeenAction(portal.slug).then((response) => {
      if (!response.error) setUpdatesSeen(true);
    });
  }, [participant, portal.slug]);

  return (
    <PublicPortalShell
      activeTab="updates"
      participant={shellParticipant}
      portal={portal}
    >
      <Box className="mx-auto w-full max-w-4xl px-4 py-8 md:px-6 md:py-12">
        <Flex align="center" className="mb-3" gap={2}>
          <UpdatesIcon className="h-5" />
          <Text
            as="h1"
            className="text-2xl tracking-tight"
            fontWeight="semibold"
          >
            Updates
          </Text>
        </Flex>
        <Text className="mb-10 max-w-2xl leading-6" color="muted">
          Product news, shipped work, and progress connected directly to public
          feedback.
        </Text>

        {updates.length > 0 ? (
          <Box className="space-y-5">
            {updates.map((update) => (
              <article
                className="border-border bg-surface group overflow-hidden rounded-2xl border-[0.5px]"
                key={update.id}
              >
                {update.coverImageUrl ? (
                  <Box
                    aria-label=""
                    className="bg-surface-muted aspect-[2.4/1] w-full bg-cover bg-center"
                    role="img"
                    style={{
                      backgroundImage: `url(${JSON.stringify(update.coverImageUrl)})`,
                    }}
                  />
                ) : null}
                <Box className="p-6 md:p-7">
                  <Text className="text-sm" color="muted">
                    {update.publishedAtLabel}
                  </Text>
                  <Link
                    className="focus-visible:ring-ring mt-2 block rounded-lg focus-visible:ring-2 focus-visible:outline-none"
                    href={getUpdatePathBySlug(portal.slug, update.slug)}
                  >
                    <Flex align="center" gap={2} justify="between">
                      <Text
                        as="h2"
                        className="text-xl tracking-tight"
                        fontWeight="semibold"
                      >
                        {update.title}
                      </Text>
                      <ArrowRightIcon className="text-text-muted h-4 shrink-0 transition-transform group-hover:translate-x-1" />
                    </Flex>
                  </Link>
                  <Text className="mt-3 max-w-2xl leading-6" color="muted">
                    {update.summary ?? update.body}
                  </Text>
                  {update.linkedItems.length > 0 ? (
                    <Flex className="mt-5 flex-wrap gap-2">
                      {update.linkedItems.slice(0, 3).map((item) => (
                        <Link
                          className="border-border bg-surface-muted/50 hover:bg-state-hover inline-flex items-center gap-2 rounded-lg border px-2.5 py-1.5 text-sm transition"
                          href={getRequestPathBySlug(portal, item.slug)}
                          key={item.id}
                        >
                          <span className="max-w-64 truncate">
                            {item.title}
                          </span>
                          <RequestStatusPill status={item.status} />
                        </Link>
                      ))}
                    </Flex>
                  ) : null}
                </Box>
              </article>
            ))}
          </Box>
        ) : (
          <Flex
            align="center"
            className="border-border bg-surface-muted/30 min-h-72 rounded-2xl border border-dashed text-center"
            direction="column"
            justify="center"
          >
            <UpdatesIcon className="text-text-muted h-6" />
            <Text className="mt-3" fontWeight="semibold">
              No published updates yet
            </Text>
            <Text className="mt-1 max-w-sm" color="muted">
              Published progress and shipped work will appear here.
            </Text>
          </Flex>
        )}
      </Box>
    </PublicPortalShell>
  );
};
