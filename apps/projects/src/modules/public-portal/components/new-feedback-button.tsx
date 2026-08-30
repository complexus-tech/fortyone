"use client";

import { PlusIcon } from "icons";
import { Button, Dialog } from "ui";
import { useNewFeedbackComposer } from "../hooks/use-new-feedback-composer";
import type { PublicPortal, PublicPortalParticipant } from "../types";
import { NewFeedbackComposerContent } from "./new-feedback-composer-content";

export const NewFeedbackButton = ({
  initialOpen = false,
  participant,
  portal,
}: {
  initialOpen?: boolean;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
}) => {
  const composer = useNewFeedbackComposer({
    initialOpen,
    participant,
    portal,
  });

  return (
    <Dialog onOpenChange={composer.handleDialogOpenChange} open={composer.open}>
      <Button
        className="h-12 w-full justify-center text-[1rem]"
        color="invert"
        leftIcon={<PlusIcon className="h-4 text-current" />}
        onClick={composer.openComposer}
        size="lg"
      >
        New Feedback
      </Button>
      <Dialog.Content className="max-w-4xl overflow-visible" hideClose>
        <NewFeedbackComposerContent composer={composer} portal={portal} />
      </Dialog.Content>
    </Dialog>
  );
};
