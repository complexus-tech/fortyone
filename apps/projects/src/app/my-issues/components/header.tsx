import { BreadCrumbs, Button, Flex } from "ui";
import { SlidersHorizontal, ListTodo } from "lucide-react";
import { HeaderContainer } from "@/components/shared";
import { NewIssueButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          { icon: <ListTodo className="h-5 w-auto" />, name: "My issues" },
          { name: "Assigned" },
        ]}
      />
      <Flex gap={2}>
        <NewIssueButton />
        <Button
          color="tertiary"
          leftIcon={<SlidersHorizontal className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
