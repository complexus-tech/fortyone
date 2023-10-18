"use client";
import { TbLayoutDashboard, TbPlus } from "react-icons/tb";
import {
  Box,
  BreadCrumbs,
  Button,
  Container,
  Flex,
  Input,
  Table,
  Text,
  Tooltip,
} from "ui";
import Link from "next/link";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import { IssueHeader } from "@/components/ui";
import { IssueStatusIcon } from "@/components/ui/issue-status-icon";
import { NewIssueDialog } from "@/components/ui/new-issue-dialog";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/issue/assignees-menu";
import { IssueCheckbox } from "@/components/ui/issue/checkbox";
import { DragHandle } from "@/components/ui/issue/drag-handle";
import { Labels } from "@/components/ui/issue/labels";
import { PrioritiesMenu } from "@/components/ui/issue/priorities-menu";
import { StatusesMenu } from "@/components/ui/issue/statuses-menu";

export default function Page(): JSX.Element {
  const headings = ["#", "Name", "Key", "Lead", "Category", "Status"];
  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "All Projects",
              icon: <TbLayoutDashboard className="h-5 w-auto" />,
            },
          ]}
        />
        <Flex gap={3}>
          <Input placeholder="Search projects" />
          <Button leftIcon={<TbPlus className="h-5 w-auto" />}>
            New project
          </Button>
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Container className="select-none bg-gray-50 py-3 dark:bg-dark-200">
          <Text fontWeight="medium">Workspace projects</Text>
        </Container>
        <Container>
          <Table>
            <Table.Head>
              <Table.Tr>
                {headings.map((heading) => (
                  <Table.Th key={heading}>{heading}</Table.Th>
                ))}
              </Table.Tr>
            </Table.Head>
            <Table.Body>
              <Table.Tr>
                <Table.Td>1</Table.Td>
                <Table.Td>Project name here</Table.Td>
                <Table.Td>PROJ-01</Table.Td>
                <Table.Td>John Doe</Table.Td>
                <Table.Td>Software</Table.Td>
                <Table.Td>
                  <IssueStatusIcon status="Backlog" />
                </Table.Td>
              </Table.Tr>
            </Table.Body>
          </Table>
        </Container>

        <RowWrapper>
          <Flex align="center" className="relative select-none" gap={2}>
            <IssueCheckbox />
            <span>🇿🇼</span>
            <Tooltip title="Issue ID: COM-12">
              <Text className="w-[65px] truncate" color="muted">
                MIGRA-01
              </Text>
            </Tooltip>
            <Link href="/issue">
              <Text className="overflow-hidden text-ellipsis whitespace-nowrap hover:opacity-90">
                Data migration for Fin connect
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
      </BodyContainer>
    </>
  );
}
