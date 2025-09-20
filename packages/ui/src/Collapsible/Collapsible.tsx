"use client";

import * as CollapsiblePrimitive from "@radix-ui/react-collapsible";

export const Collapsible = ({
  ...props
}: React.ComponentProps<typeof CollapsiblePrimitive.Root>) => {
  return <CollapsiblePrimitive.Root data-slot="collapsible" {...props} />;
};

const Trigger = ({
  ...props
}: React.ComponentProps<typeof CollapsiblePrimitive.CollapsibleTrigger>) => {
  return (
    <CollapsiblePrimitive.CollapsibleTrigger
      data-slot="collapsible-trigger"
      {...props}
    />
  );
};

const Content = ({
  ...props
}: React.ComponentProps<typeof CollapsiblePrimitive.CollapsibleContent>) => {
  return (
    <CollapsiblePrimitive.CollapsibleContent
      data-slot="collapsible-content"
      {...props}
    />
  );
};

Collapsible.Trigger = Trigger;
Collapsible.Content = Content;
