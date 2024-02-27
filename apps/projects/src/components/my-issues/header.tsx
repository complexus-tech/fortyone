import { BreadCrumbs, Button, Flex } from "ui";
import { IssueIcon, PreferencesIcon } from "icons";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton, SideDetailsSwitch } from "@/components/ui";

export const Header = ({
  isExpanded,
  setIsExpanded,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
}) => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "My issues",
            icon: <IssueIcon className="h-5 w-auto" strokeWidth={2} />,
          },
          { name: "Assigned" },
        ]}
      />
      <Flex align="center" gap={2}>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <NewIssueButton />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
