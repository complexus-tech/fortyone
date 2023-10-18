import { TbAdjustmentsHorizontal, TbLayoutDashboard } from "react-icons/tb";
import { BreadCrumbs, Button, Container, Flex } from "ui";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import { NewIssueButton } from "@/components/ui";

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Analytics",
              icon: <TbLayoutDashboard className="h-5 w-auto" />,
            },
          ]}
        />
        <Flex gap={2}>
          <Button
            color="tertiary"
            leftIcon={<TbAdjustmentsHorizontal className="h-5 w-auto" />}
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
