import { Flex, Button, Tooltip, Box, Text, TimeAgo, Menu } from "ui";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  CopyIcon,
  DeleteIcon,
  EditIcon,
  LinkIcon,
  MoreHorizontalIcon,
  PlusIcon,
} from "icons";
import { useState } from "react";
import { useLinks } from "@/lib/hooks/links";
import { AddLinkDialog } from "./add-link-dialog";
import { cn } from "lib";
import { RowWrapper } from "@/components/ui";
import Link from "next/link";
import { Link as LinkType } from "@/types";
import { useCopyToClipboard } from "@/hooks/clipboard";
import { toast } from "sonner";
import { useLinkMetadata } from "@/lib/hooks/link-metadata";
import { useDeleteLinkMutation } from "@/lib/hooks/delete-link-mutation";

const StoryLink = ({ link }: { link: LinkType }) => {
  const [_, copyLink] = useCopyToClipboard();
  const { data: metadata } = useLinkMetadata(link.url);
  const { mutateAsync: deleteLink } = useDeleteLinkMutation();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <RowWrapper key={link.id} className="gap-8 px-1 py-2">
        <Link href={link.url} target="_blank" className="flex- gap-2">
          <Flex align="center" gap={2}>
            {metadata?.image ? (
              <img
                src={metadata?.image}
                alt={metadata?.title || link?.title || link?.url}
                loading="lazy"
                className="size-6 rounded-full object-cover"
              />
            ) : (
              <LinkIcon className="size-6 -rotate-45" />
            )}
            <Text
              title={link?.title || metadata?.title}
              className="max-w-[24ch] shrink-0 truncate font-medium"
            >
              {link?.title || metadata?.title}
            </Text>
            {metadata?.description && (
              <Text
                color="muted"
                className="line-clamp-1 opacity-80"
                title={metadata?.description}
              >
                {metadata?.description?.replace("No description", "")}
              </Text>
            )}
          </Flex>
        </Link>
        <Flex align="center" gap={3} className="shrink-0">
          <Text color="muted">
            <TimeAgo timestamp={link.createdAt} />
          </Text>
          <Menu>
            <Menu.Button>
              <Button
                asIcon
                leftIcon={<MoreHorizontalIcon />}
                color="tertiary"
                variant="naked"
                rounded="full"
                size="sm"
              >
                <span className="sr-only">Delete link</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="min-w-44">
              <Menu.Group>
                <Menu.Item
                  className="tracking-wide"
                  onSelect={() => {
                    copyLink(link.url).then(() => {
                      toast.success("Success", {
                        description: "Link copied to clipboard",
                      });
                    });
                  }}
                >
                  <CopyIcon />
                  Copy link
                </Menu.Item>
              </Menu.Group>
              <Menu.Separator className="my-1.5" />
              <Menu.Group>
                <Menu.Item
                  className="tracking-wide"
                  onSelect={() => setIsOpen(true)}
                >
                  <EditIcon />
                  Edit link
                </Menu.Item>
              </Menu.Group>
              <Menu.Separator className="my-1.5" />
              <Menu.Group>
                <Menu.Item
                  className="tracking-wide"
                  onSelect={() => {
                    deleteLink({
                      linkId: link.id,
                      storyId: link.storyId,
                    }).then(() => {
                      toast.success("Success", {
                        description: "Link deleted",
                      });
                    });
                  }}
                >
                  <DeleteIcon />
                  Delete link
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </RowWrapper>
      <AddLinkDialog
        isOpen={isOpen}
        setIsOpen={setIsOpen}
        storyId={link.storyId}
        link={link}
      />
    </>
  );
};

export const Links = ({ storyId }: { storyId: string }) => {
  const { data: links = [] } = useLinks(storyId);

  const [isOpen, setIsOpen] = useState(true);
  const [isAddLinkDialogOpen, setIsAddLinkDialogOpen] = useState(false);

  return (
    <Box className="mt-3">
      {links.length > 0 && (
        <Flex
          align="center"
          justify={links.length > 0 ? "between" : "end"}
          className={cn({
            "border-b-[0.5px] border-gray-100/60 pb-2 dark:border-dark-200":
              !isOpen,
          })}
        >
          <Button
            color="tertiary"
            variant="naked"
            size="sm"
            onClick={() => {
              setIsOpen(!isOpen);
            }}
            rightIcon={
              isOpen ? (
                <ArrowDownIcon className="h-4 w-auto" />
              ) : (
                <ArrowUpIcon className="h-4 w-auto" />
              )
            }
          >
            Links
          </Button>

          <Tooltip title="Add Link">
            <Button
              color="tertiary"
              leftIcon={<PlusIcon />}
              size="sm"
              variant="naked"
              onClick={() => setIsAddLinkDialogOpen(true)}
            >
              <span className="sr-only">Add Link</span>
            </Button>
          </Tooltip>
        </Flex>
      )}

      {isOpen && links.length > 0 && (
        <Box className="mt-2 border-t-[0.5px] border-gray-100/60 pb-0 dark:border-dark-200">
          {links.map((link) => (
            <StoryLink key={link.id} link={link} />
          ))}
        </Box>
      )}

      <AddLinkDialog
        isOpen={isAddLinkDialogOpen}
        setIsOpen={setIsAddLinkDialogOpen}
        storyId={storyId}
      />
    </Box>
  );
};
