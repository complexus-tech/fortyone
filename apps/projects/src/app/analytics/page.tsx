import { BreadCrumbs, Button, Container, Flex } from "ui";
import { AnalyticsIcon, PreferencesIcon } from "icons";
import { BodyContainer, HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Analytics",
              icon: <AnalyticsIcon className="h-5 w-auto" />,
            },
          ]}
        />
        <Flex gap={2}>
          <Button
            color="tertiary"
            leftIcon={<PreferencesIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Container className="pt-4">Analytics page</Container>
      </BodyContainer>
    </>
  );
}
