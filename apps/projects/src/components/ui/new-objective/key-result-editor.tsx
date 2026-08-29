import { Box, Button, Flex, Input, Select, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import type { FormEvent } from "react";
import { formatISO } from "date-fns";
import type { NewKeyResult, MeasureType } from "@/modules/objectives/types";
import { useTerminology } from "@/hooks";
import { OkrQualityBanner } from "@/modules/objectives/components/okr-quality-banner";
import { useOkrQualityAssessment } from "@/modules/objectives/hooks/use-okr-quality-assessment";

type KeyResultEditorProps = {
  keyResult: NewKeyResult | null;
  onUpdate: (index: number, updates: Partial<NewKeyResult>) => void;
  onCancel: () => void;
  onSave: () => void;
  qualityContext?: {
    objectiveName: string;
    objectiveStartDate: string | null;
    objectiveEndDate: string | null;
    existingKeyResults: NewKeyResult[];
  };
};

export const KeyResultEditor = ({
  keyResult,
  onUpdate,
  onCancel,
  onSave,
  qualityContext,
}: KeyResultEditorProps) => {
  const { getTermDisplay } = useTerminology();
  const qualityRequest =
    keyResult?.name.trim() && qualityContext?.objectiveName.trim()
      ? {
          kind: "key_result" as const,
          draft: {
            name: keyResult.name,
            measurementType: keyResult.measurementType,
            startValue: keyResult.startValue,
            targetValue: keyResult.targetValue,
            startDate: keyResult.startDate || null,
            endDate: keyResult.endDate || null,
          },
          objective: {
            id: "draft-objective",
            name: qualityContext.objectiveName,
            startDate: qualityContext.objectiveStartDate,
            endDate: qualityContext.objectiveEndDate,
          },
          existingKeyResults: qualityContext.existingKeyResults.map(
            (existingKeyResult, index) => ({
              id: `draft-key-result-${index}`,
              name: existingKeyResult.name,
            }),
          ),
        }
      : null;
  const { assessment, isAssessing } = useOkrQualityAssessment(qualityRequest);
  const measurementTypes: { label: string; value: MeasureType }[] = [
    {
      label: "Number",
      value: "number",
    },
    {
      label: "Percentage (0-100%)",
      value: "percentage",
    },
    {
      label: "Boolean (Complete/Incomplete)",
      value: "boolean",
    },
  ];

  const handleSave = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!keyResult?.name) {
      toast.warning("Validation error", {
        description: `Please enter a name for the ${getTermDisplay("keyResultTerm")}`,
      });
      return;
    }
    if (keyResult.measurementType !== "boolean" && !keyResult.targetValue) {
      toast.warning("Validation error", {
        description: `Please enter a target value for the ${getTermDisplay("keyResultTerm")}`,
      });
      return;
    }
    onSave();
  };

  if (!keyResult) return null;

  return (
    <form
      className="border-border mb-6 space-y-4 rounded-2xl border px-5 py-4"
      onSubmit={handleSave}
    >
      <Input
        label="Name"
        onChange={(e) => {
          onUpdate(0, { name: e.target.value });
        }}
        placeholder="eg. Increase user adoption from 100 to 150"
        required
        value={keyResult.name}
      />
      <OkrQualityBanner
        assessment={assessment}
        isAssessing={isAssessing}
        onUseSuggestion={(suggestion) => {
          onUpdate(0, { name: suggestion });
        }}
      />
      <Box className="grid grid-cols-2 gap-4">
        <Input
          label="Start Date"
          onChange={(e) => {
            onUpdate(0, {
              startDate: formatISO(new Date(e.target.value), {
                representation: "date",
              }),
            });
          }}
          required
          type="date"
          value={keyResult.startDate}
        />
        <Input
          label="Deadline"
          onChange={(e) => {
            onUpdate(0, {
              endDate: formatISO(new Date(e.target.value), {
                representation: "date",
              }),
            });
          }}
          required
          type="date"
          value={keyResult.endDate}
        />
      </Box>

      <Box
        className={cn("grid grid-cols-3 gap-4", {
          "grid-cols-2": keyResult.measurementType === "boolean",
        })}
      >
        <Box>
          <Text className="mb-1.5 font-medium">Measure as</Text>
          <Select
            defaultValue={keyResult.measurementType}
            onValueChange={(measurementType: MeasureType) => {
              onUpdate(0, {
                measurementType,
                startValue: measurementType === "boolean" ? 0 : 0,
                targetValue: measurementType === "boolean" ? 1 : 0,
              });
            }}
          >
            <Select.Trigger className="bg-surface/70 h-[2.7rem] text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content defaultValue="number">
              <Select.Group>
                {measurementTypes.map((option) => (
                  <Select.Option key={option.value} value={option.value}>
                    {option.label}
                  </Select.Option>
                ))}
              </Select.Group>
            </Select.Content>
          </Select>
        </Box>

        {keyResult.measurementType === "boolean" ? (
          <Box>
            <Text className="mb-1.5 font-medium">Current status</Text>
            <Flex
              className="border-border bg-surface-muted rounded-lg border p-1"
              gap={1}
            >
              <Button
                align="center"
                className={cn("rounded-md border-0", {
                  "bg-transparent dark:bg-transparent":
                    keyResult.startValue !== 0,
                })}
                color={keyResult.startValue === 0 ? "primary" : "tertiary"}
                fullWidth
                onClick={() => {
                  onUpdate(0, { startValue: 0 });
                }}
                size="sm"
                type="button"
                variant={keyResult.startValue === 0 ? "solid" : "outline"}
              >
                Incomplete
              </Button>
              <Button
                align="center"
                className={cn("rounded-md border-0", {
                  "bg-transparent dark:bg-transparent":
                    keyResult.startValue !== 1,
                })}
                color={keyResult.startValue === 1 ? "primary" : "tertiary"}
                fullWidth
                onClick={() => {
                  onUpdate(0, { startValue: 1 });
                }}
                size="sm"
                type="button"
                variant={keyResult.startValue === 1 ? "solid" : "outline"}
              >
                Complete
              </Button>
            </Flex>
          </Box>
        ) : (
          <>
            <Input
              label="Starting Value"
              max={keyResult.measurementType === "percentage" ? 100 : undefined}
              min={keyResult.measurementType === "percentage" ? 0 : undefined}
              onChange={(e) => {
                onUpdate(0, { startValue: Number(e.target.value) });
              }}
              placeholder="0"
              required
              type="number"
              value={keyResult.startValue}
            />
            <Input
              label="Target Value"
              max={keyResult.measurementType === "percentage" ? 100 : undefined}
              min={keyResult.measurementType === "percentage" ? 0 : undefined}
              onChange={(e) => {
                onUpdate(0, { targetValue: Number(e.target.value) });
              }}
              placeholder="0"
              required
              type="number"
              value={keyResult.targetValue}
            />
          </>
        )}
      </Box>
      <Flex gap={2}>
        <Button className="capitalize" type="submit">
          Add {getTermDisplay("keyResultTerm", { capitalize: true })}
        </Button>
        <Button
          className="px-7"
          color="tertiary"
          onClick={onCancel}
          type="button"
        >
          Cancel
        </Button>
      </Flex>
    </form>
  );
};
