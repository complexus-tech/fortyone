"use client";

import { useState } from "react";
import { Box, BreadCrumbs, Button, Dialog, Flex, Text } from "ui";
import { OKRIcon, PlusIcon } from "icons";
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
  const [showUnaligned, setShowUnaligned] = useState(true);
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
                icon: <OKRIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
              },
            ]}
          />
        </Flex>
        <Flex align="center" className="gap-2">
          <Text
            className="border-border rounded-lg border px-3 py-1.5 text-[0.95rem]"
            color="muted"
          >
            All teams
          </Text>
          <Text
            className="border-border rounded-lg border px-3 py-1.5 text-[0.95rem]"
            color="muted"
          >
            {new Date().getFullYear()}
          </Text>
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

      <Flex
        align="center"
        className="border-border h-12 border-b px-5"
        justify="between"
      >
        <Flex align="center" className="gap-4">
          <Text fontWeight="medium">Strategy hierarchy</Text>
          <Text className="text-[0.95rem]" color="muted">
            Goal → pillars → objectives → key results
          </Text>
        </Flex>
        <button
          aria-pressed={showUnaligned}
          className="text-text-secondary hover:text-text-primary flex items-center gap-2 font-medium"
          onClick={() => {
            setShowUnaligned((current) => !current);
          }}
          type="button"
        >
          <span
            className={`border-border inline-flex h-5 w-9 items-center rounded-full border p-0.5 transition-colors ${
              showUnaligned ? "bg-foreground" : "bg-surface-muted"
            }`}
          >
            <span
              className={`bg-background h-3.5 w-3.5 rounded-full transition-transform ${
                showUnaligned ? "translate-x-3.5" : "translate-x-0"
              }`}
            />
          </span>
          Show unaligned
        </button>
      </Flex>

      <Box className="h-[calc(100dvh-7rem)]">
        {isStrategyPending || areObjectivesPending || !strategy ? (
          <BoardSkeleton className="h-full" layout="gantt" />
        ) : (
          <StrategyMapCanvas
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
            showUnaligned={showUnaligned}
            strategy={strategy}
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
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Delete strategic pillar?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
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
