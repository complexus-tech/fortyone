"use client";

import type { ReactNode } from "react";
import { useRef, useState } from "react";
import { CloseIcon, ObjectiveIcon, StrategyIcon } from "icons";
import { Box, Button, Flex, Input, Text } from "ui";
import { useDebouncedCallback } from "@/hooks/debounce";
import { StrategyDescriptionEditor } from "./strategy-description-editor";

const AUTOSAVE_DELAY = 1000;

type StrategyDetailsDraft = {
  description: string | null;
  name: string;
};

type StrategyNodeDetailsProps = {
  canEdit: boolean;
  description: string | null;
  entityKey: string;
  isPending: boolean;
  kind: "goal" | "pillar";
  name: string;
  objectiveCount: number;
  onClose: () => void;
  onSave: (name: string, description: string | null) => void;
  pillarCount?: number;
};

type StrategyNodeDetailsFormProps = Omit<StrategyNodeDetailsProps, "entityKey">;

const DetailMetric = ({
  icon,
  label,
  value,
}: {
  icon: ReactNode;
  label: string;
  value: number;
}) => (
  <Flex align="center" className="gap-2">
    {icon}
    <Text color="muted">
      {value} {label}
    </Text>
  </Flex>
);

const StrategyNodeDetailsForm = ({
  canEdit,
  description: initialDescription,
  isPending,
  kind,
  name: initialName,
  objectiveCount,
  onClose,
  onSave,
  pillarCount,
}: StrategyNodeDetailsFormProps) => {
  const [name, setName] = useState(initialName);
  const nameRef = useRef(initialName);
  const descriptionRef = useRef(initialDescription);
  const heading = kind === "goal" ? "Ultimate goal" : "Strategic pillar";
  const { callback: queueSave, flush: flushSave } =
    useDebouncedCallback<StrategyDetailsDraft>(
      (draft) => {
        onSave(draft.name, draft.description);
      },
      AUTOSAVE_DELAY,
      { flushOnUnmount: true },
    );

  const scheduleSave = (nextName: string, nextDescription: string | null) => {
    const trimmedName = nextName.trim();
    if (!canEdit || !trimmedName) return;

    queueSave({
      description: nextDescription,
      name: trimmedName,
    });
  };

  return (
    <Box className="border-border/70 dark:border-border dark:bg-surface absolute top-14 right-3 bottom-4 isolate z-40 w-[calc(100%-1.5rem)] overflow-y-auto rounded-xl border bg-white shadow-xl md:top-[1.625rem] md:right-6 md:bottom-[4.875rem] md:w-[34rem]">
      <Flex
        align="center"
        className="border-border/70 dark:border-border dark:bg-surface/80 sticky top-0 z-10 min-h-16 gap-6 border-b-[0.5px] bg-white/80 px-6 backdrop-blur-2xl"
        justify="between"
      >
        <Flex align="center" className="min-w-0 gap-3">
          <StrategyIcon className="text-text-muted h-5 shrink-0" />
          <Text className="truncate" fontSize="lg" fontWeight="semibold">
            {heading}
          </Text>
        </Flex>
        <Flex align="center" className="gap-2">
          {isPending ? (
            <Text className="text-[0.9rem]" color="muted">
              Saving...
            </Text>
          ) : null}
          <Button
            aria-label={`Close ${heading.toLowerCase()} details`}
            className="-mr-2"
            color="tertiary"
            leftIcon={<CloseIcon className="h-4" strokeWidth={3} />}
            onClick={onClose}
            size="sm"
            variant="naked"
          />
        </Flex>
      </Flex>

      <Box className="px-6 pt-6 pb-24">
        <Flex align="center" className="mb-3 gap-5" wrap>
          {kind === "goal" && pillarCount !== undefined ? (
            <DetailMetric
              icon={<StrategyIcon className="text-text-muted h-4 w-4" />}
              label={`pillar${pillarCount === 1 ? "" : "s"}`}
              value={pillarCount}
            />
          ) : null}
          <DetailMetric
            icon={<ObjectiveIcon className="text-text-muted h-4 w-4" />}
            label={`objective${objectiveCount === 1 ? "" : "s"}`}
            value={objectiveCount}
          />
        </Flex>
        <Input
          aria-label={`${heading} name`}
          className="h-auto border-0 bg-transparent px-0 py-1 text-2xl leading-8 font-semibold shadow-none focus-visible:ring-0 dark:bg-transparent"
          maxLength={200}
          onBlur={flushSave}
          onChange={(event) => {
            const nextName = event.target.value;
            nameRef.current = nextName;
            setName(nextName);
            scheduleSave(nextName, descriptionRef.current);
          }}
          placeholder={`Name this ${kind}`}
          readOnly={!canEdit}
          value={name}
        />
        <StrategyDescriptionEditor
          ariaLabel={`${heading} description`}
          className="mt-1.5 min-h-40"
          content={initialDescription ?? ""}
          contentClassName="min-h-40"
          editable={canEdit}
          onBlur={flushSave}
          onChange={(nextDescription) => {
            descriptionRef.current = nextDescription;
            scheduleSave(nameRef.current, nextDescription);
          }}
          placeholder="Add a description..."
        />
      </Box>
    </Box>
  );
};

export const StrategyNodeDetails = ({
  entityKey,
  ...props
}: StrategyNodeDetailsProps) => (
  <StrategyNodeDetailsForm key={entityKey} {...props} />
);
