"use client";

import Link from "next/link";
import {
  ChevronRightIcon,
  ExternalLinkIcon,
  ObjectiveIcon,
  UnlinkIcon,
} from "icons";
import { ContextMenu, Flex, Text } from "ui";
import { ContextMenuLabel } from "./strategy-map-card-primitives";

type StrategyMapObjectiveContextMenuProps = {
  canEdit: boolean;
  currentPillarId: string | null;
  objectiveId: string;
  objectivePath: string;
  objectiveStatusId: string;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onSetStatus: (statusId: string) => void;
  pillars: readonly { id: string; name: string }[];
  status?: { color: string };
  statuses: readonly { color: string; id: string; name: string }[];
};

export const StrategyMapObjectiveContextMenu = ({
  canEdit,
  currentPillarId,
  objectiveId,
  objectivePath,
  objectiveStatusId,
  onAlign,
  onSetStatus,
  pillars,
  status,
  statuses,
}: StrategyMapObjectiveContextMenuProps) => (
  <ContextMenu.Items className="w-64">
    <ContextMenu.Group>
      <ContextMenu.Item asChild>
        <Link href={objectivePath}>
          <ContextMenuLabel icon={<ExternalLinkIcon className="h-4 w-4" />}>
            Open objective
          </ContextMenuLabel>
        </Link>
      </ContextMenu.Item>
    </ContextMenu.Group>
    {canEdit ? (
      <>
        <ContextMenu.Separator />
        {statuses.length > 0 ? (
          <ContextMenu.Group>
            <ContextMenu.SubMenu>
              <ContextMenu.SubTrigger className="justify-between">
                <Flex align="center" className="gap-2">
                  <span
                    aria-hidden
                    className="h-2 w-2 rounded-full"
                    style={{
                      backgroundColor:
                        status?.color ?? "var(--color-text-muted)",
                    }}
                  />
                  <Text>Set status</Text>
                </Flex>
                <ChevronRightIcon className="text-text-muted h-4 w-4" />
              </ContextMenu.SubTrigger>
              <ContextMenu.SubItems className="min-w-44">
                <ContextMenu.Group>
                  {statuses.map((statusOption) => (
                    <ContextMenu.Item
                      active={statusOption.id === objectiveStatusId}
                      key={statusOption.id}
                      onSelect={() => {
                        onSetStatus(statusOption.id);
                      }}
                    >
                      <span
                        aria-hidden
                        className="h-2 w-2 rounded-full"
                        style={{ backgroundColor: statusOption.color }}
                      />
                      <Text>{statusOption.name}</Text>
                    </ContextMenu.Item>
                  ))}
                </ContextMenu.Group>
              </ContextMenu.SubItems>
            </ContextMenu.SubMenu>
          </ContextMenu.Group>
        ) : null}
        {pillars.length > 0 ? (
          <ContextMenu.Group>
            <ContextMenu.SubMenu>
              <ContextMenu.SubTrigger className="justify-between">
                <Flex align="center" className="gap-2">
                  <ObjectiveIcon className="text-text-secondary h-4 w-4" />
                  <Text>Align to pillar</Text>
                </Flex>
                <ChevronRightIcon className="text-text-muted h-4 w-4" />
              </ContextMenu.SubTrigger>
              <ContextMenu.SubItems className="max-w-64 min-w-48">
                <ContextMenu.Group>
                  {pillars.map((pillar) => (
                    <ContextMenu.Item
                      active={pillar.id === currentPillarId}
                      key={pillar.id}
                      onSelect={() => {
                        onAlign(objectiveId, pillar.id);
                      }}
                    >
                      <Text className="max-w-52 truncate">{pillar.name}</Text>
                    </ContextMenu.Item>
                  ))}
                </ContextMenu.Group>
              </ContextMenu.SubItems>
            </ContextMenu.SubMenu>
          </ContextMenu.Group>
        ) : null}
        {currentPillarId ? (
          <>
            <ContextMenu.Separator />
            <ContextMenu.Group>
              <ContextMenu.Item
                className="whitespace-nowrap"
                onSelect={() => {
                  onAlign(objectiveId, null);
                }}
              >
                <ContextMenuLabel icon={<UnlinkIcon className="h-4 w-4" />}>
                  Remove pillar alignment
                </ContextMenuLabel>
              </ContextMenu.Item>
            </ContextMenu.Group>
          </>
        ) : null}
      </>
    ) : null}
  </ContextMenu.Items>
);
