import { Plus } from "lucide-react";
import { Button, Flex, Text } from "ui";

export const AddLinks = () => {
  return (
    <Flex align="center">
      <Text fontWeight="medium">Links</Text>
      <Button
        className="ml-auto"
        color="tertiary"
        leftIcon={<Plus className="h-5 w-auto" strokeWidth={2} />}
        size="sm"
        variant="outline"
      >
        Add
      </Button>
    </Flex>
  );
};
