import type { FormEvent } from "react";
import React, { useEffect, useState } from "react";
import { Button, Dialog, Input, Flex, Box, Text, TextArea } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { formatISO } from "date-fns";
import { useMediaQuery, useTerminology } from "@/hooks";
import { useUpdateKeyResultMutation } from "../../hooks";
import type { KeyResult, KeyResultUpdate } from "../../types";

type UpdateKeyResultDialogProps = {
  keyResult: Omit<KeyResult, "createdBy">;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  updateMode: "progress" | "other";
};

export const UpdateKeyResultDialog = ({
  keyResult,
  isOpen,
  onOpenChange,
  updateMode,
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

  const getChangedFields = () => {
    const changes: Partial<KeyResultUpdate> = {};

    if (form.name !== keyResult.name) changes.name = form.name;
    if (form.startValue !== keyResult.startValue)
      changes.startValue = form.startValue;
    if (form.targetValue !== keyResult.targetValue)
      changes.targetValue = form.targetValue;
    if (form.currentValue !== keyResult.currentValue)
      changes.currentValue = form.currentValue;
    if (
      formatISO(new Date(form.startDate), { representation: "date" }) !==
      formatISO(new Date(keyResult.startDate), { representation: "date" })
    ) {
      changes.startDate = form.startDate;
    }
    if (
      formatISO(new Date(form.endDate), { representation: "date" }) !==
      formatISO(new Date(keyResult.endDate), { representation: "date" })
    ) {
      changes.endDate = form.endDate;
    }

    if (
      JSON.stringify(form.contributors) !==
      JSON.stringify(keyResult.contributors)
    ) {
      changes.contributors = form.contributors;
    }

    return changes;
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!form.name) {
      toast.warning("Validation error", {
        description: "Please enter a name for the key result",
      });
      return;
    }

    if (Object.keys(getChangedFields()).length === 0) {
      toast.info("No changes detected", {
        description: "The key result is already up to date",
      });
      return;
    }

    updateMutation.mutate({
      keyResultId: keyResult.id,
      objectiveId: keyResult.objectiveId,
      data: {
        ...getChangedFields(),
        ...(form.comment && { comment: form.comment }),
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

  useEffect(() => {
    if (!isOpen) {
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
    }
  }, [isOpen, keyResult]);

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content
        className={cn("max-w-2xl", { "max-w-xl": updateMode === "progress" })}
      >
        <form onSubmit={handleSubmit}>
          <Dialog.Header className="px-6">
            <Dialog.Title className="text-lg capitalize">
              Update {getTermDisplay("keyResultTerm", { capitalize: true })}{" "}
              progress
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-4 pb-0">
            {updateMode === "other" && (
              <>
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
                        startDate: formatISO(new Date(e.target.value), {
                          representation: "date",
                        }),
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
                        endDate: formatISO(new Date(e.target.value), {
                          representation: "date",
                        }),
                      });
                    }}
                    required
                    type="date"
                    value={formatISO(new Date(form.endDate), {
                      representation: "date",
                    })}
                  />
                </Box>
              </>
            )}

            <Box
              className={cn("grid grid-cols-1 gap-4", {
                "grid-cols-2":
                  updateMode === "other" &&
                  keyResult.measurementType !== "boolean",
              })}
            >
              {keyResult.measurementType === "boolean" &&
                updateMode === "progress" && (
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
                )}
              {keyResult.measurementType !== "boolean" ? (
                <>
                  {updateMode === "progress" ? (
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
                      rightIcon={
                        <Text className="pl-1" color="muted">
                          / {form.targetValue}
                        </Text>
                      }
                      type="number"
                      value={form.currentValue}
                    />
                  ) : (
                    <>
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
                    </>
                  )}
                </>
              ) : null}
            </Box>
            {updateMode === "progress" && (
              <Box className="mt-3">
                <TextArea
                  className="resize-none rounded-2xl border py-4 text-base leading-normal dark:border-dark-50/80 dark:bg-transparent"
                  label="Comment"
                  onChange={(e) => {
                    setForm({ ...form, comment: e.target.value });
                  }}
                  placeholder="Describe what progress was made, any blockers encountered, or next steps..."
                  rows={4}
                  value={form.comment}
                />
              </Box>
            )}
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
            <Button
              className="capitalize"
              disabled={Object.keys(getChangedFields()).length === 0}
              type="submit"
            >
              Update {getTermDisplay("keyResultTerm")}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
