"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogoutIcon, SettingsIcon } from "icons";
import { Avatar, Button, Menu, Text } from "ui";
import { toast } from "sonner";
import { revokeFeedbackSessionAction } from "./actions";
import type { PublicPortalGuestParticipant } from "./types";
import { getFeedbackPreferencesPath } from "./utils";

export const PublicPortalGuestMenu = ({
  participant,
  portalSlug,
}: {
  participant: PublicPortalGuestParticipant;
  portalSlug: string;
}) => {
  const router = useRouter();

  return (
    <Menu>
      <Menu.Button>
        <Button
          aria-label="Open feedback participant menu"
          asIcon
          className="size-10 p-0"
          color="tertiary"
          variant="naked"
        >
          <Avatar
            className="!size-9 text-sm font-semibold"
            name={participant.displayName}
            rounded="full"
            size="md"
            src={participant.avatarUrl}
          />
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="w-80 pt-2" sideOffset={8}>
        <Menu.Group className="px-4 pt-2.5 pb-2">
          <Text className="line-clamp-1" fontWeight="semibold">
            {participant.displayName}
          </Text>
          {participant.email ? (
            <Text className="line-clamp-1 text-[0.95rem]" color="muted">
              {participant.email}
            </Text>
          ) : null}
          <Text className="mt-1 text-xs" color="muted">
            Verified for this feedback portal
            {participant.masked ? " · Public name hidden" : ""}
          </Text>
        </Menu.Group>
        <Menu.Separator className="mb-2" />
        <Menu.Group>
          <Menu.Item>
            <Link
              className="flex w-full items-center gap-2"
              href={getFeedbackPreferencesPath(portalSlug)}
            >
              <SettingsIcon className="h-[1.15rem]" />
              Email preferences
            </Link>
          </Menu.Item>
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          <Menu.Item
            className="text-danger"
            onSelect={() => {
              void revokeFeedbackSessionAction(portalSlug).then((response) => {
                if (response.error?.message) {
                  toast.error("Unable to end feedback session", {
                    description: response.error.message,
                  });
                  return;
                }
                router.refresh();
              });
            }}
          >
            <LogoutIcon className="text-danger h-5 w-auto" />
            End feedback session
          </Menu.Item>
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
