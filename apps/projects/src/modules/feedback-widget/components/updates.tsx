import { ArrowLeft2Icon, UpdatesIcon } from "icons";
import { Box, Flex, Text } from "ui";
import { cn } from "lib";
import type { PublicPortalUpdate } from "@/shared/feedback-widget/types";
import { statusAccent } from "./utils";
import { EmptyState, WidgetBackButton } from "./widget-ui";

export const UpdateDetail = ({
  onBack,
  update,
}: {
  onBack: () => void;
  update: PublicPortalUpdate;
}) => (
  <Box className="bg-background absolute inset-0 z-20 flex min-h-0 flex-col">
    <Flex align="center" className="h-16 shrink-0 px-4">
      <WidgetBackButton aria-label="Back to updates" onClick={onBack}>
        <ArrowLeft2Icon className="h-5" />
      </WidgetBackButton>
    </Flex>
    <Box className="min-h-0 flex-1 overflow-y-auto px-6 pb-10">
      <Text
        className="text-[11px] tracking-[0.09em] uppercase"
        color="muted"
        fontWeight="semibold"
      >
        {update.publishedAtLabel} · Update
      </Text>
      <Text
        as="h1"
        className="mt-3 text-[23px] leading-8"
        fontWeight="semibold"
      >
        {update.title}
      </Text>
      {update.summary ? (
        <Text className="mt-4 text-[14px] leading-6" color="muted">
          {update.summary}
        </Text>
      ) : null}
      <Box className="border-border/70 mt-7 border-t pt-7">
        <Text
          className="text-[13px] leading-6 whitespace-pre-wrap"
          color="muted"
        >
          {update.body}
        </Text>
      </Box>
      {update.linkedItems.length > 0 ? (
        <Box className="border-border bg-surface mt-7 rounded-lg border p-4">
          <Text
            className="text-[11px] tracking-[0.08em] uppercase"
            color="muted"
            fontWeight="semibold"
          >
            Related feedback
          </Text>
          {update.linkedItems.map((item) => (
            <Flex align="center" className="mt-3 gap-3" key={item.id}>
              <span
                className={cn("size-2 rounded-sm", statusAccent(item.status))}
              />
              <Text className="min-w-0 flex-1 text-[12px]" fontWeight="medium">
                {item.title}
              </Text>
            </Flex>
          ))}
        </Box>
      ) : null}
    </Box>
  </Box>
);

export const UpdatesList = ({
  onOpen,
  updates,
}: {
  onOpen: (update: PublicPortalUpdate) => void;
  updates: PublicPortalUpdate[];
}) => {
  if (updates.length === 0) {
    return (
      <EmptyState
        body="Product news and shipped improvements will appear here."
        icon={UpdatesIcon}
        title="No updates yet"
      />
    );
  }

  return (
    <Box className="px-5 py-3">
      {updates.map((update) => (
        <button
          className="border-border/70 hover:bg-state-hover/35 focus-visible:ring-ring w-full border-b py-5 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
          key={update.id}
          onClick={() => {
            onOpen(update);
          }}
          type="button"
        >
          <Text
            className="text-[10px] tracking-[0.09em] uppercase"
            color="muted"
            fontWeight="semibold"
          >
            {update.publishedAtLabel} · Update
          </Text>
          <Text className="mt-2 text-[16px] leading-6" fontWeight="semibold">
            {update.title}
          </Text>
          <Text
            className="mt-2 line-clamp-3 text-[12px] leading-5"
            color="muted"
          >
            {update.summary || update.body}
          </Text>
        </button>
      ))}
    </Box>
  );
};
