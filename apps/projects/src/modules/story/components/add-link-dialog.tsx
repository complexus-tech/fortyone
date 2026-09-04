import { Button, Flex, Text, Dialog, Input } from "ui";
import { PlusIcon } from "icons";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { cn } from "lib";
import { toast } from "sonner";
import { useCreateLinkMutation } from "@/lib/hooks/create-link-mutation";
import type { NewLink } from "@/lib/actions/links/create-link";
import type { Link } from "@/types";
import { useUpdateLinkMutation } from "@/lib/hooks/update-link-mutation";
import { useTerminology } from "@/hooks";
import { useLinkFigmaStory } from "@/lib/hooks/figma";
import { isFigmaURL } from "@/modules/settings/workspace/integrations/figma/url";
import { useDeleteLinkMutation } from "@/lib/hooks/delete-link-mutation";
import {
  GoogleDrivePickerDialog,
  parseGoogleDriveURL,
  useAttachGoogleDriveFiles,
} from "@/modules/google-drive";
import type { GoogleDriveURL } from "@/modules/google-drive";

type PendingGoogleDriveLink = {
  fallbackLinkId: string;
  parsedURL: GoogleDriveURL;
};

const getErrorCode = (error: unknown) => {
  if (!error || typeof error !== "object" || !("code" in error)) return null;
  const code = (error as { code?: unknown }).code;
  return typeof code === "string" ? code : null;
};

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
  const deleteLink = useDeleteLinkMutation();
  const linkFigmaStory = useLinkFigmaStory();
  const attachGoogleDriveFiles = useAttachGoogleDriveFiles(
    { id: storyId, type: "story" },
    { notifyOnError: false, notifyOnSuccess: false },
  );
  const [pendingGoogleDriveLink, setPendingGoogleDriveLink] =
    useState<PendingGoogleDriveLink | null>(null);
  const [form, setForm] = useState<NewLink>({
    url: link?.url || "",
    title: link?.title || "",
    storyId,
  });
  const isEditing = Boolean(link);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const createGoogleDrivePreview = async (
    fallbackLinkId: string,
    parsedURL: GoogleDriveURL,
  ) => {
    try {
      await attachGoogleDriveFiles.mutateAsync([
        {
          id: parsedURL.fileId,
          ...(parsedURL.mimeType ? { mimeType: parsedURL.mimeType } : {}),
          ...(parsedURL.resourceKey
            ? { resourceKey: parsedURL.resourceKey }
            : {}),
        },
      ]);
      deleteLink.mutate({ linkId: fallbackLinkId, storyId });
    } catch (error) {
      if (getErrorCode(error) === "permission_denied") {
        setPendingGoogleDriveLink({ fallbackLinkId, parsedURL });
        return;
      }

      toast.error("Couldn’t create Google Drive preview", {
        description: "The Google link was saved. Try again in a moment.",
      });
    }
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const googleDriveURL = parseGoogleDriveURL(form.url);
    if (googleDriveURL) {
      setIsOpen(false);
      if (link) {
        updateLink(
          {
            linkId: link.id,
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
            onSuccess: () => {
              void createGoogleDrivePreview(link.id, googleDriveURL);
            },
          },
        );
      } else {
        createLink(form, {
          onError: () => {
            setIsOpen(true);
          },
          onSuccess: (response) => {
            if (!response.data) {
              setIsOpen(true);
              return;
            }
            void createGoogleDrivePreview(response.data.id, googleDriveURL);
          },
        });
      }
      return;
    }

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
      if (isFigmaURL(form.url)) {
        linkFigmaStory.mutate(
          { storyId, title: form.title, url: form.url },
          {
            onError: () => {
              setIsOpen(true);
            },
            onSuccess: (result) => {
              if (result.kind === "generic") {
                toast.success("Figma link saved", {
                  description: "Saved as a normal link without a preview.",
                });
              }
            },
          },
        );
        return;
      }
      createLink(form, {
        onError: () => {
          setIsOpen(true);
        },
      });
    }
  };

  return (
    <>
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
                    isEditing ? null : <PlusIcon className="text-current" />
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
      {pendingGoogleDriveLink ? (
        <GoogleDrivePickerDialog
          fileIds={[pendingGoogleDriveLink.parsedURL.fileId]}
          onAttached={() => {
            deleteLink.mutate({
              linkId: pendingGoogleDriveLink.fallbackLinkId,
              storyId,
            });
            setPendingGoogleDriveLink(null);
          }}
          onClose={() => {
            setPendingGoogleDriveLink(null);
          }}
          target={{ id: storyId, type: "story" }}
        />
      ) : null}
    </>
  );
};
