import { Dot } from "lucide-react";
import { Button, Flex } from "ui";

export const Labels = () => {
  return (
    <Flex align="center" gap={1}>
      <Button
        color="tertiary"
        leftIcon={<Dot className="h-4 w-auto text-info" strokeWidth={9} />}
        rounded="xl"
        size="xs"
        variant="outline"
      >
        Feature
      </Button>
      <Button
        color="tertiary"
        leftIcon={<Dot className="h-4 w-auto text-danger" strokeWidth={9} />}
        rounded="xl"
        size="xs"
        variant="outline"
      >
        Bug
      </Button>
    </Flex>
  );
};
