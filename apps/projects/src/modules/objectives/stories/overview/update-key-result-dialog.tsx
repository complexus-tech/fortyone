import type { FormEvent } from "react";
import React, { useEffect, useState } from "react";
import { Button, Dialog, Input, Flex, Box, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { useUpdateKeyResultMutation } from "../../hooks";
import type { KeyResult } from "../../types";

type UpdateKeyResultDialogProps = {
  keyResult: Omit<KeyResult, "createdBy" | "lastUpdatedBy">;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
};

export const UpdateKeyResultDialog = ({
  keyResult,
  isOpen,
  onOpenChange,
}: UpdateKeyResultDialogProps) => {
  const updateMutation = useUpdateKeyResultMutation();
  const [form, setForm] = useState({
    name: keyResult.name,
    startValue: keyResult.startValue,
    targetValue: keyResult.targetValue,
    currentValue: keyResult.currentValue,
  });

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!form.name) {
      toast.warning("Validation error", {
        description: "Please enter a name for the key result",
      });
      return;
    }

    updateMutation.mutate({
      keyResultId: keyResult.id,
      objectiveId: keyResult.objectiveId,
      data: {
        name: form.name,
        startValue: form.startValue,
        targetValue: form.targetValue,
        currentValue: form.currentValue,
      },
    });
    onOpenChange(false);
    setForm({
      name: keyResult.name,
      startValue: keyResult.startValue,
      targetValue: keyResult.targetValue,
      currentValue: keyResult.currentValue,
    });
  };

  useEffect(() => {
    setForm({
      name: keyResult.name,
      startValue: keyResult.startValue,
      targetValue: keyResult.targetValue,
      currentValue: keyResult.currentValue,
    });
  }, [keyResult]);

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content>
        <form onSubmit={handleSubmit}>
          <Dialog.Header className="px-6">
            <Dialog.Title className="text-lg">Update Key Result</Dialog.Title>
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
            <Box className="grid grid-cols-3 gap-4">
              {keyResult.measurementType === "boolean" ? (
                <Box className="col-span-3">
                  <Text className="mb-[0.35rem]">Current Value</Text>
                  <Flex
                    className="rounded-[0.45rem] border bg-white/70 p-1 dark:border-dark-50/80 dark:bg-dark/20"
                    gap={1}
                  >
                    <Button
                      align="center"
                      className={cn("rounded-[0.35rem] border-0", {
                        "bg-transparent dark:bg-transparent":
                          form.currentValue !== 0,
                      })}
                      color={form.currentValue === 0 ? "primary" : "tertiary"}
                      fullWidth
                      onClick={() => {
                        setForm({ ...form, currentValue: 0 });
                      }}
                      size="sm"
                      type="button"
                      variant={form.currentValue === 0 ? "solid" : "outline"}
                    >
                      Incomplete
                    </Button>
                    <Button
                      align="center"
                      className={cn("rounded-[0.35rem] border-0", {
                        "bg-transparent dark:bg-transparent":
                          form.currentValue !== 1,
                      })}
                      color={form.currentValue === 1 ? "primary" : "tertiary"}
                      fullWidth
                      onClick={() => {
                        setForm({ ...form, currentValue: 1 });
                      }}
                      size="sm"
                      type="button"
                      variant={form.currentValue === 1 ? "solid" : "outline"}
                    >
                      Complete
                    </Button>
                  </Flex>
                </Box>
              ) : (
                <>
                  <Input
                    className="h-[2.7rem]"
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
                    label="Current Value"
                    onChange={(e) => {
                      setForm({
                        ...form,
                        currentValue: Number(e.target.value),
                      });
                    }}
                    placeholder="0"
                    required
                    type="number"
                    value={form.currentValue}
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
                onOpenChange(false);
              }}
              type="button"
            >
              Cancel
            </Button>
            <Button type="submit">Update Key Result</Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
