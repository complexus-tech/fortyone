import { BreadCrumbs, Button, Container, Flex } from "ui";
import { Settings2 } from "lucide-react";
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
            },
          ]}
        />
        <Flex gap={2}>
          <Button
            color="tertiary"
            leftIcon={<Settings2 className="h-4 w-auto" />}
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
