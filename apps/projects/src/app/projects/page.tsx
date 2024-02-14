"use client";

import { BreadCrumbs, Button, Container, Flex, Text, Tooltip } from "ui";
import Link from "next/link";
import { Plus, Settings2 } from "lucide-react";
import { BodyContainer, HeaderContainer } from "@/components/layout";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/issue/assignees-menu";
import { IssueCheckbox } from "@/components/ui/issue/checkbox";
import { Labels } from "@/components/ui/issue/labels";

type Project = {
  id: number;
  code: string;
  lead: string;
  name: string;
  description: string;
  date: string;
};

export default function Page(): JSX.Element {
  const projects: Project[] = [
    {
      id: 1,
      code: "COM-12",
      lead: "John Doe",
      name: "🇿🇼 Data migration for Fin connect",
      description: "The quick brown fox jumps over the lazy dog.",
      date: "Sep 27",
    },
    {
      id: 2,
      code: "COM-12",
      lead: "John Doe",
      name: "🇿🇼 Data migration for Fin connect",
      description: "The quick brown fox jumps over the lazy dog.",
      date: "Sep 27",
    },
  ];

  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "All Projects",
            },
          ]}
        />
        <Flex gap={3}>
          <Button
            color="tertiary"
            leftIcon={<Settings2 className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <Button leftIcon={<Plus className="h-5 w-auto" />} size="sm">
            New project
          </Button>
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Container className="select-none bg-gray-50 py-3 dark:bg-dark-200">
          <Text fontWeight="medium">Workspace projects</Text>
        </Container>

        {projects.map(({ id, name, code }) => (
          <RowWrapper key={id}>
            <Flex align="center" className="relative select-none" gap={2}>
              <IssueCheckbox />
              <Tooltip title="Project ID: COM-12">
                <Text
                  align="left"
                  className="w-[60px]"
                  color="muted"
                  textOverflow="truncate"
                >
                  {code}
                </Text>
              </Tooltip>
              <Link href="/issue">
                <Text className="overflow-hidden text-ellipsis whitespace-nowrap hover:opacity-90">
                  {name}
                </Text>
              </Link>
            </Flex>
            <Flex align="center" gap={3}>
              <Labels />
              <Tooltip title="Created on Sep 27, 2021">
                <Text color="muted">Sep 27</Text>
              </Tooltip>
              <AssigneesMenu isSearchEnabled />
            </Flex>
          </RowWrapper>
        ))}
      </BodyContainer>
    </>
  );
}
