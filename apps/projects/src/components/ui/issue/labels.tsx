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
  return (
    <Flex align="center" gap={1}>
      <Button
        className="gap-1 pl-1.5"
        color="tertiary"
        leftIcon={<Dot className="size-[0.45rem] text-info" />}
        rounded="xl"
        size="xs"
        variant="outline"
      >
        Feature
      </Button>
      <Button
        className="gap-1 pl-1.5"
        color="tertiary"
        leftIcon={<Dot className="size-[0.45rem] text-danger" />}
        rounded="xl"
        size="xs"
        variant="outline"
      >
        Bug
      </Button>
    </Flex>
  );
};
