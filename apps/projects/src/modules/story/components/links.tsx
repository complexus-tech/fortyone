/* eslint-disable @next/next/no-img-element -- ok for metadata image */
import { Flex, Button, Tooltip, Box, Text, TimeAgo, Menu } from "ui";
import {
  ArrowDown2Icon,
  ArrowUp2Icon,
  CopyIcon,
  DeleteIcon,
  EditIcon,
  LinkIcon,
  MoreHorizontalIcon,
  PlusIcon,
} from "icons";
import { useState } from "react";
import { toast } from "sonner";
import { RowWrapper } from "@/components/ui";
import type { Link as LinkType } from "@/types";
import { useCopyToClipboard } from "@/hooks/clipboard";
import { useLinkMetadata } from "@/lib/hooks/link-metadata";
import { useDeleteLinkMutation } from "@/lib/hooks/delete-link-mutation";
import { useUserRole } from "@/hooks";
import { useStoryFigmaLinks } from "@/lib/hooks/figma";
import {
  getGoogleFileTypeLabel,
  GoogleFileTypeIcon,
  hasNativeGoogleWorkspaceIcon,
  parseGoogleDriveURL,
  useGoogleDriveFiles,
} from "@/modules/google-drive";
import { AddLinkDialog } from "./add-link-dialog";

const StoryLink = ({ canEdit, link }: { canEdit: boolean; link: LinkType }) => {
  const [_, copyLink] = useCopyToClipboard();
  const googleDriveURL = parseGoogleDriveURL(link.url);
  const { data: metadata } = useLinkMetadata(link.url, {
    enabled: !googleDriveURL,
  });
  const { mutateAsync: deleteLink } = useDeleteLinkMutation();
  const [isOpen, setIsOpen] = useState(false);

  if (googleDriveURL) {
    const mimeType =
      googleDriveURL.mimeType ?? "application/vnd.google-apps.file";
    const title =
      link.title || metadata?.title || getGoogleFileTypeLabel(mimeType);

    return (
      <>
        <Flex
          align="center"
          className="border-border bg-surface mt-2 gap-3 rounded-xl border px-4 py-3"
          justify="between"
        >
          <a
            className="min-w-0 flex-1"
            href={link.url}
            rel="noopener noreferrer"
            target="_blank"
          >
            <Flex align="center" className="min-w-0" gap={3}>
              <GoogleFileTypeIcon
                className="size-9 bg-transparent"
                mimeType={mimeType}
              />
              <Box className="min-w-0">
                <Text className="truncate font-medium" title={title}>
                  {title}
                </Text>
                <Text color="muted">
                  {hasNativeGoogleWorkspaceIcon(mimeType)
                    ? "Google Drive link"
                    : `${getGoogleFileTypeLabel(mimeType)} · Google Drive link`}
                </Text>
              </Box>
            </Flex>
          </a>
          <Flex align="center" className="shrink-0" gap={1}>
            <Menu>
              <Menu.Button>
                <Button
                  aria-label={`Actions for ${title}`}
                  asIcon
                  color="tertiary"
                  rounded="full"
                  size="sm"
                  variant="naked"
                >
                  <MoreHorizontalIcon />
                </Button>
              </Menu.Button>
              <Menu.Items align="end" className="min-w-44">
                <Menu.Group>
                  <Menu.Item
                    onSelect={() => {
                      copyLink(link.url).then(() => {
                        toast.success("Link copied");
                      });
                    }}
                  >
                    <CopyIcon /> Copy link
                  </Menu.Item>
                  {canEdit ? (
                    <>
                      <Menu.Item
                        onSelect={() => {
                          setIsOpen(true);
                        }}
                      >
                        <EditIcon /> Edit link
                      </Menu.Item>
                      <Menu.Item
                        className="text-error"
                        onSelect={() => {
                          void deleteLink({
                            linkId: link.id,
                            storyId: link.storyId,
                          });
                        }}
                      >
                        <DeleteIcon /> Delete link
                      </Menu.Item>
                    </>
                  ) : null}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        </Flex>
        <AddLinkDialog
          isOpen={isOpen}
          link={link}
          setIsOpen={setIsOpen}
          storyId={link.storyId}
        />
      </>
    );
  }

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
              <LinkIcon className="mx-0.5 h-[1.3rem]" />
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
  const { data: figmaLinks = [] } = useStoryFigmaLinks(storyId);
  const { data: googleDriveFiles = [] } = useGoogleDriveFiles({
    id: storyId,
    type: "story",
  });
  const mirroredFigmaURLs = new Set(
    figmaLinks.map(({ artifact }) => artifact.canonicalUrl),
  );
  const attachedGoogleFileIds = new Set(
    googleDriveFiles.flatMap((file) => {
      const parsed = parseGoogleDriveURL(file.webViewLink);
      return parsed ? [parsed.fileId] : [];
    }),
  );
  const visibleLinks = links.filter((link) => {
    if (mirroredFigmaURLs.has(link.url)) return false;
    const googleDriveURL = parseGoogleDriveURL(link.url);
    return !(
      googleDriveURL && attachedGoogleFileIds.has(googleDriveURL.fileId)
    );
  });

  return (
    <Box className="mt-4">
      {visibleLinks.length > 0 && (
        <Flex
          align="center"
          className="border-border border-b-[0.5px] pb-2"
          justify={visibleLinks.length > 0 ? "between" : "end"}
        >
          <Button
            className="font-semibold"
            color="tertiary"
            leftIcon={<LinkIcon className="mr-0.5 h-5" />}
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
            size="sm"
            variant="naked"
          >
            External links
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

      {isLinksOpen && visibleLinks.length > 0 ? (
        <Box>
          {visibleLinks.map((link) => (
            <StoryLink
              canEdit={userRole !== "guest"}
              key={link.id}
              link={link}
            />
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
