"use client";

import { useState } from "react";
import { Box, BreadCrumbs, Button, Dialog, Flex, Text } from "ui";
import { MinusIcon, PlusIcon, StrategyIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { useUserRole } from "@/hooks";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { StrategyEditorDialog } from "./strategy-editor-dialog";
import { StrategyMapCanvas } from "./strategy-map-canvas";
import { useStrategyMap, useStrategyMutations } from "./hooks";
import type { StrategicPillar } from "./types";

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
  const [isGoalOpen, setIsGoalOpen] = useState(false);
  const [isPillarOpen, setIsPillarOpen] = useState(false);
  const [pillarToDelete, setPillarToDelete] = useState<string | null>(null);
  const [pillarToEdit, setPillarToEdit] = useState<StrategicPillar | null>(
    null,
  );
  const [zoom, setZoom] = useState(1);
  const [resetSignal, setResetSignal] = useState(0);
  const showUnaligned = true;
  const isGuest = userRole === "guest";

  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Strategy Map",
                icon: (
                  <StrategyIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
                ),
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
            color="invert"
            disabled={isGuest}
            leftIcon={<PlusIcon className="h-[1.1rem] text-current" />}
            onClick={() => {
              setIsPillarOpen(true);
            }}
            size="sm"
          >
            Add pillar
          </Button>
        </Flex>
      </HeaderContainer>

      <Box className="h-[calc(100dvh-4rem)]">
        {isStrategyPending || areObjectivesPending || !strategy ? (
          <BoardSkeleton className="h-full" layout="gantt" />
        ) : (
          <StrategyMapCanvas
            canEdit={!isGuest}
            objectives={objectives}
            onAlign={(objectiveId, pillarId) => {
              if (!isGuest) alignObjective.mutate({ objectiveId, pillarId });
            }}
            onDeletePillar={(pillarId) => {
              if (!isGuest) setPillarToDelete(pillarId);
            }}
            onEditGoal={() => {
              if (!isGuest) setIsGoalOpen(true);
            }}
            onEditPillar={(pillar) => {
              if (!isGuest) setPillarToEdit(pillar);
            }}
            resetSignal={resetSignal}
            showUnaligned={showUnaligned}
            strategy={strategy}
            zoom={zoom}
          />
        )}
      </Box>

      <StrategyEditorDialog
        initialDescription={strategy?.description}
        initialName={strategy?.ultimateGoal}
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
        title="Edit ultimate goal"
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
      <StrategyEditorDialog
        initialDescription={pillarToEdit?.description}
        initialName={pillarToEdit?.name}
        isOpen={Boolean(pillarToEdit)}
        isPending={updatePillar.isPending}
        nameLabel="Pillar name"
        onOpenChange={(isOpen) => {
          if (!isOpen) setPillarToEdit(null);
        }}
        onSave={(name, description) => {
          if (!pillarToEdit) return;
          updatePillar.mutate(
            {
              pillarId: pillarToEdit.id,
              data: { name, description },
            },
            {
              onSuccess: () => {
                setPillarToEdit(null);
              },
            },
          );
        }}
        title="Edit strategic pillar"
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
