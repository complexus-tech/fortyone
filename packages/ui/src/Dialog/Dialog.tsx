"use client";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import {
  ComponentProps,
  ComponentPropsWithoutRef,
  ElementRef,
  HTMLAttributes,
  forwardRef,
} from "react";
import { Box } from "../Box/Box";

import { cn } from "lib";
import { CloseIcon } from "icons";

const DialogTrigger = DialogPrimitive.Trigger;

const DialogPortal = ({ ...props }: DialogPrimitive.DialogPortalProps) => (
  <DialogPrimitive.Portal {...props} />
);
DialogPortal.displayName = DialogPrimitive.Portal.displayName;

const DialogOverlay = forwardRef<
  ElementRef<typeof DialogPrimitive.Overlay>,
  ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cn(
      "fixed inset-0 z-50 bg-black/20 dark:bg-black/20 flex items-start justify-center",
      className
    )}
    {...props}
  />
));
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName;

const DialogClose = ({ className }: { className?: string }) => (
  <DialogPrimitive.Close
    data-testid="close-modal"
    className={cn(
      "rounded-lg inline-block hover:bg-gray-50 dark:hover:bg-dark-100 p-[2px] transition outline-none dark:text-gray-200",
      className
    )}
  >
    <CloseIcon className="h-6 w-auto" />
    <span className="sr-only">Close</span>
  </DialogPrimitive.Close>
);

const DialogContent = forwardRef<
  ElementRef<typeof DialogPrimitive.Content>,
  ComponentPropsWithoutRef<typeof DialogPrimitive.Content> & {
    hideClose?: boolean;
    size?: "sm" | "md" | "lg" | "xl";
  }
>(({ className, children, hideClose, size = "md", ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay>
      <DialogPrimitive.Content
        ref={ref}
        className={cn(
          "w-full bg-white/90 mt-[10%] shadow-xl border-[0.5px] border-gray-100 shadow-dark/20 dark:shadow-dark dark:border-dark-50 dark:bg-dark-300/90 backdrop-blur rounded-xl overflow-hidden max-w-3xl relative",
          {
            "max-w-md": size === "sm",
            "max-w-xl": size === "md",
            "max-w-5xl": size === "lg",
            "max-w-7xl": size === "xl",
          }
        )}
        {...props}
      >
        {children}
        {!hideClose && <DialogClose className="top-4 right-4 absolute" />}
      </DialogPrimitive.Content>
    </DialogOverlay>
  </DialogPortal>
));
DialogContent.displayName = DialogPrimitive.Content.displayName;

const DialogHeader = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn("py-4", className)} {...props} />
);
DialogHeader.displayName = "DialogHeader";

const DialogFooter = ({
  className,
  variant = "naked",
  justify,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  variant?: "naked" | "bordered";
  justify?: "start" | "end" | "center" | "between";
}) => (
  <div
    className={cn(
      "flex px-6 pb-[0.8rem]",
      {
        "border-t-[0.5px] border-gray-100 dark:border-dark-50/80 pt-[0.8rem]":
          variant !== "bordered",
        "justify-start": justify === "start",
        "justify-end": justify === "end",
        "justify-center": justify === "center",
        "justify-between": justify === "between",
      },
      className
    )}
    {...props}
  />
);
DialogFooter.displayName = "DialogFooter";

const DialogTitle = forwardRef<
  ElementRef<typeof DialogPrimitive.Title>,
  ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title
    ref={ref}
    className={cn(
      "font-medium leading-none tracking-tight dark:text-white",
      className
    )}
    {...props}
  />
));
DialogTitle.displayName = DialogPrimitive.Title.displayName;

const DialogDescription = forwardRef<
  ElementRef<typeof DialogPrimitive.Description>,
  ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description
    ref={ref}
    className={cn("text-[0.95rem] px-6", className)}
    {...props}
  />
));
DialogDescription.displayName = DialogPrimitive.Description.displayName;

type DialogProps = ComponentProps<typeof DialogPrimitive.Root>;

type BodyProps = ComponentProps<typeof Box>;

const Body = ({ className, ...props }: BodyProps) => (
  <Box
    className={cn(
      "px-6 pt-2 pb-4 dark:text-white max-h-[80vh] overflow-y-auto",
      className
    )}
    {...props}
  />
);

export const Dialog = ({ children, ...rest }: DialogProps) => (
  <DialogPrimitive.Root {...rest}>{children}</DialogPrimitive.Root>
);

Dialog.Header = DialogHeader;
Dialog.Trigger = DialogTrigger;
Dialog.Title = DialogTitle;
Dialog.Description = DialogDescription;
Dialog.Content = DialogContent;
Dialog.Body = Body;
Dialog.Footer = DialogFooter;
Dialog.Close = DialogClose;
