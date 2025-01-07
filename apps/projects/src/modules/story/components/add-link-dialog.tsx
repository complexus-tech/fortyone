import { Button, Flex, Text, Dialog, Input } from "ui";
import { PlusIcon } from "icons";
import { useCreateLinkMutation } from "@/lib/hooks/create-link-mutation";
import { ChangeEvent, FormEvent, useState } from "react";
import { NewLink } from "@/lib/actions/links/create-link";
import { Link } from "@/types";
import { cn } from "lib";

export const AddLinkDialog = ({
  isOpen,
  setIsOpen,
  storyId,
  link,
}: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
  storyId: string;
  link?: Link;
}) => {
  const { mutateAsync: createLink } = useCreateLinkMutation();
  const [form, setForm] = useState<NewLink>({
    url: link?.url || "",
    title: link?.title || "",
    storyId,
  });
  const isEditing = !!link;

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    await createLink(form).then(() => {
      setIsOpen(false);
    });
  };

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose={false}>
        <Dialog.Header className="flex items-center justify-between px-6 pb-2">
          <Dialog.Title>
            <Text fontSize="lg" fontWeight="medium">
              {isEditing ? "Edit link" : "Add link to story"}
            </Text>
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="pb-5">
          <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
            <Input
              label="URL"
              name="url"
              onChange={handleChange}
              placeholder="https://..."
              required
              type="url"
            />
            <Input
              label="Title"
              name="title"
              onChange={handleChange}
              placeholder="Enter title..."
            />
            <Flex align="center" className="mt-2" gap={2} justify="end">
              <Button
                color="tertiary"
                type="button"
                onClick={() => {
                  setIsOpen(false);
                }}
                variant="outline"
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  isEditing ? null : (
                    <PlusIcon className="text-white dark:text-gray-200" />
                  )
                }
                className={cn({
                  "px-4": isEditing,
                })}
                type="submit"
              >
                {isEditing ? "Save" : "Add link"}
              </Button>
            </Flex>
          </form>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
