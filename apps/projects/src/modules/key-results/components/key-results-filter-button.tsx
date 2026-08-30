"use client";

import { ArrowDownIcon, FilterIcon } from "icons";
import { Button, Divider, Flex, Popover, Text } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import { getActiveKeyResultFilterCount } from "../key-results-filter-utils";
import type { KeyResultFilters } from "../types";
import {
  KeyResultsFilterSection,
  KeyResultsFilterValueEditor,
  useKeyResultsFilterData,
} from "./key-results-filter-controls";

export const KeyResultsFilterButton = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: (filters: KeyResultFilters) => void;
}) => {
  const { members, objectives, teams } = useKeyResultsFilterData();
  const { getTermDisplay } = useTerminology();
  const activeFilterCount = getActiveKeyResultFilterCount(filters);
  const keyResultLabel = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const objectiveLabel = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const buttonLabel =
    activeFilterCount > 0
      ? `${activeFilterCount} filter${activeFilterCount === 1 ? "" : "s"} applied`
      : "Filters";
  const filterEditorProps = {
    filters,
    members,
    objectiveLabel,
    objectives,
    setFilters,
    teams,
  };

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          aria-label={buttonLabel}
          className="relative"
          color="tertiary"
          leftIcon={<FilterIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          {activeFilterCount > 0 ? (
            <span
              aria-hidden="true"
              className="bg-primary absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full"
            >
              <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-75" />
            </span>
          ) : null}
          <span className="hidden md:inline">{buttonLabel}</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        className="bg-surface-elevated dark:bg-surface-elevated/80 mr-0 max-h-[87vh] w-80 overflow-y-auto rounded-2xl pb-2 md:w-140"
      >
        <Flex align="center" className="h-11 px-4" justify="between">
          <Text
            color="muted"
            fontSize="sm"
            fontWeight="semibold"
            transform="uppercase"
          >
            Apply Filters
          </Text>
          {activeFilterCount > 0 ? (
            <Button
              className="text-primary dark:text-primary"
              color="tertiary"
              onClick={() => {
                setFilters({});
              }}
              size="sm"
              variant="naked"
            >
              Clear filters
            </Button>
          ) : null}
        </Flex>
        <Divider className="mt-1.5" />
        <KeyResultsFilterSection title="Lead">
          <KeyResultsFilterValueEditor {...filterEditorProps} field="leadIds" />
        </KeyResultsFilterSection>
        <Divider />
        <KeyResultsFilterSection title="Delivery date">
          <KeyResultsFilterValueEditor {...filterEditorProps} field="endDate" />
        </KeyResultsFilterSection>
        <Divider />
        <KeyResultsFilterSection title="Measurement type">
          <KeyResultsFilterValueEditor
            {...filterEditorProps}
            field="measurementTypes"
          />
        </KeyResultsFilterSection>
        {teams.length > 0 ? (
          <>
            <Divider />
            <KeyResultsFilterSection title="Team">
              <KeyResultsFilterValueEditor
                {...filterEditorProps}
                field="teamIds"
              />
            </KeyResultsFilterSection>
          </>
        ) : null}
        {objectives.length > 0 ? (
          <>
            <Divider />
            <KeyResultsFilterSection
              title={getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              })}
            >
              <KeyResultsFilterValueEditor
                {...filterEditorProps}
                field="objectiveIds"
              />
            </KeyResultsFilterSection>
          </>
        ) : null}
        <Text className="sr-only">Filter {keyResultLabel}</Text>
      </Popover.Content>
    </Popover>
  );
};
