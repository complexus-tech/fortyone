import { Badge, Box, Button, Flex, Menu, ProgressBar, Text, Wrapper } from "ui";
import { DeleteIcon, EditIcon, MoreHorizontalIcon, OKRIcon } from "icons";
import type { NewKeyResult } from "@/modules/objectives/types";

type KeyResultsListProps = {
  keyResults: NewKeyResult[];
  onEdit: (index: number) => void;
  onRemove: (index: number) => void;
};

const RenderValue = ({
  value,
  measurementType,
}: {
  value: number;
  measurementType: NewKeyResult["measurementType"];
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
  name,
  startValue,
  targetValue,
  measurementType,
  onEdit,
  onRemove,
  index,
}: NewKeyResult & {
  index: number;
  onEdit: (index: number) => void;
  onRemove: (index: number) => void;
}) => {
  return (
    <Wrapper className="flex items-center justify-between gap-2 rounded-[0.65rem] py-2.5">
      <Flex align="center" gap={3}>
        <Badge
          className="aspect-square h-8 border-opacity-50 dark:border-opacity-50"
          color="tertiary"
        >
          <OKRIcon className="h-6" strokeWidth={2.8} />
        </Badge>
        <Box>
          <Text>{name}</Text>
          <Text className="text-[0.95rem] opacity-80" color="muted">
            Last updated now
          </Text>
        </Box>
      </Flex>
      <Flex
        align="center"
        className="divide-x divide-gray-100 dark:divide-dark-100/80"
      >
        <Flex align="center" className="gap-0.5 px-6" direction="column">
          <Text color="muted">Current</Text>
          <RenderValue measurementType={measurementType} value={startValue} />
        </Flex>
        <Flex align="center" className="gap-0.5 px-6" direction="column">
          <Text color="muted">Target</Text>
          <RenderValue measurementType={measurementType} value={targetValue} />
        </Flex>
        <Flex align="center" className="gap-0.5 px-5" direction="column">
          <Text color="muted">Progress</Text>
          <Flex align="center" gap={2}>
            <ProgressBar className="w-16" progress={75} />
            <Text className="text-[0.95rem]">75%</Text>
          </Flex>
        </Flex>

        <Box className="h-full py-2 pl-4">
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
            <Menu.Items className="w-32">
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    onEdit(index);
                  }}
                >
                  <EditIcon />
                  Edit
                </Menu.Item>
                <Menu.Item
                  onSelect={() => {
                    onRemove(index);
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
    </Wrapper>
  );
};

export const KeyResultsList = ({
  keyResults,
  onEdit,
  onRemove,
}: KeyResultsListProps) => {
  return (
    <Flex direction="column" gap={2}>
      {keyResults.map((kr, index) => (
        <Okr
          key={kr.name}
          {...kr}
          index={index}
          onEdit={onEdit}
          onRemove={onRemove}
        />
      ))}
    </Flex>
  );
};
