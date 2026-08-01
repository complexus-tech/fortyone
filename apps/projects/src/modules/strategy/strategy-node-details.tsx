"use client";

import type { FormEvent, ReactNode } from "react";
import { useState } from "react";
import { CloseIcon, ObjectiveIcon, StrategyIcon } from "icons";
import { Box, Button, Divider, Flex, Input, Text, TextArea } from "ui";

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
  const [description, setDescription] = useState(initialDescription ?? "");
  const normalizedInitialDescription = initialDescription?.trim() || null;
  const hasChanges =
    name.trim() !== initialName.trim() ||
    (description.trim() || null) !== normalizedInitialDescription;
  const heading = kind === "goal" ? "Ultimate goal" : "Strategic pillar";

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedName = name.trim();
    if (!canEdit || !trimmedName || !hasChanges) return;
    onSave(trimmedName, description.trim() || null);
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

      <form onSubmit={handleSubmit}>
        <Box className="px-6 pt-6 pb-24">
          <Input
            aria-label={`${heading} name`}
            className="h-auto border-0 bg-transparent px-0 py-1 text-2xl leading-8 font-semibold shadow-none focus-visible:ring-0 dark:bg-transparent"
            maxLength={200}
            onChange={(event) => {
              setName(event.target.value);
            }}
            placeholder={`Name this ${kind}`}
            readOnly={!canEdit}
            value={name}
          />
          <TextArea
            aria-label={`${heading} description`}
            className="mt-3 min-h-32 resize-none border-0 bg-transparent px-0 py-1.5 text-[1.05rem] leading-6 shadow-none focus-visible:ring-0 dark:bg-transparent"
            maxLength={1000}
            onChange={(event) => {
              setDescription(event.target.value);
            }}
            placeholder="Add a description..."
            readOnly={!canEdit}
            rows={5}
            value={description}
          />

          <Divider className="my-5 opacity-60" />

          <Flex align="center" className="gap-5" wrap>
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

          {canEdit ? (
            <Flex className="mt-8" justify="end">
              <Button
                color="invert"
                disabled={isPending || !name.trim() || !hasChanges}
                type="submit"
              >
                {isPending ? "Saving..." : "Save changes"}
              </Button>
            </Flex>
          ) : null}
        </Box>
      </form>
    </Box>
  );
};

export const StrategyNodeDetails = ({
  entityKey,
  ...props
}: StrategyNodeDetailsProps) => (
  <StrategyNodeDetailsForm key={entityKey} {...props} />
);
