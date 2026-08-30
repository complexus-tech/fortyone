"use client";

import { StrategyIcon, UnlinkIcon } from "icons";
import { Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import { useAlignObjectiveMutation, useStrategyMap } from "./hooks";

export const ObjectivePillarProperty = ({
  buttonClassName,
  canUpdate,
  layout = "inline",
  objectiveId,
  onAlignmentSettled,
  variant = "naked",
}: {
  buttonClassName?: string;
  canUpdate: boolean;
  layout?: "detail" | "inline";
  objectiveId: string;
  onAlignmentSettled?: () => void;
  variant?: "naked" | "solid";
}) => {
  const { data: strategy } = useStrategyMap();
  const alignObjective = useAlignObjectiveMutation({ onAlignmentSettled });
  const pillars = strategy?.pillars ?? [];
  const currentPillar = pillars.find((pillar) =>
    pillar.objectiveIds.includes(objectiveId),
  );

  if (pillars.length === 0) return null;

  const editor = (
    <Menu>
      <Menu.Button>
        <Button
          align="left"
          className={cn(
            "max-w-full min-w-0 justify-start font-normal",
            buttonClassName,
          )}
          color="tertiary"
          disabled={!canUpdate || alignObjective.isPending}
          leftIcon={<StrategyIcon className="h-4 w-4 shrink-0" />}
          size="sm"
          type="button"
          variant={variant}
        >
          <span className="truncate">
            {currentPillar?.name ?? "Align to pillar"}
          </span>
        </Button>
      </Menu.Button>
      <Menu.Items align="start" className="w-64">
        <Menu.Group>
          {pillars.map((pillar) => (
            <Menu.Item
              active={pillar.id === currentPillar?.id}
              key={pillar.id}
              onSelect={() => {
                alignObjective.mutate({
                  objectiveId,
                  pillarId: pillar.id,
                });
              }}
            >
              <Text className="truncate">{pillar.name}</Text>
            </Menu.Item>
          ))}
        </Menu.Group>
        {currentPillar ? (
          <>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  alignObjective.mutate({ objectiveId, pillarId: null });
                }}
              >
                <UnlinkIcon className="h-4 w-4" />
                Remove pillar alignment
              </Menu.Item>
            </Menu.Group>
          </>
        ) : null}
      </Menu.Items>
    </Menu>
  );

  if (layout === "detail") {
    return (
      <Flex align="center" className="min-h-9" gap={4}>
        <Text className="w-28 shrink-0" color="muted">
          Pillar
        </Text>
        <Box className="min-w-0 flex-1">{editor}</Box>
      </Flex>
    );
  }

  return editor;
};
