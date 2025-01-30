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
} from "ui";
import { DeleteIcon, EditIcon, MoreHorizontalIcon, OKRIcon } from "icons";
import { useParams } from "next/navigation";
import { useKeyResults } from "../../hooks";
import type { KeyResult } from "../../types";

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
        className="px-1"
        color={value ? "success" : "warning"}
        rounded="full"
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
  updatedAt,
}: KeyResult) => {
  return (
    <Wrapper className="flex items-center justify-between gap-2 rounded-[0.65rem]">
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
          <RenderValue measurementType={measurementType} value={startValue} />
        </Flex>
        <Flex align="center" className="px-6" direction="column" gap={1}>
          <Text color="muted">Target</Text>
          <RenderValue measurementType={measurementType} value={targetValue} />
        </Flex>
        <Flex align="center" className="px-5" direction="column" gap={1}>
          <Text color="muted">Progress</Text>
          <Flex align="center" gap={2}>
            <ProgressBar className="w-16" progress={75} />
            <Text className="text-[0.95rem]">75%</Text>
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
            <Menu.Items className="w-32">
              <Menu.Group>
                <Menu.Item>
                  <EditIcon />
                  Edit
                </Menu.Item>
                <Menu.Item>
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

export const KeyResults = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: keyResults = [] } = useKeyResults(objectiveId);

  return (
    <Box className="my-8">
      <Flex align="center" className="mb-3" justify="between">
        <Text className="text-lg antialiased" fontWeight="semibold">
          Key Results ({keyResults.length})
        </Text>
        <Button color="tertiary" size="sm">
          Add Key Result
        </Button>
      </Flex>
      {keyResults.map((keyResult) => (
        <Okr key={keyResult.id} {...keyResult} />
      ))}
    </Box>
  );
};
