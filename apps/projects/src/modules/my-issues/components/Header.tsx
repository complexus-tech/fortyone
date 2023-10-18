import { TbAdjustmentsHorizontal } from "react-icons/tb";
import { BreadCrumbs, Button, Flex } from "ui";
import { HeaderContainer } from "@/components/shared";
import { NewIssueButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[{ name: "My issues" }, { name: "Assigned" }]}
      />
      <Flex gap={2}>
        <NewIssueButton />
        <Button
          color="tertiary"
          leftIcon={<TbAdjustmentsHorizontal className="h-5 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
