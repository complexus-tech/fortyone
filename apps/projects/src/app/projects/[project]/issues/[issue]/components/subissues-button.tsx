import { Plus } from "lucide-react";
import { Button, Flex } from "ui";

export const SubissuesButton = () => {
  return (
    <Flex justify="end">
      <Button
        color="tertiary"
        leftIcon={<Plus className="h-5 w-auto" />}
        size="sm"
        variant="naked"
      >
        Add sub issue
      </Button>
    </Flex>
  );
};
