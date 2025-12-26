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

// Custom styles for dialog animations
const dialogAnimationStyles = `
  @keyframes dialog-overlay-show {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  
  @keyframes dialog-overlay-hide {
    from { opacity: 1; }
    to { opacity: 0; }
  }
  
  @keyframes dialog-content-show {
    from { 
      opacity: 0; 
      transform: translateY(-16px) scale(0.95); 
    }
    to { 
      opacity: 1; 
      transform: translateY(0px) scale(1); 
    }
  }
  
  @keyframes dialog-content-hide {
    from { 
      opacity: 1; 
      transform: translateY(0px) scale(1); 
    }
    to { 
      opacity: 0; 
      transform: translateY(-16px) scale(0.95); 
    }
  }
  
  .dialog-overlay-animate[data-state="open"] {
    animation: dialog-overlay-show 300ms ease-out;
  }
  
  .dialog-overlay-animate[data-state="closed"] {
    animation: dialog-overlay-hide 300ms ease-in;
  }
  
  .dialog-content-animate[data-state="open"] {
    animation: dialog-content-show 300ms ease-out;
  }
  
  .dialog-content-animate[data-state="closed"] {
    animation: dialog-content-hide 300ms ease-in;
  }
`;

// Inject styles if they don't exist
if (
  typeof document !== "undefined" &&
  !document.getElementById("dialog-animations")
) {
  const style = document.createElement("style");
  style.id = "dialog-animations";
  style.textContent = dialogAnimationStyles;
  document.head.appendChild(style);
}

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
      "fixed inset-0 z-50 flex items-start justify-center bg-black/10 dark:bg-black/40 ",
      "dialog-overlay-animate",
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
      "inline-block rounded-[0.6rem] p-1 outline-none transition hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-dark-100",
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
    overlayClassName?: string;
  }
>(
  (
    { className, children, hideClose, overlayClassName, size = "md", ...props },
    ref
  ) => (
    <DialogPortal>
      <DialogOverlay className={overlayClassName}>
        <DialogPrimitive.Content
          ref={ref}
          className={cn(
            "relative mt-[15%] md:mt-[10%] w-full mx-3.5 max-w-3xl backdrop-blur-lg overflow-hidden rounded-3xl border-[0.5px] border-gray-200 bg-white/80 dark:border-dark-50 dark:bg-dark-300/90",
            "dialog-content-animate outline-transparent",
            {
              "max-w-md": size === "sm",
              "max-w-xl": size === "md",
              "max-w-5xl": size === "lg",
              "max-w-7xl": size === "xl",
            },
            className
          )}
          {...props}
        >
          {children}
          {!hideClose && <DialogClose className="absolute right-4 top-4" />}
        </DialogPrimitive.Content>
      </DialogOverlay>
    </DialogPortal>
  )
);
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
        "border-t-[0.5px] border-gray-100 pt-[0.8rem] dark:border-dark-50":
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
    className={cn(
      "px-6 text-[0.95rem] text-gray dark:text-gray-300",
      className
    )}
    {...props}
  />
));
DialogDescription.displayName = DialogPrimitive.Description.displayName;

type DialogProps = ComponentProps<typeof DialogPrimitive.Root>;

type BodyProps = ComponentProps<typeof Box>;

const Body = ({ className, ...props }: BodyProps) => (
  <Box
    className={cn(
      "max-h-[80dvh] overflow-y-auto px-6 pb-4 pt-2 dark:text-white",
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
