import { useParams } from "next/navigation";
import type { FormEvent } from "react";
import React, { useState } from "react";
import type { ButtonProps } from "ui";
import { Button, Dialog, Input, Select, Flex, Box, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { useCreateKeyResultMutation } from "../../hooks";
import type { NewKeyResult, MeasureType } from "../../types";

export const NewKeyResultButton = ({
  color = "tertiary",
  ...rest
}: ButtonProps) => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const keyResultMutation = useCreateKeyResultMutation();
  const [isOpen, setIsOpen] = useState(false);
  const [form, setForm] = useState<NewKeyResult>({
    name: "",
    startValue: 0,
    targetValue: 0,
    measurementType: "number",
  });

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!form.name) {
      toast.warning("Validation error", {
        description: "Please enter a name for the key result",
      });
      return;
    }
    keyResultMutation.mutate({
      objectiveId,
      ...form,
    });
    setIsOpen(false);
    setForm({
      name: "",
      startValue: 0,
      targetValue: 0,
      measurementType: "number",
    });
  };

  return (
    <>
      <Button
        color={color}
        onClick={() => {
          setIsOpen(true);
        }}
        {...rest}
      >
        Add Key Result
      </Button>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content className="max-w-2xl">
          <form onSubmit={handleSubmit}>
            <Dialog.Header className="px-6">
              <Dialog.Title className="text-lg">Create Key Result</Dialog.Title>
            </Dialog.Header>
            <Dialog.Body className="space-y-4">
              <Input
                label="Name"
                onChange={(e) => {
                  setForm({ ...form, name: e.target.value });
                }}
                placeholder="Enter a name for the key result"
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
                    onValueChange={(value) => {
                      const measurementType =
                        value.toLowerCase() as MeasureType;
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
                        {["Number", "Percentage", "Boolean"].map((option) => (
                          <Select.Option
                            key={option}
                            value={option.toLowerCase()}
                          >
                            {option}
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
                    measurementType: "number",
                  });
                }}
                type="button"
              >
                Cancel
              </Button>
              <Button type="submit">Create Key Result</Button>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
