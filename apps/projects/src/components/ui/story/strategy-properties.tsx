"use client";

import { cn } from "lib";
import { ObjectiveIcon, OKRIcon } from "icons";
import { Box, Button, Flex, Text, Tooltip } from "ui";
import { useTerminology } from "@/hooks";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { getKeyResultProgress } from "@/modules/key-results/utils";
import type { Story } from "@/modules/stories/types";
import {
  KeyResultMenu,
  ObjectiveKeyResultMenu,
} from "./objective-key-result-menu";

type StoryStrategyPropertiesProps = Pick<
  Story,
  "keyResultId" | "objective" | "objectiveId" | "teamId"
> & {
  asKanban?: boolean;
  disabled: boolean;
  handleUpdate: (data: Partial<Story>) => void;
  showKeyResult: boolean;
  showObjective: boolean;
};

const propertyLabelClassName = (
  asKanban: boolean | undefined,
  maxWidth: string,
) =>
  cn(maxWidth, "truncate", {
    "inline-block": asKanban,
    "hidden @7xl:inline-block": !asKanban,
  });

export const StoryStrategyProperties = ({
  asKanban,
  disabled,
  handleUpdate,
  keyResultId,
  objective,
  objectiveId,
  showKeyResult,
  showObjective,
  teamId,
}: StoryStrategyPropertiesProps) => {
  const { getTermDisplay } = useTerminology();
  const selectedObjective =
    objectiveId && objective?.id === objectiveId ? objective : null;
  const { data: objectiveDetails } = useObjective(objectiveId, teamId);
  const resolvedObjective = objectiveDetails ?? selectedObjective;
  const { data: keyResults = [] } = useKeyResults(
    objectiveId ?? "",
    Boolean(objectiveId && keyResultId),
  );
  const selectedKeyResult = keyResults.find(({ id }) => id === keyResultId);
  const objectiveName =
    typeof resolvedObjective?.name === "string"
      ? resolvedObjective.name.trim() || null
      : null;
  const keyResultName =
    typeof selectedKeyResult?.name === "string"
      ? selectedKeyResult.name.trim() || null
      : null;
  const objectiveLabel =
    objectiveName ?? getTermDisplay("objectiveTerm", { capitalize: true });
  const keyResultLabel =
    keyResultName ?? getTermDisplay("keyResultTerm", { capitalize: true });
  const objectiveSummary =
    typeof objectiveDetails?.shortSummary === "string"
      ? objectiveDetails.shortSummary.trim() || null
      : null;
  const keyResultProgress = selectedKeyResult
    ? getKeyResultProgress(selectedKeyResult)
    : null;

  return (
    <>
      {showObjective && resolvedObjective && objectiveId ? (
        <Tooltip
          className="w-80 max-w-[calc(100vw-2rem)] py-3"
          collisionPadding={16}
          title={
            <Flex align="start" gap={2}>
              <ObjectiveIcon className="relative top-[3px] h-4 shrink-0" />
              <Box className="min-w-0">
                <Text
                  className={cn({ "mb-1.5": objectiveSummary })}
                  fontSize="md"
                >
                  {objectiveLabel}
                </Text>
                {objectiveSummary ? (
                  <Text
                    className="line-clamp-3 min-w-0 break-words"
                    color="muted"
                  >
                    {objectiveSummary}
                  </Text>
                ) : null}
              </Box>
            </Flex>
          }
        >
          <div className="inline-flex">
            <ObjectiveKeyResultMenu
              keyResultId={keyResultId}
              objectiveId={objectiveId}
              onChange={handleUpdate}
              teamId={teamId}
            >
              <Button
                aria-label={objectiveLabel}
                className="gap-1 px-2"
                color="tertiary"
                disabled={disabled}
                rounded="md"
                size="xs"
                type="button"
                variant="outline"
              >
                <ObjectiveIcon className="h-4" />
                <span className={propertyLabelClassName(asKanban, "max-w-32")}>
                  {objectiveLabel}
                </span>
              </Button>
            </ObjectiveKeyResultMenu>
          </div>
        </Tooltip>
      ) : null}
      {showKeyResult && objectiveId && keyResultId ? (
        <Tooltip
          className="w-72 max-w-[calc(100vw-2rem)] py-3"
          collisionPadding={16}
          title={
            <Box className="min-w-0">
              <Flex align="start" gap={2}>
                <OKRIcon
                  className="relative top-[3px] h-4 shrink-0"
                  strokeWidth={2.4}
                />
                <Text className="line-clamp-2 min-w-0" fontSize="md">
                  {keyResultLabel}
                </Text>
              </Flex>
              {keyResultProgress !== null ? (
                <Flex className="mt-2.5" justify="between">
                  <Text color="muted">Progress</Text>
                  <Text>{keyResultProgress}%</Text>
                </Flex>
              ) : null}
            </Box>
          }
        >
          <div className="inline-flex">
            <KeyResultMenu
              keyResultId={keyResultId}
              objectiveId={objectiveId}
              onChange={(nextKeyResultId) => {
                handleUpdate({ keyResultId: nextKeyResultId });
              }}
            >
              <Button
                aria-label={keyResultLabel}
                className="gap-1 px-2"
                color="tertiary"
                disabled={disabled}
                rounded="md"
                size="xs"
                type="button"
                variant="outline"
              >
                <OKRIcon className="h-4" strokeWidth={2.4} />
                <span className={propertyLabelClassName(asKanban, "max-w-32")}>
                  {keyResultLabel}
                </span>
              </Button>
            </KeyResultMenu>
          </div>
        </Tooltip>
      ) : null}
    </>
  );
};
