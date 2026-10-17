import type { RefObject } from "react";
import Image from "next/image";
import { ReplyIcon } from "icons";
import { Button, Dialog } from "ui";
import {
  NativeLinkNavigationProvider,
  useNativeLinkNavigation,
} from "@/components/ui/link-preview/native-navigation";
import type { NativeLinkNavigation } from "@/components/ui/link-preview/native-navigation";
import { ChatLinkPreview } from "./link-preview";
import type { ChatResource } from "./types";
import {
  ResourceIcon,
  resourceMetadata,
  resourceTitle,
} from "./resource-presentation";

export const ResourcePreview = ({
  resource,
  url,
  conversationName,
  triggerRef,
  returningToMessage,
  onClose,
  onViewMessage,
}: {
  resource: ChatResource | null;
  url: string | null;
  conversationName: string;
  triggerRef: RefObject<HTMLButtonElement | null>;
  returningToMessage: RefObject<boolean>;
  onClose: () => void;
  onViewMessage: (messageId: string) => void;
}) => {
  const navigate = useNativeLinkNavigation();
  const openNative: NativeLinkNavigation = (target, trigger) => {
    const handled = navigate?.(target, trigger) ?? false;
    if (handled) {
      returningToMessage.current = true;
      onClose();
    }
    return handled;
  };
  return (
    <NativeLinkNavigationProvider onOpen={openNative}>
      <Dialog
        onOpenChange={(open) => {
          if (!open) onClose();
        }}
        open={Boolean(resource)}
      >
        <Dialog.Content
          className="flex max-h-[85dvh] flex-col"
          onCloseAutoFocus={(event) => {
            event.preventDefault();
            if (!returningToMessage.current) triggerRef.current?.focus();
          }}
        >
          <Dialog.Header className="px-6 pr-14 pb-2">
            <Dialog.Title className="text-lg">
              {resource ? resourceTitle(resource) : "Shared item"}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description className="pb-4">
            Shared in {conversationName}
          </Dialog.Description>
          <Dialog.Body className="min-h-0 flex-1 overflow-y-auto pb-6">
            {resource ? (
              <>
                {resource.kind === "link" ? (
                  <ChatLinkPreview
                    key={resource.url}
                    title={resource.title}
                    url={resource.url}
                  />
                ) : (
                  <div className="border-border/60 bg-surface-muted mb-4 flex min-h-40 items-center justify-center rounded-lg border p-6">
                    {resource.mediaType.toLowerCase().startsWith("image/") &&
                    url ? (
                      <Image
                        alt={resource.fileName}
                        className="max-h-[45dvh] w-auto object-contain"
                        height={768}
                        src={url}
                        unoptimized
                        width={1024}
                      />
                    ) : (
                      <ResourceIcon resource={resource} />
                    )}
                  </div>
                )}
                <p className="text-text-muted mt-3">
                  {resourceMetadata(resource)}
                </p>
                {resource.description ? (
                  <p className="mt-3 text-base">{resource.description}</p>
                ) : null}
                <p className="text-text-muted mt-5">
                  Shared by {resource.author.fullName} ·{" "}
                  {new Date(resource.createdAt).toLocaleString()}
                </p>
              </>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 justify-end gap-2 pt-4">
            {url ? (
              <a
                className="text-primary rounded-md px-3 py-2 font-medium underline-offset-4 hover:underline"
                href={url}
                rel="noreferrer noopener"
                target="_blank"
              >
                {resource?.kind === "file" ? "Open file" : "Open link"}
              </a>
            ) : null}
            <Button
              leftIcon={<ReplyIcon className="size-4" />}
              onClick={() => {
                if (!resource) return;
                returningToMessage.current = true;
                onClose();
                onViewMessage(resource.messageId);
              }}
              size="sm"
            >
              View in conversation
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </NativeLinkNavigationProvider>
  );
};
