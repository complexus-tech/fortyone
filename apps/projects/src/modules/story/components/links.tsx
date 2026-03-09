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
import type { ParsedFigmaUrl } from "@/lib/utils/figma";
import { parseFigmaUrl } from "@/lib/utils/figma";
import { useUserRole } from "@/hooks";
import { AddLinkDialog } from "./add-link-dialog";

const LinkActions = ({
  link,
  onEdit,
}: {
  link: LinkType;
  onEdit: () => void;
}) => {
  const [_, copyLink] = useCopyToClipboard();
  const { mutateAsync: deleteLink } = useDeleteLinkMutation();

  return (
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
            <Menu.Item className="tracking-wide" onSelect={onEdit}>
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
  );
};

const ExternalStoryLink = ({ link }: { link: LinkType }) => {
  const { data: metadata } = useLinkMetadata(link.url);
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
        <LinkActions
          link={link}
          onEdit={() => {
            setIsOpen(true);
          }}
        />
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

const FigmaBrandIcon = () => {
  return (
    <svg
      aria-hidden="true"
      className="h-5 w-5 shrink-0"
      focusable="false"
      height="16"
      role="img"
      viewBox="0 0 16 16"
      width="16"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M5.33334 15C5.95218 15 6.54567 14.7541 6.98326 14.3166C7.42085 13.879 7.66668 13.2855 7.66668 12.6666V10.3333H5.33334C4.7145 10.3333 4.12101 10.5791 3.68342 11.0167C3.24583 11.4543 3 12.0478 3 12.6666C3 13.2855 3.24583 13.879 3.68342 14.3166C4.12101 14.7541 4.7145 15 5.33334 15Z"
        fill="#0ACF83"
      />
      <path
        d="M3 8.00004C3 7.3812 3.24583 6.78771 3.68342 6.35012C4.12101 5.91254 4.7145 5.6667 5.33334 5.6667H7.66668V10.3333H5.33334C4.7145 10.3333 4.12101 10.0875 3.68342 9.64996C3.24583 9.21238 3 8.61888 3 8.00004Z"
        fill="#A259FF"
      />
      <path
        d="M3 3.33334C3 2.71481 3.24558 2.1216 3.68277 1.68406C4.11997 1.24653 4.71299 1.00048 5.33152 1H7.66486V5.66668L5.33334 5.6667C4.7145 5.6667 4.12101 5.42085 3.68342 4.98326C3.24583 4.54567 3 3.95218 3 3.33334Z"
        fill="#F24E1E"
      />
      <path
        d="M7.66681 1H10.0001C10.619 1 11.2125 1.24583 11.6501 1.68342C12.0877 2.12101 12.3335 2.7145 12.3335 3.33334C12.3335 3.95218 12.0877 4.54567 11.6501 4.98326C11.2125 5.42085 10.619 5.66668 10.0001 5.66668L7.66668 5.6667L7.66681 1Z"
        fill="#FF7262"
      />
      <path
        d="M12.3335 8.00004C12.3335 8.61888 12.0877 9.21238 11.6501 9.64996C11.2125 10.0875 10.619 10.3334 10.0001 10.3334C9.38131 10.3334 8.78781 10.0875 8.35023 9.64996C7.91264 9.21238 7.66681 8.61888 7.66681 8.00004C7.66681 7.3812 7.91264 6.78771 8.35023 6.35012C8.78781 5.91254 9.38131 5.66668 10.0001 5.66668C10.619 5.66668 11.2125 5.91254 11.6501 6.35012C12.0877 6.78771 12.3335 7.3812 12.3335 8.00004Z"
        fill="#1ABCFE"
      />
    </svg>
  );
};

const GENERIC_FIGMA_TITLES = new Set([
  "figma",
  "figma file link",
  "figma frame link",
  "figma design file",
]);

const isGenericFigmaTitle = (value?: string | null) => {
  if (!value) {
    return false;
  }

  const normalized = value.trim().toLowerCase();

  if (GENERIC_FIGMA_TITLES.has(normalized)) {
    return true;
  }

  return normalized.startsWith("figma -") || normalized.startsWith("figma |");
};

const FigmaStoryLink = ({
  link,
  figma,
}: {
  link: LinkType;
  figma: ParsedFigmaUrl;
}) => {
  const { data: metadata } = useLinkMetadata(figma.canonicalUrl);
  const [isOpen, setIsOpen] = useState(false);
  const metadataTitle = isGenericFigmaTitle(metadata?.title)
    ? undefined
    : metadata?.title;
  const previewTitle =
    !isGenericFigmaTitle(link.title) && link.title
      ? link.title
      : figma.name || metadataTitle || figma.fileKey;

  return (
    <>
      <RowWrapper className="gap-8 px-1 py-2 md:px-1" key={link.id}>
        <a
          className="flex-1 gap-2"
          href={figma.canonicalUrl}
          rel="noopener"
          target="_blank"
        >
          <Flex align="center" gap={2}>
            <FigmaBrandIcon />
            <Text
              className="line-clamp-1 max-w-[24ch] font-medium md:shrink-0"
              title={previewTitle}
            >
              {previewTitle}
            </Text>
            <Text className="line-clamp-1 opacity-80" color="muted">
              Created with Figma
            </Text>
          </Flex>
        </a>
        <LinkActions
          link={link}
          onEdit={() => {
            setIsOpen(true);
          }}
        />
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
  const classifiedLinks = links.map((link) => ({
    link,
    figma: parseFigmaUrl(link.url),
  }));

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
          {classifiedLinks.map(({ link, figma }) =>
            figma ? (
              <FigmaStoryLink figma={figma} key={link.id} link={link} />
            ) : (
              <ExternalStoryLink key={link.id} link={link} />
            ),
          )}
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
