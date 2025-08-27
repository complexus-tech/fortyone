import { Box, Button, Flex, Input, Select, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import type { FormEvent } from "react";
import { formatISO } from "date-fns";
import type { NewKeyResult, MeasureType } from "@/modules/objectives/types";
import { useTerminology } from "@/hooks";

type KeyResultEditorProps = {
  keyResult: NewKeyResult | null;
  onUpdate: (index: number, updates: Partial<NewKeyResult>) => void;
  onCancel: () => void;
  onSave: () => void;
};

export const KeyResultEditor = ({
  keyResult,
  onUpdate,
  onCancel,
  onSave,
}: KeyResultEditorProps) => {
  const { getTermDisplay } = useTerminology();
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
      className="mb-6 space-y-4 rounded-2xl border border-gray-100/80 px-5 py-4 dark:border-dark-100"
      onSubmit={handleSave}
    >
      <Input
        autoFocus
        label="Name"
        onChange={(e) => {
          onUpdate(0, { name: e.target.value });
        }}
        placeholder="eg. Increase user adoption from 100 to 150"
        required
        value={keyResult.name}
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
            <Select.Trigger className="h-[2.7rem] bg-white/70 text-base dark:bg-dark/20">
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
              className="rounded-[0.6rem] border p-1 dark:border-dark-50/80 dark:bg-dark-100/50"
              gap={1}
            >
              <Button
                align="center"
                className={cn("rounded-[0.4rem] border-0", {
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
                className={cn("rounded-[0.4rem] border-0", {
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
