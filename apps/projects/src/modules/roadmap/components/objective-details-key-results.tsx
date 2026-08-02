"use client";

import { format } from "date-fns";
import { OKRIcon as KeyResultIcon } from "icons";
import { Badge, Box, Flex, ProgressBar, Text, Wrapper } from "ui";
import { KeyResultContextMenu } from "@/modules/key-results/components/key-result-context-menu";
import type { KeyResult } from "@/modules/objectives/types";

const getKeyResultProgress = (keyResult: KeyResult) => {
  if (keyResult.measurementType === "boolean") {
    return keyResult.currentValue === 1 ? 100 : 0;
  }

  if (keyResult.measurementType === "percentage") {
    return Math.min(100, Math.max(0, keyResult.currentValue));
  }

  const totalChange = keyResult.targetValue - keyResult.startValue;
  if (totalChange === 0) return 0;

  return Math.min(
    100,
    Math.max(
      0,
      Math.round(
        ((keyResult.currentValue - keyResult.startValue) / totalChange) * 100,
      ),
    ),
  );
};

const formatDate = (date: string | null | undefined) => {
  if (!date) return "No target date";
  return format(new Date(date), "MMM d, yyyy");
};

export const ObjectiveDetailsKeyResults = ({
  keyResults,
  onSelect,
}: {
  keyResults: KeyResult[];
  onSelect?: (keyResult: KeyResult) => void;
}) => {
  return (
    <>
      <Flex align="center" className="mb-3" justify="between">
        <Text>Key results</Text>
        <Text className="text-[0.95rem]" color="muted">
          {keyResults.length}
        </Text>
      </Flex>
      {keyResults.length > 0 ? (
        <Flex direction="column" gap={2}>
          {keyResults.map((keyResult) => {
            const progress = getKeyResultProgress(keyResult);

            return (
              <KeyResultContextMenu
                key={keyResult.id}
                keyResult={keyResult}
                onOpenDetails={() => {
                  onSelect?.(keyResult);
                }}
              >
                <Wrapper className="flex items-center gap-3 rounded-xl px-3.5 py-3 shadow-none">
                  <button
                    className="focus-visible:ring-primary flex min-w-0 flex-1 items-center gap-3 rounded-lg text-left outline-none focus-visible:ring-1"
                    onClick={() => {
                      onSelect?.(keyResult);
                    }}
                    type="button"
                  >
                    <Badge
                      className="aspect-square h-9 shrink-0"
                      color="tertiary"
                    >
                      <KeyResultIcon strokeWidth={2.8} />
                    </Badge>
                    <Box className="min-w-0 flex-1">
                      <Flex align="center" gap={3} justify="between">
                        <Text
                          className="line-clamp-1 min-w-0"
                          title={keyResult.name}
                        >
                          {keyResult.name}
                        </Text>
                        <Text className="shrink-0 text-[0.95rem]" color="muted">
                          {progress}%
                        </Text>
                      </Flex>
                      <Flex align="center" className="mt-1.5" gap={3}>
                        <Text
                          className="line-clamp-1 text-[0.95rem]"
                          color="muted"
                        >
                          {formatDate(keyResult.endDate)}
                        </Text>
                        <ProgressBar
                          className="ml-auto w-20 shrink-0"
                          progress={progress}
                        />
                      </Flex>
                    </Box>
                  </button>
                </Wrapper>
              </KeyResultContextMenu>
            );
          })}
        </Flex>
      ) : (
        <Text color="muted">No key results yet.</Text>
      )}
    </>
  );
};
