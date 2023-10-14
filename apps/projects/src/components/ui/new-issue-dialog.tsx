"use client";
import { TbPlus } from "react-icons/tb";
import { Button, Badge, Dialog, Flex, Switch, Text } from "ui";
import { CgArrowsExpandRight } from "react-icons/cg";
import { IoIosArrowForward } from "react-icons/io";
import type { Dispatch, SetStateAction } from "react";

export const NewIssueDialog = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size="lg">
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Badge color="tertiary">COMP-1</Badge>
            <IoIosArrowForward className="h-3 w-auto opacity-40" />
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
        <Dialog.Body>
          <textarea
            autoComplete="off"
            className="w-full resize-none bg-transparent py-2 text-2xl outline-none"
            placeholder="Issue title"
            spellCheck={false}
          />
          <textarea
            className="mb-4 min-h-[5rem] w-full resize-none bg-transparent text-lg text-gray-200/80 outline-none"
            placeholder="Issue description"
          />
          <Flex gap={1}>
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
              Create more <Switch />
            </label>
          </Text>
          <Button leftIcon={<TbPlus className="h-5 w-auto" />} size="md">
            Create issue
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
