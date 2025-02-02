import {
  Text,
  Box,
  Flex,
  Wrapper,
  Button,
  ProgressBar,
  Menu,
  Badge,
  TimeAgo,
  Divider,
} from "ui";
import { DeleteIcon, EditIcon, MoreHorizontalIcon, OKRIcon } from "icons";
import { useParams } from "next/navigation";
import { useState } from "react";
import { ConfirmDialog } from "@/components/ui";
import { useKeyResults } from "../../hooks";
import type { KeyResult } from "../../types";
import { useDeleteKeyResultMutation } from "../../hooks/use-delete-key-result-mutation";
import { NewKeyResultButton } from "./new-key-result";
import { UpdateKeyResultDialog } from "./update-key-result-dialog";

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
        className="rounded-[0.35rem] px-1"
        color={value ? "success" : "warning"}
      >
        {value ? "Complete" : "Incomplete"}
      </Badge>
    );
  }
  return <Text className="text-[0.95rem]">{formatter.format(value)}</Text>;
};

const Okr = ({
  id,
  objectiveId,
  name,
  startValue,
  targetValue,
  currentValue,
  measurementType,
  updatedAt,
}: KeyResult) => {
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isUpdateOpen, setIsUpdateOpen] = useState(false);
  const { mutate: deleteKeyResult } = useDeleteKeyResultMutation();

  const getProgress = () => {
    if (measurementType === "boolean") {
      return currentValue === 1 ? 100 : 0;
    }
    if (measurementType === "percentage") {
      return currentValue;
    }
    // Calculate progress relative to start value
    const totalChange = targetValue - startValue;
    const actualChange = currentValue - startValue;
    return Math.round((actualChange / totalChange) * 100);
  };

  const handleDelete = () => {
    deleteKeyResult({ keyResultId: id, objectiveId });
  };

  return (
    <Wrapper className="flex items-center justify-between gap-2 rounded-[0.65rem] py-3">
      <Flex align="center" gap={3}>
        <Badge
          className="aspect-square h-9 border-opacity-50 dark:border-opacity-50"
          color="tertiary"
        >
          <OKRIcon strokeWidth={2.8} />
        </Badge>
        <Box>
          <Text>{name}</Text>
          <Text className="text-[0.95rem] opacity-80" color="muted">
            Last updated <TimeAgo timestamp={updatedAt} />
          </Text>
        </Box>
      </Flex>
      <Flex
        align="center"
        className="divide-x divide-gray-100 dark:divide-dark-100/80"
      >
        <Flex align="center" className="px-6" direction="column" gap={1}>
          <Text color="muted">Current</Text>
          <RenderValue measurementType={measurementType} value={currentValue} />
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

        <Box className="h-full py-2 pl-6">
          <Menu>
            <Menu.Button>
              <Button
                asIcon
                color="tertiary"
                leftIcon={<MoreHorizontalIcon />}
                rounded="full"
                size="sm"
              >
                <span className="sr-only">Edit</span>
              </Button>
            </Menu.Button>
            <Menu.Items>
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    setIsUpdateOpen(true);
                  }}
                >
                  <EditIcon />
                  Update...
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
        </Box>
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
          objectiveId,
          name,
          startValue,
          targetValue,
          currentValue,
          measurementType,
          createdAt: "",
          updatedAt,
        }}
        onOpenChange={setIsUpdateOpen}
      />
    </Wrapper>
  );
};

export const KeyResults = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: keyResults = [] } = useKeyResults(objectiveId);

  return (
    <Box className="my-8">
      <Flex align="center" className="mb-3" justify="between">
        <Text className="text-lg antialiased" fontWeight="semibold">
          Key Results
        </Text>
        {keyResults.length > 0 && <NewKeyResultButton size="sm" />}
      </Flex>
      <Divider />
      {keyResults.length > 0 ? (
        <Flex className="mt-3" direction="column" gap={3}>
          {keyResults.map((keyResult) => (
            <Okr key={keyResult.id} {...keyResult} />
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
          <OKRIcon className="h-12" />
          <Text className="max-w-lg text-center" color="muted">
            You haven&apos;t added any key results yet, add key results to your
            objective to track your progress
          </Text>
          <NewKeyResultButton />
        </Flex>
      )}
    </Box>
  );
};
