import { Button, Dialog, Select, TextArea, Text } from "ui";
import type { Dispatch, SetStateAction } from "react";

export const InviteMembersDialog = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content className="max-w-2xl">
        <Dialog.Header className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
          <Dialog.Title className="px-6 pt-0.5 text-lg">
            Invite members to your workspace
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="py-6">
          <TextArea
            className="border-[0.5px] dark:bg-transparent"
            label="Email addresses"
            placeholder="Enter email addresses"
            rows={3}
          />
          <Text className="mb-2 mt-6">Role</Text>

          <Select>
            <Select.Trigger className="h-[2.8rem] border-[0.5px] px-4 text-base dark:bg-transparent">
              <Select.Input placeholder="Select role" />
            </Select.Trigger>
            <Select.Content>
              <Select.Option value="admin">Admin</Select.Option>
              <Select.Option value="member">Member</Select.Option>
            </Select.Content>
          </Select>

          <Text className="mb-2 mt-6">Teams</Text>

          <Select>
            <Select.Trigger className="h-[2.8rem] border-[0.5px] px-4 text-base dark:bg-transparent">
              <Select.Input placeholder="Select role" />
            </Select.Trigger>
            <Select.Content>
              <Select.Option value="admin">Admin</Select.Option>
              <Select.Option value="member">Member</Select.Option>
            </Select.Content>
          </Select>
        </Dialog.Body>
        <Dialog.Footer className="justify-end">
          <Button>Send invites</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
