// "use client";

// import { Command as CommandPrimitive } from "cmdk";

// import { cn } from "lib";
// import {
//   ComponentProps,
//   ComponentPropsWithoutRef,
//   ElementRef,
//   HTMLAttributes,
//   forwardRef,
// } from "react";
// import { cva, VariantProps } from "cva";
// import { SearchIcon } from "icons";

// type CommandProps = ComponentProps<typeof CommandPrimitive>;
// export const Command = ({ className, ...props }: CommandProps) => (
//   <CommandPrimitive
//     className={cn(
//       "flex h-full w-full flex-col overflow-hidden rounded-md",
//       className
//     )}
//     {...props}
//   />
// );
// Command.displayName = CommandPrimitive.displayName;

// const CommandInput = forwardRef<
//   React.ElementRef<typeof CommandPrimitive.Input>,
//   React.ComponentPropsWithoutRef<typeof CommandPrimitive.Input>
// >(({ className, ...props }, ref) => (
//   <div className="flex items-center mb-1.5" cmdk-input-wrapper="">
//     <SearchIcon className="h-[1.15rem] w-auto relative left-3 opacity-60" />
//     <CommandPrimitive.Input
//       ref={ref}
//       className={cn(
//         "bg-transparent py-[0.15rem] pl-4 outline-none w-full",
//         className
//       )}
//       {...props}
//     />
//   </div>
// ));

// CommandInput.displayName = CommandPrimitive.Input.displayName;

// const contentClasses = cva(
//   "bg-white dark:bg-dark-200/70 dark:text-gray-200 bg-opacity-80 dark:bg-opacity-100 backdrop-blur text-gray-300 z-50 border border-gray-100 dark:border-dark-100 w-max shadow-lg shadow-dark/10 dark:shadow-dark/20 mt-1 py-2",
//   {
//     variants: {
//       rounded: {
//         sm: "rounded",
//         md: "rounded-lg",
//         lg: "rounded-xl",
//       },
//     },
//     defaultVariants: {
//       rounded: "lg",
//     },
//   }
// );

// const CommandList = forwardRef<
//   React.ElementRef<typeof CommandPrimitive.List>,
//   React.ComponentPropsWithoutRef<typeof CommandPrimitive.List> &
//     VariantProps<typeof contentClasses>
// >(({ className, rounded, ...props }, ref) => (
//   <CommandPrimitive.List
//     ref={ref}
//     className={cn(contentClasses({ rounded }), className)}
//     {...props}
//   />
// ));

// CommandList.displayName = CommandPrimitive.List.displayName;

// const CommandEmpty = forwardRef<
//   React.ElementRef<typeof CommandPrimitive.Empty>,
//   React.ComponentPropsWithoutRef<typeof CommandPrimitive.Empty>
// >((props, ref) => (
//   <CommandPrimitive.Empty
//     ref={ref}
//     className="py-4 text-center text-sm"
//     {...props}
//   />
// ));

// CommandEmpty.displayName = CommandPrimitive.Empty.displayName;

// const CommandGroup = forwardRef<
//   React.ElementRef<typeof CommandPrimitive.Group>,
//   React.ComponentPropsWithoutRef<typeof CommandPrimitive.Group>
// >(({ className, ...props }, ref) => (
//   <CommandPrimitive.Group
//     ref={ref}
//     className={cn("px-2 pt-2", className)}
//     {...props}
//   />
// ));

// CommandGroup.displayName = CommandPrimitive.Group.displayName;

// const CommandSeparator = forwardRef<
//   ElementRef<typeof CommandPrimitive.Separator>,
//   ComponentPropsWithoutRef<typeof CommandPrimitive.Separator>
// >(({ className, ...props }, ref) => (
//   <CommandPrimitive.Separator
//     ref={ref}
//     className={cn("border-t border-gray-100 dark:border-dark-100", className)}
//     {...props}
//   />
// ));
// CommandSeparator.displayName = CommandPrimitive.Separator.displayName;

// const CommandItem = forwardRef<
//   ElementRef<typeof CommandPrimitive.Item>,
//   ComponentPropsWithoutRef<typeof CommandPrimitive.Item> & { active?: boolean }
// >(({ className, active, ...props }, ref) => (
//   <CommandPrimitive.Item
//     ref={ref}
//     className={cn(
//       "flex gap-2 mb-1 items-center select-none focus:dark:bg-dark-50/80 hover:dark:bg-dark-50 hover:bg-gray-50 focus:bg-gray-50 rounded-lg w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
//       {
//         "bg-gray-50/80 dark:bg-dark-50/60": active,
//       },
//       className
//     )}
//     {...props}
//   />
// ));

// CommandItem.displayName = CommandPrimitive.Item.displayName;

// const CommandShortcut = ({
//   className,
//   ...props
// }: HTMLAttributes<HTMLSpanElement>) => {
//   return (
//     <span
//       className={cn("ml-auto text-xs tracking-widest", className)}
//       {...props}
//     />
//   );
// };

// Command.Input = CommandInput;
// Command.List = CommandList;
// Command.Empty = CommandEmpty;
// Command.Group = CommandGroup;
// Command.Item = CommandItem;
// Command.Shortcut = CommandShortcut;
// Command.Separator = CommandSeparator;
