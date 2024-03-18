"use client";
import { BreadCrumbs, Button, Container, Flex, Select } from "ui";
import { AnalyticsIcon, PreferencesIcon } from "icons";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import { NewStoryButton } from "@/components/ui";

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Reports",
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
          <NewStoryButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Container className="pt-4">
          Reports page
          <Select>
            <Select.Trigger className="w-40">
              <Select.Input placeholder="Select" />
            </Select.Trigger>
            <Select.Content>
              <Select.Group>
                <Select.Option value="option1">Option 1</Select.Option>
                <Select.Option value="option2">Option 2</Select.Option>
              </Select.Group>
            </Select.Content>
          </Select>
        </Container>
      </BodyContainer>
    </>
  );
}
