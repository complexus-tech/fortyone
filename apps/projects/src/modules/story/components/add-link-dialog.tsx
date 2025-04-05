import { Button, Flex, Text, Dialog, Input } from "ui";
import { PlusIcon } from "icons";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { cn } from "lib";
import { useCreateLinkMutation } from "@/lib/hooks/create-link-mutation";
import type { NewLink } from "@/lib/actions/links/create-link";
import type { Link } from "@/types";
import { useUpdateLinkMutation } from "@/lib/hooks/update-link-mutation";
import { useTerminology } from "@/hooks";

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
  const { getTermDisplay } = useTerminology();
  const { mutate: createLink } = useCreateLinkMutation();
  const { mutate: updateLink } = useUpdateLinkMutation();
  const [form, setForm] = useState<NewLink>({
    url: link?.url || "",
    title: link?.title || "",
    storyId,
  });
  const isEditing = Boolean(link);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsOpen(false);
    if (isEditing) {
      updateLink(
        {
          linkId: link!.id,
          payload: {
            title: form.title,
            url: form.url,
          },
          storyId,
        },
        {
          onError: () => {
            setIsOpen(true);
          },
        },
      );
    } else {
      createLink(form, {
        onError: () => {
          setIsOpen(true);
        },
      });
    }
  };

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose={false}>
        <Dialog.Header className="flex items-center justify-between px-6 pb-2">
          <Dialog.Title>
            <Text fontSize="lg" fontWeight="medium">
              {isEditing
                ? "Update link"
                : `Add link to ${getTermDisplay("storyTerm")}`}
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
              value={form.url}
            />
            <Input
              label="Title"
              name="title"
              onChange={handleChange}
              placeholder="Enter title..."
              value={form.title}
            />
            <Flex align="center" className="mt-2" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                }}
                type="button"
                variant="outline"
              >
                Cancel
              </Button>
              <Button
                className={cn({
                  "px-4": isEditing,
                })}
                leftIcon={
                  isEditing ? null : (
                    <PlusIcon className="text-white dark:text-gray-200" />
                  )
                }
                type="submit"
              >
                {isEditing ? "Update" : "Add link"}
              </Button>
            </Flex>
          </form>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
