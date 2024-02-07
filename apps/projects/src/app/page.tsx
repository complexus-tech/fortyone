import { BreadCrumbs, Button, Container, Flex } from "ui";
import { Columns3, SlidersHorizontal } from "lucide-react";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import { NewIssueButton } from "@/components/ui";

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Dashboard",
              icon: <Columns3 className="h-5 w-auto" />,
            },
          ]}
        />
        <Flex gap={2}>
          <Button
            color="tertiary"
            leftIcon={<SlidersHorizontal className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Container className="pt-4">Dashboard page</Container>
      </BodyContainer>
    </>
  );
}
