import type { ReactNode } from "react";
import { Dialog, Flex } from "ui";

export const StoryComposerHeader = ({
  actions,
  templatePicker,
  title,
}: {
  actions: ReactNode;
  templatePicker?: ReactNode;
  title: ReactNode;
}) => (
  <Dialog.Header className="flex shrink-0 items-start justify-between gap-3 px-6 pt-6">
    <Flex align="center" className="min-w-0 flex-wrap gap-2">
      <Dialog.Title className="flex items-center gap-1 text-lg">
        {title}
      </Dialog.Title>
      {templatePicker}
    </Flex>
    {actions}
  </Dialog.Header>
);
