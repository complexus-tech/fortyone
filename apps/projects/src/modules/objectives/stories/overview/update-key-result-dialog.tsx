import type { FormEvent } from "react";
import React, { useEffect, useState } from "react";
import { Button, Dialog, Input, Flex, Box, Text, TextArea } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { formatISO } from "date-fns";
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
    startDate: keyResult.startDate,
    endDate: keyResult.endDate,
    comment: "",
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
        startDate: form.startDate,
        endDate: form.endDate,
      },
    });
    onOpenChange(false);
    setForm({
      name: keyResult.name,
      startValue: keyResult.startValue,
      targetValue: keyResult.targetValue,
      currentValue: keyResult.currentValue,
      contributors: keyResult.contributors,
      startDate: keyResult.startDate,
      endDate: keyResult.endDate,
      comment: "",
    });
  };

  useEffect(() => {
    setForm({
      name: keyResult.name,
      startValue: keyResult.startValue,
      targetValue: keyResult.targetValue,
      currentValue: keyResult.currentValue,
      contributors: keyResult.contributors,
      startDate: keyResult.startDate,
      endDate: keyResult.endDate,
      comment: "",
    });
  }, [keyResult]);

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content className="max-w-2xl">
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
            <Box className="grid grid-cols-2 gap-4">
              <Input
                label="Start Date"
                onChange={(e) => {
                  setForm({
                    ...form,
                    startDate: new Date(e.target.value).toISOString(),
                  });
                }}
                required
                type="date"
                value={formatISO(new Date(form.startDate), {
                  representation: "date",
                })}
              />
              <Input
                label="Deadline"
                onChange={(e) => {
                  setForm({
                    ...form,
                    endDate: new Date(e.target.value).toISOString(),
                  });
                }}
                required
                type="date"
                value={formatISO(new Date(form.endDate), {
                  representation: "date",
                })}
              />
            </Box>
            <Box className="grid grid-cols-3 gap-4">
              {keyResult.measurementType === "boolean" ? (
                <Box className="col-span-3">
                  <Text className="mb-[0.35rem]">Current Value</Text>
                  <Flex
                    className="rounded-xl border border-gray-100 bg-white/70 p-1 dark:border-dark-50/80 dark:bg-dark/20"
                    gap={1}
                  >
                    <Button
                      align="center"
                      className={cn("border-0", {
                        "bg-transparent dark:bg-transparent":
                          form.currentValue !== 0,
                      })}
                      color={form.currentValue === 0 ? "primary" : "tertiary"}
                      fullWidth
                      onClick={() => {
                        setForm({ ...form, currentValue: 0 });
                      }}
                      type="button"
                      variant={form.currentValue === 0 ? "solid" : "outline"}
                    >
                      Incomplete
                    </Button>
                    <Button
                      align="center"
                      className={cn("border-0", {
                        "bg-transparent dark:bg-transparent":
                          form.currentValue !== 1,
                      })}
                      color={form.currentValue === 1 ? "primary" : "tertiary"}
                      fullWidth
                      onClick={() => {
                        setForm({ ...form, currentValue: 1 });
                      }}
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
            <Box className="mt-3">
              <TextArea
                className="resize-none rounded-2xl border py-4 text-base leading-normal dark:border-dark-50/80 dark:bg-transparent"
                label="Comment"
                onChange={(e) => {
                  setForm({ ...form, comment: e.target.value });
                }}
                placeholder="Write your update here..."
                rows={4}
                value={form.comment}
              />
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
