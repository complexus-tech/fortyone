import type { FormEvent } from "react";
import React, { useEffect, useState } from "react";
import { Button, Dialog, Input, Flex, Box, Text } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { useMediaQuery, useTerminology } from "@/hooks";
import { useUpdateKeyResultMutation } from "../../hooks";
import type { KeyResult } from "../../types";

type UpdateKeyResultDialogProps = {
  keyResult: Omit<KeyResult, "createdBy">;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
};

export const UpdateKeyResultDialog = ({
  keyResult,
  isOpen,
  onOpenChange,
}: UpdateKeyResultDialogProps) => {
  const { getTermDisplay } = useTerminology();
  const updateMutation = useUpdateKeyResultMutation();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [form, setForm] = useState({
    name: keyResult.name,
    startValue: keyResult.startValue,
    targetValue: keyResult.targetValue,
    currentValue: keyResult.currentValue,
    contributors: keyResult.contributors,
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
        contributors: form.contributors,
      },
    });
    onOpenChange(false);
    setForm({
      name: keyResult.name,
      startValue: keyResult.startValue,
      targetValue: keyResult.targetValue,
      currentValue: keyResult.currentValue,
      contributors: keyResult.contributors,
    });
  };

  useEffect(() => {
    setForm({
      name: keyResult.name,
      startValue: keyResult.startValue,
      targetValue: keyResult.targetValue,
      currentValue: keyResult.currentValue,
      contributors: keyResult.contributors,
    });
  }, [keyResult]);

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content>
        <form onSubmit={handleSubmit}>
          <Dialog.Header className="px-6">
            <Dialog.Title className="text-lg capitalize">
              Update {getTermDisplay("keyResultTerm", { capitalize: true })}{" "}
              progress
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
                  <Box className="opacity-40 focus-within:opacity-80">
                    <Input
                      className="h-[2.7rem]"
                      label={isMobile ? "Starting" : "Starting Value"}
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
                  </Box>
                  <Input
                    className="h-[2.7rem]"
                    label={isMobile ? "Current" : "Current Value"}
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
                  <Box className="opacity-40 focus-within:opacity-80">
                    <Input
                      className="h-[2.7rem]"
                      label={isMobile ? "Target" : "Target Value"}
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
                  </Box>
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
            <Button className="capitalize" type="submit">
              Update {getTermDisplay("keyResultTerm")}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
