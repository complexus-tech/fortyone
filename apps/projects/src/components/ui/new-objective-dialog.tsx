"use client";
import { type Dispatch, type SetStateAction } from "react";
import {
  Button,
  Badge,
  Text,
  Dialog,
  Flex,
  TextEditor,
  DatePicker,
  Menu,
} from "ui";
import { useEditor } from "@tiptap/react";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExt from "@tiptap/extension-text";
import {
  CalendarPlusIcon,
  CalendarIcon,
  PlusIcon,
  CheckIcon,
  ArrowRightIcon,
} from "icons";
import { useMembers } from "@/lib/hooks/members";
import { useLocalStorage } from "@/hooks";
import type { Team } from "@/modules/teams/types";
import { useTeams } from "@/modules/teams/hooks/teams";

export const NewObjectiveDialog = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const [activeTeam, setActiveTeam] = useLocalStorage<Team>(
    "activeTeam",
    teams.at(0)!,
  );
  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Objective name" }),
    ],
    content: "",
    editable: true,
  });
  const descriptionEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({
        placeholder: "Write a description, goals, etc. for this objective...",
      }),
    ],
    content: "",
    editable: true,
  });

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size="lg">
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Menu>
              <Menu.Button>
                <Button
                  className="gap-1 font-semibold tracking-wide"
                  color="tertiary"
                  leftIcon={<span>{activeTeam.icon}</span>}
                  size="xs"
                >
                  {activeTeam.code}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="w-52">
                <Menu.Group>
                  {teams.map((team) => (
                    <Menu.Item
                      active={team.id === activeTeam.id}
                      className="justify-between gap-3"
                      key={team.id}
                      onClick={() => {
                        setActiveTeam(team);
                      }}
                    >
                      <span className="flex items-center gap-1.5">
                        <span className="shrink-0">{team.icon}</span>
                        <span className="block truncate">{team.name}</span>
                      </span>
                      {team.id === activeTeam.id && (
                        <CheckIcon className="h-[1.1rem] w-auto" />
                      )}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
            <ArrowRightIcon className="h-4 w-auto opacity-40" strokeWidth={3} />
            <Text color="muted">New objective</Text>
          </Dialog.Title>
          <Dialog.Close />
        </Dialog.Header>
        <Dialog.Body className="pt-0">
          <TextEditor
            asTitle
            className="text-3xl font-medium"
            editor={titleEditor}
          />
          <TextEditor editor={descriptionEditor} />
          <Flex className="mt-8" gap={1}>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Sep 27
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarPlusIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Due date
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Sep 27
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Sep 27
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer className="flex items-center justify-end gap-2">
          <Button
            className="px-5"
            color="tertiary"
            onClick={() => {
              setIsOpen(false);
            }}
            variant="outline"
          >
            Discard
          </Button>
          <Button leftIcon={<PlusIcon className="h-5 w-auto" />}>
            Create objective
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
