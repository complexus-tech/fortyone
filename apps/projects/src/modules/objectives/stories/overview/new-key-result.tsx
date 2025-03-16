import { useParams } from "next/navigation";
import type { FormEvent } from "react";
import React, { useState } from "react";
import type { ButtonProps } from "ui";
import { Button, Dialog, Input, Select, Flex, Box, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { useTerminology } from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useCreateKeyResultMutation, useObjective } from "../../hooks";
import type { NewKeyResult, MeasureType } from "../../types";

export const NewKeyResultButton = ({
  color = "tertiary",
  ...rest
}: ButtonProps) => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: objective } = useObjective(objectiveId);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const keyResultMutation = useCreateKeyResultMutation();
  const [isOpen, setIsOpen] = useState(false);
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
  const [form, setForm] = useState<NewKeyResult>({
    name: "",
    startValue: 0,
    targetValue: 0,
    currentValue: 0,
    measurementType: "number",
  });

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!form.name) {
      toast.warning("Validation error", {
        description: `Please enter a name for the ${getTermDisplay("keyResultTerm")}`,
      });
      return;
    }

    if (form.measurementType === "percentage") {
      if (
        form.startValue < 0 ||
        form.startValue > 100 ||
        form.targetValue < 0 ||
        form.targetValue > 100
      ) {
        toast.warning("Validation error", {
          description: "Percentage values must be between 0 and 100",
        });
        return;
      }
    }

    keyResultMutation.mutate({
      objectiveId,
      ...form,
      currentValue: form.measurementType === "boolean" ? 0 : form.startValue,
    });
    setIsOpen(false);
    setForm({
      name: "",
      startValue: 0,
      targetValue: 0,
      currentValue: 0,
      measurementType: "number",
    });
  };

  return (
    <>
      {isAdminOrOwner ? (
        <Button
          color={color}
          onClick={() => {
            setIsOpen(true);
          }}
          {...rest}
        >
          Add {getTermDisplay("keyResultTerm", { capitalize: true })}
        </Button>
      ) : null}
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content className="max-w-2xl">
          <form onSubmit={handleSubmit}>
            <Dialog.Header className="px-6">
              <Dialog.Title className="text-lg capitalize">
                Create {getTermDisplay("keyResultTerm", { capitalize: true })}
              </Dialog.Title>
            </Dialog.Header>
            <Dialog.Body className="space-y-4">
              <Input
                label="Name"
                onChange={(e) => {
                  setForm({ ...form, name: e.target.value });
                }}
                placeholder={`Enter a name for the ${getTermDisplay("keyResultTerm")}`}
                required
                value={form.name}
              />
              <Box
                className={cn("grid grid-cols-3 gap-4", {
                  "grid-cols-7": form.measurementType === "boolean",
                })}
              >
                <Box
                  className={cn({
                    "col-span-3": form.measurementType === "boolean",
                  })}
                >
                  <Text className="mb-[0.35rem]">Measurement Type</Text>
                  <Select
                    defaultValue={form.measurementType}
                    onValueChange={(measurementType: MeasureType) => {
                      setForm({
                        ...form,
                        measurementType,
                        startValue: measurementType === "boolean" ? 0 : 0,
                        targetValue: measurementType === "boolean" ? 1 : 0,
                      });
                    }}
                  >
                    <Select.Trigger className="h-[2.7rem] bg-white/70 text-base dark:bg-dark/20">
                      <Select.Input placeholder="Select measurement type" />
                    </Select.Trigger>
                    <Select.Content>
                      <Select.Group>
                        {measurementTypes.map((option) => (
                          <Select.Option
                            key={option.value}
                            value={option.value}
                          >
                            {option.label}
                          </Select.Option>
                        ))}
                      </Select.Group>
                    </Select.Content>
                  </Select>
                </Box>
                {form.measurementType === "boolean" ? (
                  <Box className="col-span-4">
                    <Text className="mb-[0.35rem]">Current Value</Text>
                    <Flex
                      className="rounded-[0.45rem] border bg-white/70 p-1 dark:border-dark-50/80 dark:bg-dark/20"
                      gap={1}
                    >
                      <Button
                        align="center"
                        className={cn("rounded-[0.35rem] border-0", {
                          "bg-transparent dark:bg-transparent":
                            form.startValue !== 0,
                        })}
                        color={form.startValue === 0 ? "primary" : "tertiary"}
                        fullWidth
                        onClick={() => {
                          setForm({ ...form, startValue: 0 });
                        }}
                        size="sm"
                        type="button"
                        variant={form.startValue === 0 ? "solid" : "outline"}
                      >
                        Incomplete
                      </Button>
                      <Button
                        align="center"
                        className={cn("rounded-[0.35rem] border-0", {
                          "bg-transparent dark:bg-transparent":
                            form.startValue !== 1,
                        })}
                        color={form.startValue === 1 ? "primary" : "tertiary"}
                        fullWidth
                        onClick={() => {
                          setForm({ ...form, startValue: 1 });
                        }}
                        size="sm"
                        type="button"
                        variant={form.startValue === 1 ? "solid" : "outline"}
                      >
                        Complete
                      </Button>
                    </Flex>
                  </Box>
                ) : (
                  <>
                    <Input
                      className="h-[2.7rem] bg-gray-50/30"
                      label="Starting Value"
                      max={
                        form.measurementType === "percentage" ? 100 : undefined
                      }
                      min={
                        form.measurementType === "percentage" ? 0 : undefined
                      }
                      onChange={(e) => {
                        setForm({
                          ...form,
                          startValue: Number(e.target.value),
                        });
                      }}
                      placeholder="0"
                      required
                      type="number"
                      value={form.startValue}
                    />
                    <Input
                      className="h-[2.7rem]"
                      label="Target Value"
                      max={
                        form.measurementType === "percentage" ? 100 : undefined
                      }
                      min={
                        form.measurementType === "percentage" ? 0 : undefined
                      }
                      onChange={(e) => {
                        setForm({
                          ...form,
                          targetValue: Number(e.target.value),
                        });
                      }}
                      placeholder="0"
                      required
                      type="number"
                      value={form.targetValue}
                    />
                  </>
                )}
              </Box>
            </Dialog.Body>
            <Dialog.Footer className="justify-end gap-2 border-0">
              <Button
                className="px-6"
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                  setForm({
                    name: "",
                    startValue: 0,
                    targetValue: 0,
                    currentValue: 0,
                    measurementType: "number",
                  });
                }}
                type="button"
              >
                Cancel
              </Button>
              <Button className="capitalize" type="submit">
                Create {getTermDisplay("keyResultTerm", { capitalize: true })}
              </Button>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
