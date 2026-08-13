import Link from "next/link";
import { ArrowLeftIcon, UpdatesIcon } from "icons";
import { Box, Flex, Text } from "ui";
import { PublicPortalShell } from "./portal-shell";
import { RequestStatusPill } from "./request-card";
import type {
  PublicPortal,
  PublicPortalParticipant,
  PublicPortalUpdate,
} from "./types";
import { getPortalPath, getRequestPathBySlug } from "./utils";

export const PublicPortalUpdateDetailPage = ({
  participant,
  portal,
  update,
}: {
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  update: PublicPortalUpdate;
}) => (
  <PublicPortalShell
    activeTab="updates"
    participant={participant}
    portal={portal}
  >
    <Box className="mx-auto w-full max-w-4xl px-4 py-8 md:px-6 md:py-12">
      <Link
        className="text-text-muted hover:text-foreground inline-flex items-center gap-2 text-sm transition"
        href={getPortalPath(portal, "updates")}
      >
        <ArrowLeftIcon className="h-4" />
        All updates
      </Link>
      <article className="mt-8">
        <Flex align="center" className="text-text-muted" gap={2}>
          <UpdatesIcon className="h-4" />
          <Text className="text-sm">{update.publishedAtLabel}</Text>
        </Flex>
        <Text
          as="h1"
          className="mt-4 max-w-3xl text-3xl leading-tight tracking-tight md:text-4xl"
          fontWeight="semibold"
        >
          {update.title}
        </Text>
        {update.summary ? (
          <Text className="mt-4 max-w-2xl text-lg leading-7" color="muted">
            {update.summary}
          </Text>
        ) : null}
        {update.coverImageUrl ? (
          <Box
            className="bg-surface-muted mt-8 aspect-[2.2/1] w-full rounded-2xl bg-cover bg-center"
            role="presentation"
            style={{
              backgroundImage: `url(${JSON.stringify(update.coverImageUrl)})`,
            }}
          />
        ) : null}
        <Text className="mt-9 max-w-3xl text-[1.05rem] leading-8 whitespace-pre-wrap">
          {update.body}
        </Text>
      </article>

      {update.linkedItems.length > 0 ? (
        <Box className="border-border mt-12 border-t pt-8">
          <Text as="h2" className="text-lg" fontWeight="semibold">
            Feedback included in this update
          </Text>
          <Box className="mt-4 grid gap-3">
            {update.linkedItems.map((item) => (
              <Link
                className="border-border bg-surface hover:bg-state-hover flex items-center justify-between gap-4 rounded-xl border-[0.5px] px-4 py-3 transition"
                href={getRequestPathBySlug(portal, item.slug)}
                key={item.id}
              >
                <Text className="min-w-0 truncate" fontWeight="medium">
                  {item.title}
                </Text>
                <RequestStatusPill status={item.status} />
              </Link>
            ))}
          </Box>
        </Box>
      ) : null}
    </Box>
  </PublicPortalShell>
);
