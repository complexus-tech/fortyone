import { TagsIcon } from "icons";
import type { Icon } from "icons/src/types";
import { Button, Flex } from "ui";

const Dot = (props: Icon) => {
  return (
    <svg
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <circle cx={12} cy={12} fill="currentColor" r={12} />
    </svg>
  );
};

export const Labels = () => {
  const labelss = [
    {
      color: "#06b6d4",
      name: "Feature",
    },
    {
      color: "#f43f5e",
      name: "Bug",
    },
    {
      color: "#f59e0b",
      name: "Improvement",
    },
    {
      color: "#6ee7b7",
      name: "Task",
    },
    {
      color: "#f59e0b",
      name: "Story",
    },
  ];

  const labels: any[] = [];

  if (labels.length === 0)
    return (
      <Button
        className="gap-1 pl-1.5"
        color="tertiary"
        leftIcon={<TagsIcon className="h-4 w-auto" />}
        rounded="xl"
        size="xs"
        variant="outline"
      >
        <span className="sr-only">No labels</span>
      </Button>
    );
  return (
    <Flex align="center" gap={1}>
      {labels.length <= 2 ? (
        labels.map(({ name, color }, idx) => (
          <Button
            key={idx}
            className="gap-1 pl-1.5"
            color="tertiary"
            leftIcon={
              <Dot className="size-[0.45rem]" style={{ color: color }} />
            }
            rounded="xl"
            size="xs"
            variant="outline"
          >
            {name}
          </Button>
        ))
      ) : (
        <Button
          className="gap-1 pl-1.5"
          color="tertiary"
          leftIcon={<Dot className="size-[0.45rem] text-info" />}
          rounded="xl"
          size="xs"
          variant="outline"
        >
          {labels.length} labels
        </Button>
      )}
    </Flex>
  );
};
