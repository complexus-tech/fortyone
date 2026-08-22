import {
  Text,
  Box,
  Flex,
  Wrapper,
  Button,
  ProgressBar,
  Menu,
  Badge,
  Divider,
  Checkbox,
  Tooltip,
  Avatar,
} from "ui";
import {
  AiIcon,
  DeleteIcon,
  EditIcon,
  MoreHorizontalIcon,
  OKRIcon,
  InfoIcon,
  CalendarIcon,
} from "icons";
import { useParams, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { experimental_useObject as useObject } from "@ai-sdk/react";
import { differenceInDays, format } from "date-fns";
import { cn } from "lib";
import { toast } from "sonner";
import { ConfirmDialog, RowWrapper, AssigneesMenu } from "@/components/ui";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useMediaQuery, useTerminology } from "@/hooks";
import { Thinking } from "@/components/ui/chat/thinking";
import { RoadmapEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import {
  useCreateKeyResultsMutation,
  useKeyResults,
  useObjective,
  useUpdateKeyResultMutation,
} from "@/modules/objectives/hooks";
import { useMembers } from "@/lib/hooks/members";
import { toKeyResultCreateInput } from "@/modules/objectives/key-result-generation";
import type { KeyResult } from "../../types";
import { useDeleteKeyResultMutation } from "../../hooks/use-delete-key-result-mutation";
import { keyResultGenerationSchema } from "../../schemas/key-result-generation";
import { NewKeyResultButton } from "./new-key-result";
import { UpdateKeyResultDialog } from "./update-key-result-dialog";
import { KeyResultsSkeleton } from "./key-results-skeleton";
import {
  KeyResultDeepLinkTarget,
  resolveTargetKeyResultId,
} from "./key-result-deep-link";

export type KeyResultsViewport = "desktop" | "mobile";

const RenderValue = ({
  value,
  measurementType,
}: {
  value: number;
  measurementType: KeyResult["measurementType"];
}) => {
  const percentFormatter = new Intl.NumberFormat("en-US", {
    style: "percent",
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  });
  const formatter = new Intl.NumberFormat("en-US", {
    style: "decimal",
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  });

  if (measurementType === "percentage") {
    return (
      <Text className="text-[0.95rem]">
        {percentFormatter.format(value / 100)}
      </Text>
    );
  }
  if (measurementType === "boolean") {
    return (
      <Badge
        className="h-[1.6rem] rounded px-1 text-[0.95rem] leading-[1.6rem]"
        color={value ? "success" : "tertiary"}
      >
        {value ? "Complete" : "Incomplete"}
      </Badge>
    );
  }
  return <Text className="text-[0.95rem]">{formatter.format(value)}</Text>;
};

const Okr = ({
  id,
  sequenceId,
  objectiveId,
  name,
  startValue,
  targetValue,
  currentValue,
  measurementType,
  lead,
  contributors,
  startDate,
  endDate,
  createdBy,
  updatedAt,
  isDeepLinkTarget,
  viewport,
}: KeyResult & {
  isDeepLinkTarget: boolean;
  viewport: KeyResultsViewport;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const { isAdminOrOwner } = useIsAdminOrOwner(createdBy);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isUpdateOpen, setIsUpdateOpen] = useState(false);
  const [updateMode, setUpdateMode] = useState<"progress" | "other">(
    "progress",
  );
  const { data: members = [] } = useMembers();
  const { mutate: deleteKeyResult } = useDeleteKeyResultMutation();
  const { mutate: updateKeyResult } = useUpdateKeyResultMutation();

  const leadMember = members.find((member) => member.id === lead);

  const getProgress = () => {
    if (measurementType === "boolean") {
      return currentValue === 1 ? 100 : 0;
    }
    // Calculate progress relative to start value
    const totalChange = targetValue - startValue;
    const actualChange = currentValue - startValue;
    return Math.min(Math.round((actualChange / totalChange) * 100), 100);
  };

  const getOverDueColor = () => {
    const daysLeft = differenceInDays(new Date(), new Date(endDate));
    if (getProgress() < 100) {
      if (new Date() > new Date(endDate)) {
        return "danger";
      }
      if (daysLeft <= 7) {
        return "warning";
      }
    }
    return "muted";
  };

  const getOverDueText = () => {
    const color = getOverDueColor();
    if (color === "danger") {
      return (
        <Text>
          {getTermDisplay("keyResultTerm", {
            variant: "singular",
          })}{" "}
          is overdue
        </Text>
      );
    }
    if (color === "warning") {
      return (
        <Text>
          {getTermDisplay("keyResultTerm", {
            variant: "singular",
          })}{" "}
          is due soon
        </Text>
      );
    }
    return null;
  };

  const handleDelete = () => {
    deleteKeyResult({ keyResultId: id, objectiveId });
  };

  return (
    <KeyResultDeepLinkTarget
      id={id}
      isTarget={isDeepLinkTarget}
      name={name}
      viewport={viewport}
    >
      <Wrapper className="flex flex-col gap-4 py-2.5 md:flex-row md:items-center md:justify-between md:gap-2 md:px-3.5">
        <Flex align="center" gap={3}>
          <Badge className="aspect-square h-9" color="tertiary">
            <OKRIcon strokeWidth={2.8} />
          </Badge>
          <Box>
            <Text className="line-clamp-1" title={name}>
              {name}
            </Text>
            <Tooltip title={getOverDueText()}>
              <Text
                className="flex items-center gap-1 text-[0.95rem]"
                color={getOverDueColor()}
              >
                <CalendarIcon className="h-4 text-current dark:text-current" />
                {format(new Date(startDate), "MMM d, yyyy")} -{" "}
                {format(new Date(endDate), "MMM d, yyyy")}
              </Text>
            </Tooltip>
          </Box>
        </Flex>
        <Flex
          align="center"
          className="divide-border shrink-0 justify-between divide-x"
        >
          <Flex
            align="center"
            className="pr-6 md:px-6"
            direction="column"
            gap={1}
          >
            <Text color="muted">Current</Text>
            <RenderValue
              measurementType={measurementType}
              value={currentValue}
            />
          </Flex>
          {measurementType !== "boolean" && (
            <Flex align="center" className="px-6" direction="column" gap={1}>
              <Text color="muted">Target</Text>
              <RenderValue
                measurementType={measurementType}
                value={targetValue}
              />
            </Flex>
          )}
          <Flex align="center" className="px-5" direction="column" gap={1}>
            <Text color="muted">Progress</Text>
            <Flex align="center" gap={2}>
              <ProgressBar className="w-16" progress={getProgress()} />
              <Text className="text-[0.95rem]">{getProgress()}%</Text>
            </Flex>
          </Flex>

          {isAdminOrOwner ? (
            <Flex align="center" className="h-full py-2 pl-4" gap={2}>
              <Tooltip title="Lead">
                <span>
                  <AssigneesMenu>
                    <AssigneesMenu.Trigger>
                      <Button
                        asIcon
                        className={cn("font-medium", {
                          "text-text-secondary": !lead,
                        })}
                        color="tertiary"
                        leftIcon={
                          <Avatar
                            className={cn({
                              "text-foreground/80": !lead,
                            })}
                            name={leadMember?.username}
                            size="sm"
                            src={leadMember?.avatarUrl}
                          />
                        }
                        rounded="full"
                        size="sm"
                        type="button"
                        variant="naked"
                      >
                        <span className="sr-only">{leadMember?.username}</span>
                      </Button>
                    </AssigneesMenu.Trigger>
                    <AssigneesMenu.Items
                      assigneeId={lead}
                      onAssigneeSelected={(leadUser) => {
                        updateKeyResult({
                          keyResultId: id,
                          objectiveId,
                          data: { lead: leadUser },
                        });
                      }}
                      placeholder="Assign lead..."
                      teamId={teamId}
                    />
                  </AssigneesMenu>
                </span>
              </Tooltip>
              <Menu>
                <Menu.Button>
                  <Button
                    asIcon
                    className="border-border"
                    color="tertiary"
                    leftIcon={<MoreHorizontalIcon />}
                    rounded="full"
                    size="xs"
                  >
                    <span className="sr-only">Edit</span>
                  </Button>
                </Menu.Button>
                <Menu.Items>
                  <Menu.Group>
                    <Menu.Item
                      onSelect={() => {
                        setUpdateMode("other");
                        setIsUpdateOpen(true);
                      }}
                    >
                      <EditIcon />
                      Edit...
                    </Menu.Item>
                    <Menu.Item
                      onSelect={() => {
                        setUpdateMode("progress");
                        setIsUpdateOpen(true);
                      }}
                    >
                      <OKRIcon />
                      Update progress...
                    </Menu.Item>
                    <Menu.Item
                      onSelect={() => {
                        setIsDeleteOpen(true);
                      }}
                    >
                      <DeleteIcon />
                      Delete
                    </Menu.Item>
                  </Menu.Group>
                </Menu.Items>
              </Menu>
            </Flex>
          ) : null}
        </Flex>
        <ConfirmDialog
          confirmText="Yes, Delete"
          description="Are you sure you want to delete this key result? This action cannot be undone."
          isOpen={isDeleteOpen}
          onClose={() => {
            setIsDeleteOpen(false);
          }}
          onConfirm={handleDelete}
          title="Delete Key Result"
        />
        <UpdateKeyResultDialog
          isOpen={isUpdateOpen}
          keyResult={{
            id,
            sequenceId,
            objectiveId,
            name,
            startValue,
            targetValue,
            currentValue,
            measurementType,
            lead,
            contributors,
            startDate,
            endDate,
            createdAt: "",
            updatedAt,
          }}
          onOpenChange={setIsUpdateOpen}
          updateMode={updateMode}
        />
      </Wrapper>
    </KeyResultDeepLinkTarget>
  );
};

export const KeyResults = ({ viewport }: { viewport: KeyResultsViewport }) => {
  const { getTermDisplay } = useTerminology();
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const searchParams = useSearchParams();
  const createKeyResultsMutation = useCreateKeyResultsMutation();
  const { data: keyResults = [], isPending } = useKeyResults(objectiveId);
  const { data: objective } = useObjective(objectiveId);
  const [manualSelectedKeyResults, setManualSelectedKeyResults] = useState<
    Set<string>
  >(new Set());
  const [hasCustomSelection, setHasCustomSelection] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(true);
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const targetKeyResultId = resolveTargetKeyResultId(
    searchParams.get("keyResultId"),
    keyResults.map((keyResult) => keyResult.id),
  );
  const isActiveViewport = viewport === (isDesktop ? "desktop" : "mobile");

  const { object, submit, isLoading } = useObject({
    api: "/api/suggest-key-results",
    schema: keyResultGenerationSchema,
  });

  const suggestedKeyResultNames = useMemo(
    () =>
      (object?.keyResults ?? [])
        .map((keyResult) => keyResult?.name)
        .filter((name): name is string => Boolean(name)),
    [object?.keyResults],
  );

  const selectedKeyResults = useMemo(() => {
    if (hasCustomSelection) {
      return manualSelectedKeyResults;
    }
    return new Set(suggestedKeyResultNames);
  }, [hasCustomSelection, manualSelectedKeyResults, suggestedKeyResultNames]);

  const toggleSelection = (keyResultName: string) => {
    setHasCustomSelection(true);
    setManualSelectedKeyResults((prev) => {
      const baseSelection = hasCustomSelection
        ? prev
        : new Set(suggestedKeyResultNames);
      const newSet = new Set(baseSelection);
      if (newSet.has(keyResultName)) {
        newSet.delete(keyResultName);
      } else {
        newSet.add(keyResultName);
      }
      return newSet;
    });
  };

  const clearSelection = () => {
    setHasCustomSelection(true);
    setManualSelectedKeyResults(new Set());
  };

  const handleAddSelected = async () => {
    const suggestions = object?.keyResults?.filter((kr) =>
      selectedKeyResults.has(kr?.name ?? ""),
    );
    if (!objective || !suggestions?.length) return;

    const parsedSuggestions = suggestions.map((suggestion) =>
      keyResultGenerationSchema.shape.keyResults.element.safeParse(suggestion),
    );
    if (parsedSuggestions.some((result) => !result.success)) {
      toast.error("Suggestions are still being prepared", {
        description: "Wait for Maya to finish, then try creating them again.",
      });
      return;
    }

    try {
      await createKeyResultsMutation.mutateAsync({
        objectiveId,
        keyResults: parsedSuggestions.map((result) =>
          toKeyResultCreateInput(result.data!, objective),
        ),
      });
      clearSelection();
      setShowSuggestions(false);
    } catch {
      // The mutation keeps the selected suggestions visible for retry.
    }
  };

  const handleCancel = () => {
    clearSelection();
    setShowSuggestions(false);
  };

  if (isPending) {
    return (
      <Box className="my-8">
        <Flex align="center" className="mb-3" justify="between">
          <Text
            className="text-lg capitalize antialiased"
            fontWeight="semibold"
          >
            {getTermDisplay("keyResultTerm", {
              variant: "plural",
              capitalize: true,
            })}
          </Text>

          <NewKeyResultButton className="capitalize" size="sm" />
        </Flex>
        <Divider className="my-3" />
        <KeyResultsSkeleton />
      </Box>
    );
  }

  return (
    <Box className="my-8">
      <Flex align="center" className="mb-3" justify="between">
        <Text className="text-lg capitalize" fontWeight="semibold">
          {getTermDisplay("keyResultTerm", {
            variant: "plural",
            capitalize: true,
          })}
        </Text>
        <Flex align="center" gap={2}>
          {isAdminOrOwner ? (
            <Button
              color="tertiary"
              disabled={isLoading}
              leftIcon={<AiIcon className="text-primary dark:text-primary" />}
              onClick={() => {
                setHasCustomSelection(false);
                setManualSelectedKeyResults(new Set());
                setShowSuggestions(true);
                submit({ objective, keyResults });
              }}
              size="sm"
              variant="naked"
            >
              {isLoading ? (
                <Thinking
                  message={isDesktop ? "Maya is thinking" : "Thinking"}
                />
              ) : (
                <>
                  Suggest{" "}
                  {isDesktop
                    ? getTermDisplay("keyResultTerm", {
                        capitalize: true,
                        variant: "plural",
                      })
                    : null}
                </>
              )}
            </Button>
          ) : null}

          {keyResults.length > 0 && (
            <NewKeyResultButton className="capitalize" size="sm" />
          )}
        </Flex>
      </Flex>
      <Divider />

      {object?.keyResults && showSuggestions ? (
        <Box className="my-3">
          {object.keyResults.length > 0 ? (
            <>
              <Box className="border-border/80 shadow-shadow rounded-2xl border shadow-lg dark:shadow-none">
                {object.keyResults.map((keyResult) => {
                  if (!keyResult?.name) return null;
                  return (
                    <RowWrapper
                      className="border-border/80 gap-6 border-b px-2 last-of-type:border-b-0 md:px-4"
                      key={keyResult.name}
                    >
                      <Flex align="center" className="flex-1" gap={2}>
                        <Badge className="aspect-square h-9" color="tertiary">
                          <AiIcon />
                        </Badge>
                        <Text
                          className="line-clamp-1"
                          color={
                            selectedKeyResults.has(keyResult.name)
                              ? undefined
                              : "muted"
                          }
                        >
                          {keyResult.name}
                        </Text>
                      </Flex>
                      <Checkbox
                        checked={selectedKeyResults.has(keyResult.name)}
                        className="shrink-0"
                        onCheckedChange={() => {
                          toggleSelection(keyResult.name!);
                        }}
                      />
                    </RowWrapper>
                  );
                })}
              </Box>
              <Flex className="mt-2" gap={2} justify="end">
                <Button color="tertiary" onClick={handleCancel} variant="naked">
                  Cancel
                </Button>
                <Button
                  disabled={
                    selectedKeyResults.size === 0 ||
                    isLoading ||
                    createKeyResultsMutation.isPending
                  }
                  onClick={() => void handleAddSelected()}
                >
                  {createKeyResultsMutation.isPending ? (
                    "Creating..."
                  ) : (
                    <>
                      Create {selectedKeyResults.size}{" "}
                      {getTermDisplay("keyResultTerm", {
                        capitalize: true,
                        variant:
                          selectedKeyResults.size === 1 ? "singular" : "plural",
                      })}
                    </>
                  )}
                </Button>
              </Flex>
            </>
          ) : (
            <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 flex items-center justify-between gap-2 border p-4">
              <Flex align="center" gap={2}>
                <InfoIcon className="text-warning dark:text-warning" />
                <Text>
                  Could not generate sub{" "}
                  {getTermDisplay("storyTerm", {
                    variant: "plural",
                  })}
                  , make sure your parent {getTermDisplay("storyTerm")} is
                  actionable
                </Text>
              </Flex>
              <Button
                color="warning"
                onClick={() => {
                  setHasCustomSelection(false);
                  setManualSelectedKeyResults(new Set());
                  setShowSuggestions(true);
                  submit({ objective, keyResults });
                }}
              >
                Try again
              </Button>
            </Wrapper>
          )}
        </Box>
      ) : null}

      {keyResults.length > 0 ? (
        <Flex className="mt-3" direction="column" gap={3}>
          {keyResults.map((keyResult) => (
            <Okr
              isDeepLinkTarget={
                isActiveViewport ? targetKeyResultId === keyResult.id : false
              }
              key={keyResult.id}
              {...keyResult}
              objectiveId={objectiveId}
              viewport={viewport}
            />
          ))}
        </Flex>
      ) : (
        <Flex
          align="center"
          className="mt-12"
          direction="column"
          gap={4}
          justify="center"
        >
          <RoadmapEmptyIllustration className="w-52" />
          <Text className="max-w-lg text-center" color="muted">
            You haven&apos;t added any{" "}
            {getTermDisplay("keyResultTerm", {
              variant: "plural",
            })}{" "}
            yet, add{" "}
            {getTermDisplay("keyResultTerm", {
              variant: "plural",
            })}{" "}
            to your{" "}
            {getTermDisplay("objectiveTerm", {
              variant: "plural",
            })}{" "}
            to track your progress
          </Text>
          <NewKeyResultButton className="capitalize" />
        </Flex>
      )}
    </Box>
  );
};
