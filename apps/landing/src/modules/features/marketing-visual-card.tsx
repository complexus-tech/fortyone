import { Box, Text } from "ui";
import type { MarketingVisual } from "@/components/shared/marketing-detail-page";
import {
  FEATURE_STORY_META_TEXT_CLASS,
  FEATURE_STORY_SURFACE_CLASS,
  FEATURE_STORY_TEXT_CLASS,
} from "@/modules/home/feature-story-section";

export function MarketingVisualCard({ visual }: { visual: MarketingVisual }) {
  return (
    <Box aria-hidden="true" className="flex h-full flex-col gap-3">
      <Box className={`${FEATURE_STORY_SURFACE_CLASS} px-4 py-3`}>
        <Box className="flex items-center justify-between gap-3">
          <Box className="min-w-0">
            <Text
              className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}
            >
              {visual.subheading}
            </Text>
            <Text
              className={`${FEATURE_STORY_TEXT_CLASS} text-foreground truncate font-semibold`}
            >
              {visual.heading}
            </Text>
          </Box>
          {visual.badge ? (
            <Text
              className={`${FEATURE_STORY_META_TEXT_CLASS} bg-primary/10 text-primary shrink-0 rounded-lg px-2.5 py-1 font-semibold`}
            >
              {visual.badge}
            </Text>
          ) : null}
        </Box>
      </Box>

      <Box className={`${FEATURE_STORY_SURFACE_CLASS} grid flex-1 gap-2.5 p-4`}>
        {visual.rows.map((row) => (
          <Box
            className="bg-surface-muted rounded-lg px-3 py-2.5"
            key={row.label}
          >
            <Box className="flex items-center justify-between gap-3">
              <Text
                className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}
              >
                {row.label}
              </Text>
              <Text
                className={`${FEATURE_STORY_META_TEXT_CLASS} text-foreground text-right font-semibold`}
              >
                {row.value}
              </Text>
            </Box>
            {row.width ? (
              <Box className="bg-background mt-2 h-1.5 overflow-hidden rounded-full">
                <Box
                  className="bg-primary h-full rounded-full"
                  style={{ width: row.width }}
                />
              </Box>
            ) : null}
          </Box>
        ))}
      </Box>

      {visual.note ? (
        <Box className={`${FEATURE_STORY_SURFACE_CLASS} px-4 py-3`}>
          <Text className={`${FEATURE_STORY_META_TEXT_CLASS} text-text-muted`}>
            {visual.note}
          </Text>
        </Box>
      ) : null}
    </Box>
  );
}
