"use client";
import { useState, type Dispatch, type SetStateAction } from "react";
import {
  Button,
  Badge,
  Dialog,
  Flex,
  Switch,
  Text,
  ContentEditable,
  TextEditor,
} from "ui";
import { CgArrowsExpandRight } from "react-icons/cg";
import { ChevronRight, Plus } from "lucide-react";

export const NewIssueDialog = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  const [title, setTitle] = useState("");

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size="lg">
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Badge color="tertiary">COMP-1</Badge>
            <ChevronRight className="h-4 w-auto opacity-40" strokeWidth={3} />
            <Text color="muted">New issue</Text>
          </Dialog.Title>
          <Flex gap={4}>
            <Button
              className="px-[0.35rem] dark:hover:bg-dark-100"
              color="tertiary"
              href="/"
              size="xs"
              variant="naked"
            >
              <CgArrowsExpandRight className="h-[1.2rem] w-auto" />
              <span className="sr-only">Expand issue to full screen</span>
            </Button>
            <Dialog.Close />
          </Flex>
        </Dialog.Header>
        <Dialog.Body className="pt-0">
          <ContentEditable
            className="mb-1 py-2 text-2xl"
            placeholder="Issue title"
            setValue={setTitle}
            value={title}
          />
          <TextEditor
            content={`
          <h1>Test heading</h1>
          `}
          />

          <Flex className="mt-4" gap={1}>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer className="flex items-center justify-between gap-2">
          <Text color="muted">
            <label className="flex items-center gap-2" htmlFor="more">
              Create more <Switch id="more" />
            </label>
          </Text>
          <Button leftIcon={<Plus className="h-5 w-auto" />} size="md">
            Create issue
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
