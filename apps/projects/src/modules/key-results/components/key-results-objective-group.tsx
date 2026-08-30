"use client";

import { ArrowDownIcon, ObjectiveIcon, PlusIcon } from "icons";
import { usePathname } from "next/navigation";
import { Box, Button, Checkbox, Container, Flex, ProgressBar, Text } from "ui";
import { cn } from "lib";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { TeamColor } from "@/components/ui/team-color";
import { useLocalStorage } from "@/hooks/local-storage";
import { useTerminology } from "@/hooks/use-terminology-display";
import { NewKeyResultButton } from "@/modules/objectives/public/key-results";
import { hexToRgba } from "@/utils/color";
import type { ObjectiveKeyResultGroup } from "../utils";
import { KeyResultRow } from "./key-result-row";
import type { KeyResultsMember } from "./key-results-member";
import {
  isKeyResultGroupSelected,
  setKeyResultGroupSelection,
  setKeyResultSelection,
} from "./key-results-selection";

export const KeyResultsObjectiveGroup = ({
  group,
  memberById,
  selectedKeyResultIds,
  setSelectedKeyResultIds,
  teamColorById,
}: {
  group: ObjectiveKeyResultGroup;
  memberById: ReadonlyMap<string, KeyResultsMember>;
  selectedKeyResultIds: ReadonlySet<string>;
  setSelectedKeyResultIds: (ids: Set<string>) => void;
  teamColorById: ReadonlyMap<string, string>;
}) => {
  const pathname = usePathname();
  const { getTermDisplay } = useTerminology();
  const keyResultLabel = getTermDisplay("keyResultTerm");
  const [isCollapsed, setIsCollapsed] = useLocalStorage(
    `${pathname}:key-results:${group.objectiveId}`,
    false,
  );
  const groupKeyResultIds = group.keyResults.map((keyResult) => keyResult.id);
  const isGroupSelected = isKeyResultGroupSelected(
    groupKeyResultIds,
    selectedKeyResultIds,
  );

  return (
    <Box>
      <Container className="border-border bg-surface-muted/85 sticky top-0 z-1 border-b-[0.5px] py-[0.4rem] backdrop-blur select-none">
        <Flex align="center" justify="between">
          <Flex align="center" className="relative min-w-0 gap-1.5">
            <Checkbox
              checked={isGroupSelected}
              className="absolute -left-[1.6rem] hidden rounded md:inline"
              onCheckedChange={(checked) => {
                setSelectedKeyResultIds(
                  setKeyResultGroupSelection(
                    selectedKeyResultIds,
                    groupKeyResultIds,
                    Boolean(checked),
                  ),
                );
              }}
            />
            <Button
              className="min-w-0"
              color="tertiary"
              leftIcon={<ObjectiveIcon className="h-[1.1rem] shrink-0" />}
              onClick={() => {
                setIsCollapsed(!isCollapsed);
              }}
              rightIcon={
                <ArrowDownIcon
                  className={cn(
                    "text-text-muted h-4 w-auto shrink-0 transition",
                    {
                      "-rotate-90": isCollapsed,
                    },
                  )}
                  strokeWidth={1}
                />
              }
              size="sm"
              variant="naked"
            >
              <Text className="truncate" fontWeight="medium">
                {group.objectiveName}
              </Text>
            </Button>
            <Button
              className="pointer-events-none gap-1 pr-2"
              color="tertiary"
              leftIcon={<TeamColor color={teamColorById.get(group.teamId)} />}
              rounded="md"
              size="xs"
              style={{
                backgroundColor: hexToRgba(
                  teamColorById.get(group.teamId),
                  0.1,
                ),
                borderColor: hexToRgba(teamColorById.get(group.teamId), 0.2),
              }}
              tabIndex={-1}
              variant="outline"
            >
              {group.teamName}
            </Button>
            <Text className="shrink-0 whitespace-nowrap" color="muted">
              {group.keyResults.length}{" "}
              {getTermDisplay("keyResultTerm", {
                variant: group.keyResults.length === 1 ? "singular" : "plural",
              })}
            </Text>
          </Flex>
          <Flex align="center" className="shrink-0" gap={2}>
            <Flex
              align="center"
              className="hidden w-24 shrink-0 md:flex"
              gap={2}
            >
              <ProgressBar className="w-12" progress={group.averageProgress} />
              <Text color="muted">{group.averageProgress}%</Text>
            </Flex>
            <NewKeyResultButton
              aria-label={`Add ${keyResultLabel} to ${group.objectiveName}`}
              asIcon
              iconOnly
              leftIcon={
                <PlusIcon className="text-foreground h-[1.1rem] w-auto" />
              }
              objectiveId={group.objectiveId}
              size="sm"
              variant="outline"
            />
          </Flex>
        </Flex>
      </Container>
      {!isCollapsed
        ? group.keyResults.map((keyResult) => (
            <KeyResultRow
              key={keyResult.id}
              keyResult={keyResult}
              memberById={memberById}
              selected={selectedKeyResultIds.has(keyResult.id)}
              setSelected={(selected) => {
                setSelectedKeyResultIds(
                  setKeyResultSelection(
                    selectedKeyResultIds,
                    keyResult.id,
                    selected,
                  ),
                );
              }}
            />
          ))
        : null}
      {!isCollapsed ? (
        <RowWrapper className="grid h-12 py-0 md:grid-cols-2">
          <Text className="min-w-0 truncate whitespace-nowrap" color="muted">
            Showing{" "}
            <span className="font-semibold">{group.keyResults.length}</span>{" "}
            {getTermDisplay("keyResultTerm", {
              variant: group.keyResults.length === 1 ? "singular" : "plural",
            })}{" "}
            for <span className="font-semibold">{group.objectiveName}</span>
          </Text>
        </RowWrapper>
      ) : null}
    </Box>
  );
};
