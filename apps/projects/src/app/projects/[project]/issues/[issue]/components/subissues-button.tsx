import { PlusIcon } from "icons";
import { Button, Flex } from "ui";

export const SubissuesButton = () => {
  return (
    <Flex justify="end">
      <Button
        color="tertiary"
        leftIcon={<PlusIcon className="h-5 w-auto" />}
        size="sm"
        variant="naked"
      >
        Add sub issue
      </Button>
    </Flex>
  );
};
