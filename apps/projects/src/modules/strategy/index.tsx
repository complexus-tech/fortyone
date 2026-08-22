"use client";

import { useState } from "react";
import { Box, BreadCrumbs, Button, Dialog, Flex, Text } from "ui";
import { MinusIcon, PlusIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useUserRole } from "@/hooks";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { StrategyEditorDialog } from "./strategy-editor-dialog";
import { StrategyMapCanvas } from "./strategy-map-canvas";
import { StrategyMapSkeleton } from "./strategy-map-skeleton";
import {
  StrategySelectedDetails,
  type SelectedStrategyNode,
} from "./strategy-selected-details";
import { useStrategyMap, useStrategyMutations } from "./hooks";

export const WorkspaceStrategyMapPage = () => {
  const { userRole } = useUserRole();
  const { data: strategy, isPending: isStrategyPending } = useStrategyMap();
  const { data: objectives = [], isPending: areObjectivesPending } =
    useObjectives();
  const {
    updateStrategy,
    createPillar,
    updatePillar,
    alignObjective,
    deletePillar,
  } = useStrategyMutations();
  const [isPillarOpen, setIsPillarOpen] = useState(false);
  const [isGoalOpen, setIsGoalOpen] = useState(false);
  const [pillarToDelete, setPillarToDelete] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<SelectedStrategyNode | null>(
    null,
  );
  const [zoom, setZoom] = useState(1);
  const [resetSignal, setResetSignal] = useState(0);
  const showUnaligned = true;
  const isGuest = userRole === "guest";

  return (
    <>
      <HeaderContainer className="bg-surface-muted/20 dark:bg-surface-elevated/20 justify-between backdrop-blur-xl">
        <Flex gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Strategy Map",
              },
            ]}
          />
        </Flex>
        <Flex align="center" className="gap-2">
          <Flex
            align="center"
            className="border-border bg-surface/80 dark:bg-surface/70 overflow-hidden rounded-lg border backdrop-blur"
          >
            <button
              aria-label="Zoom out"
              className="hover:bg-state-hover grid h-8 w-9 place-items-center"
              onClick={() => {
                setZoom((value) =>
                  Math.max(0.5, Number((value - 0.1).toFixed(2))),
                );
              }}
              type="button"
            >
              <MinusIcon className="h-4 w-4" />
            </button>
            <Text className="border-border min-w-14 border-x text-center tabular-nums">
              {Math.round(zoom * 100)}%
            </Text>
            <button
              aria-label="Zoom in"
              className="hover:bg-state-hover grid h-8 w-9 place-items-center"
              onClick={() => {
                setZoom((value) =>
                  Math.min(1.6, Number((value + 0.1).toFixed(2))),
                );
              }}
              type="button"
            >
              <PlusIcon className="h-4 w-4" />
            </button>
            <button
              className="border-border hover:bg-state-hover h-8 border-l px-3 font-medium"
              onClick={() => {
                setZoom(1);
                setResetSignal((value) => value + 1);
              }}
              type="button"
            >
              Reset
            </button>
          </Flex>
          <Button
            color="primary"
            disabled={isGuest}
            leftIcon={<PlusIcon className="h-[1.1rem] text-current" />}
            onClick={() => {
              setIsPillarOpen(true);
            }}
            size="sm"
          >
            Add strategic pillar
          </Button>
        </Flex>
      </HeaderContainer>

      <Box className="relative h-[calc(100dvh-4rem)]">
        {isStrategyPending || areObjectivesPending || !strategy ? (
          <StrategyMapSkeleton />
        ) : (
          <StrategyMapCanvas
            canEdit={!isGuest}
            objectives={objectives}
            onAddPillar={() => {
              if (!isGuest) setIsPillarOpen(true);
            }}
            onAlign={(objectiveId, pillarId) => {
              if (!isGuest) alignObjective.mutate({ objectiveId, pillarId });
            }}
            onDeletePillar={(pillarId) => {
              if (!isGuest) setPillarToDelete(pillarId);
            }}
            onSelectGoal={() => {
              if (!strategy.ultimateGoal.trim() && !isGuest) {
                setIsGoalOpen(true);
                return;
              }
              setSelectedNode({ type: "goal" });
            }}
            onSelectKeyResult={(objective, keyResult) => {
              setSelectedNode({
                keyResultId: keyResult.id,
                objectiveId: objective.id,
                type: "key-result",
              });
            }}
            onSelectObjective={(objective) => {
              setSelectedNode({
                objectiveId: objective.id,
                type: "objective",
              });
            }}
            onSelectPillar={(pillar) => {
              setSelectedNode({ pillarId: pillar.id, type: "pillar" });
            }}
            onZoomChange={setZoom}
            resetSignal={resetSignal}
            showUnaligned={showUnaligned}
            strategy={strategy}
            zoom={zoom}
          />
        )}

        {strategy ? (
          <StrategySelectedDetails
            canEdit={!isGuest}
            objectives={objectives}
            onClose={() => {
              setSelectedNode(null);
            }}
            onSaveGoal={(ultimateGoal, description) => {
              updateStrategy.mutate({ ultimateGoal, description });
            }}
            onSavePillar={(pillarId, name, description) => {
              updatePillar.mutate({
                pillarId,
                data: { name, description },
              });
            }}
            onSelectKeyResult={(objective, keyResult) => {
              setSelectedNode({
                keyResultId: keyResult.id,
                objectiveId: objective.id,
                type: "key-result",
              });
            }}
            selectedNode={selectedNode}
            strategy={strategy}
          />
        ) : null}
      </Box>

      <StrategyEditorDialog
        initialDescription={strategy?.description ?? ""}
        initialName={strategy?.ultimateGoal ?? ""}
        isOpen={isGoalOpen}
        isPending={updateStrategy.isPending}
        nameLabel="Ultimate goal"
        onOpenChange={setIsGoalOpen}
        onSave={(ultimateGoal, description) => {
          updateStrategy.mutate(
            { ultimateGoal, description },
            {
              onSuccess: () => {
                setIsGoalOpen(false);
              },
            },
          );
        }}
        title="Add ultimate goal"
      />
      <StrategyEditorDialog
        isOpen={isPillarOpen}
        isPending={createPillar.isPending}
        nameLabel="Pillar name"
        onOpenChange={setIsPillarOpen}
        onSave={(name, description) => {
          createPillar.mutate(
            {
              name,
              description,
              orderIndex: strategy?.pillars.length ?? 0,
            },
            {
              onSuccess: () => {
                setIsPillarOpen(false);
              },
            },
          );
        }}
        title="Add strategic pillar"
      />
      <Dialog
        onOpenChange={(isOpen) => {
          if (!isOpen) setPillarToDelete(null);
        }}
        open={Boolean(pillarToDelete)}
      >
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-xl">
              Delete strategic pillar?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text className="text-[1rem] leading-6" color="muted">
              Objectives connected to this pillar will move to the unaligned
              section. The objectives and their key results will not be deleted.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              onClick={() => {
                setPillarToDelete(null);
              }}
              type="button"
            >
              Cancel
            </Button>
            <Button
              color="danger"
              disabled={deletePillar.isPending}
              onClick={() => {
                if (!pillarToDelete) return;
                deletePillar.mutate(pillarToDelete, {
                  onSuccess: () => {
                    setPillarToDelete(null);
                  },
                });
              }}
              type="button"
            >
              Delete pillar
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
