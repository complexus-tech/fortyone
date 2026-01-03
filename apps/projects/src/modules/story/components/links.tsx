/* eslint-disable @next/next/no-img-element -- ok for metadata image */
import { Flex, Button, Tooltip, Box, Text, TimeAgo, Menu } from "ui";
import {
  ArrowDown2Icon,
  ArrowUp2Icon,
  CopyIcon,
  DeleteIcon,
  EditIcon,
  MoreHorizontalIcon,
  NewTabIcon,
  PlusIcon,
} from "icons";
import { useState } from "react";
import { cn } from "lib";
import { toast } from "sonner";
import { RowWrapper } from "@/components/ui";
import type { Link as LinkType } from "@/types";
import { useCopyToClipboard } from "@/hooks/clipboard";
import { useLinkMetadata } from "@/lib/hooks/link-metadata";
import { useDeleteLinkMutation } from "@/lib/hooks/delete-link-mutation";
import { useUserRole } from "@/hooks";
import { AddLinkDialog } from "./add-link-dialog";

const StoryLink = ({ link }: { link: LinkType }) => {
  const [_, copyLink] = useCopyToClipboard();
  const { data: metadata } = useLinkMetadata(link.url);
  const { mutateAsync: deleteLink } = useDeleteLinkMutation();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <RowWrapper className="gap-8 px-1 py-2 md:px-1" key={link.id}>
        <a
          className="flex-1 gap-2"
          href={link.url}
          rel="noopener"
          target="_blank"
        >
          <Flex align="center" gap={2}>
            {metadata?.image ? (
              <img
                alt={metadata.title || link.title || link.url}
                className="size-6 rounded-lg object-cover"
                src={metadata.image}
              />
            ) : (
              <NewTabIcon className="text-info/80 dark:text-info/80 mx-0.5 h-[1.3rem]" />
            )}
            <Text
              className="line-clamp-1 max-w-[24ch] font-medium md:shrink-0"
              title={link.title || metadata?.title}
            >
              {link.title || metadata?.title || link.url}
            </Text>
            {metadata?.description ? (
              <Text
                className="line-clamp-1 opacity-80"
                color="muted"
                title={metadata.description}
              >
                {metadata.description.replace("No description", "")}
              </Text>
            ) : null}
          </Flex>
        </a>
        <Flex align="center" className="shrink-0" gap={3}>
          <Text color="muted">
            <TimeAgo timestamp={link.createdAt} />
          </Text>
          <Menu>
            <Menu.Button>
              <Button
                asIcon
                color="tertiary"
                leftIcon={<MoreHorizontalIcon />}
                rounded="full"
                size="sm"
                variant="naked"
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
                  onSelect={() => {
                    setIsOpen(true);
                  }}
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
        link={link}
        setIsOpen={setIsOpen}
        storyId={link.storyId}
      />
    </>
  );
};

export const Links = ({
  storyId,
  isLinksOpen,
  setIsLinksOpen,
  links,
}: {
  storyId: string;
  isLinksOpen: boolean;
  setIsLinksOpen: (isOpen: boolean) => void;
  links: LinkType[];
}) => {
  const [isAddLinkDialogOpen, setIsAddLinkDialogOpen] = useState(false);
  const { userRole } = useUserRole();

  return (
    <Box className="mt-4">
      {links.length > 0 && (
        <Flex
          align="center"
          className={cn({
            "border-border d border-b-[0.5px] pb-2": !isLinksOpen,
          })}
          justify={links.length > 0 ? "between" : "end"}
        >
          <Button
            color="tertiary"
            onClick={() => {
              setIsLinksOpen(!isLinksOpen);
            }}
            rightIcon={
              isLinksOpen ? (
                <ArrowDown2Icon className="h-4" />
              ) : (
                <ArrowUp2Icon className="h-4" />
              )
            }
            leftIcon={<NewTabIcon className="mr-0.5 h-4" />}
            size="sm"
            variant="naked"
            className="font-semibold"
          >
            External links {links.length > 0 ? `(${links.length})` : ""}
          </Button>

          {userRole !== "guest" && (
            <Tooltip title="Add Link">
              <Button
                color="tertiary"
                leftIcon={<PlusIcon />}
                onClick={() => {
                  setIsAddLinkDialogOpen(true);
                }}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">Add Link</span>
              </Button>
            </Tooltip>
          )}
        </Flex>
      )}

      {isLinksOpen && links.length > 0 ? (
        <Box className="border-border d mt-2 border-t-[0.5px] pb-0">
          {links.map((link) => (
            <StoryLink key={link.id} link={link} />
          ))}
        </Box>
      ) : null}

      <AddLinkDialog
        isOpen={isAddLinkDialogOpen}
        setIsOpen={setIsAddLinkDialogOpen}
        storyId={storyId}
      />
    </Box>
  );
};
