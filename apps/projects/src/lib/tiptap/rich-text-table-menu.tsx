"use client";

import {
  useEffect,
  useRef,
  useState,
  type ComponentType,
  type SVGProps,
} from "react";
import { createPortal } from "react-dom";
import { useEditorState, type Editor } from "@tiptap/react";
import { addColumn, addRow, TableMap, type TableRect } from "@tiptap/pm/tables";
import { Button, Menu } from "ui";
import {
  DeleteColumnIcon,
  DeleteIcon,
  DeleteRowIcon,
  InsertColumnLeftIcon,
  InsertColumnRightIcon,
  InsertRowDownIcon,
  InsertRowUpIcon,
  LayoutTable01Icon,
  MoreVerticalIcon,
  PlusIcon,
} from "icons";
import { cn } from "lib";
import { getRichTextOverlayRoot } from "./rich-text-overlay";

type TableIconComponent = ComponentType<SVGProps<SVGSVGElement>>;

type TableAction = {
  danger?: boolean;
  disabled?: boolean;
  icon: TableIconComponent;
  label: string;
  onSelect: () => void;
};

const getActiveTableElement = (editor: Editor) => {
  const domAtPosition = editor.view.domAtPos(editor.state.selection.from).node;
  const element =
    domAtPosition instanceof Element
      ? domAtPosition
      : domAtPosition.parentElement;
  return (
    element?.closest<HTMLElement>("table") ??
    element?.closest<HTMLElement>(".tableWrapper") ??
    null
  );
};

type TableBounds = {
  bottom: number;
  height: number;
  left: number;
  right: number;
  top: number;
  viewportLeft: number;
  width: number;
};

const toTableBounds = (
  rect: DOMRect,
  overlayRoot: HTMLElement,
): TableBounds => {
  const isDocumentRoot = overlayRoot === document.body;
  const rootRect = isDocumentRoot ? null : overlayRoot.getBoundingClientRect();
  const left =
    rect.left -
    (rootRect?.left ?? 0) +
    (isDocumentRoot ? window.scrollX : overlayRoot.scrollLeft);
  const top =
    rect.top -
    (rootRect?.top ?? 0) +
    (isDocumentRoot ? window.scrollY : overlayRoot.scrollTop);

  return {
    bottom: top + rect.height,
    height: rect.height,
    left,
    right: left + rect.width,
    top,
    viewportLeft: rect.left,
    width: rect.width,
  };
};

const tableBoundsAreEqual = (
  current: TableBounds | null,
  next: TableBounds | null,
) =>
  current?.bottom === next?.bottom &&
  current?.height === next?.height &&
  current?.left === next?.left &&
  current?.right === next?.right &&
  current?.top === next?.top &&
  current?.viewportLeft === next?.viewportLeft &&
  current?.width === next?.width;

const getScrollContainer = (element: HTMLElement) => {
  let parent = element.parentElement;
  while (parent) {
    const overflowY = window.getComputedStyle(parent).overflowY;
    if (overflowY === "auto" || overflowY === "scroll") return parent;
    parent = parent.parentElement;
  }
  return null;
};

const getActiveTableRect = (editor: Editor): TableRect | null => {
  const { $from } = editor.state.selection;
  for (let depth = $from.depth; depth > 0; depth -= 1) {
    const table = $from.node(depth);
    if (table.type.spec.tableRole !== "table") continue;
    const map = TableMap.get(table);
    return {
      bottom: map.height,
      left: 0,
      map,
      right: map.width,
      table,
      tableStart: $from.start(depth),
      top: 0,
    };
  }
  return null;
};

const appendTableEdge = (editor: Editor, direction: "column" | "row") => {
  const rect = getActiveTableRect(editor);
  if (!rect) return false;
  const transaction =
    direction === "row"
      ? addRow(editor.state.tr, rect, rect.map.height)
      : addColumn(editor.state.tr, rect, rect.map.width);
  editor.view.dispatch(transaction.scrollIntoView());
  return true;
};

const TableActionIcon = ({ action }: { action: TableAction }) => {
  const Icon = action.icon;
  return (
    <Icon
      className={cn("h-5 w-auto shrink-0", {
        "text-danger dark:!text-danger": action.danger,
      })}
      strokeWidth={2}
    />
  );
};

const TableActionItem = ({
  action,
  onActionSelect,
}: {
  action: TableAction;
  onActionSelect: () => void;
}) => (
  <Menu.Item
    className={cn({ "text-danger dark:!text-danger": action.danger })}
    disabled={action.disabled}
    onSelect={() => {
      onActionSelect();
      action.onSelect();
    }}
  >
    <TableActionIcon action={action} />
    <span>{action.label}</span>
  </Menu.Item>
);

const AddTableEdge = ({
  bounds,
  direction,
  editor,
}: {
  bounds: TableBounds;
  direction: "column" | "row";
  editor: Editor;
}) => (
  <button
    aria-label={direction === "row" ? "Add row below" : "Add column right"}
    className={cn(
      "not-prose bg-surface-muted/70 hover:bg-state-hover focus-visible:ring-ring dark:bg-surface-muted/70 absolute z-40 flex items-center justify-center rounded-xl transition-colors outline-none focus-visible:ring-1",
      direction === "row" ? "h-5" : "w-5",
    )}
    onClick={() => {
      appendTableEdge(editor, direction);
    }}
    onMouseDown={(event) => {
      event.preventDefault();
    }}
    style={
      direction === "row"
        ? {
            left: bounds.left,
            top: bounds.bottom + 4,
            width: bounds.width,
          }
        : {
            height: bounds.height,
            left: bounds.right + 4,
            top: bounds.top,
          }
    }
    type="button"
  >
    <PlusIcon className="h-4" strokeWidth={3} />
  </button>
);

const useActiveTableBounds = (
  editor: Editor,
  overlayRoot: HTMLElement,
  scrollTarget: HTMLElement | null,
) => {
  const [bounds, setBounds] = useState<TableBounds | null>(null);

  useEffect(() => {
    const resolvedScrollTarget =
      scrollTarget ?? getScrollContainer(editor.view.dom);
    const updateBounds = () => {
      const activeTable = getActiveTableElement(editor);
      if (!activeTable) {
        setBounds((current) => (current === null ? current : null));
        return;
      }

      const tableRect = activeTable.getBoundingClientRect();
      const viewportRect = resolvedScrollTarget?.getBoundingClientRect();
      const isVisible =
        !viewportRect ||
        (tableRect.bottom > viewportRect.top &&
          tableRect.top < viewportRect.bottom &&
          tableRect.right > viewportRect.left &&
          tableRect.left < viewportRect.right);
      const nextBounds = isVisible
        ? toTableBounds(tableRect, overlayRoot)
        : null;

      setBounds((current) =>
        tableBoundsAreEqual(current, nextBounds) ? current : nextBounds,
      );
    };

    const animationFrame = window.requestAnimationFrame(updateBounds);

    const resizeObserver =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(updateBounds);
    resizeObserver?.observe(editor.view.dom);
    if (overlayRoot !== editor.view.dom) resizeObserver?.observe(overlayRoot);
    editor.view.dom.addEventListener("scroll", updateBounds, {
      capture: true,
      passive: true,
    });
    resolvedScrollTarget?.addEventListener("scroll", updateBounds, {
      passive: true,
    });
    window.addEventListener("scroll", updateBounds, { passive: true });
    window.addEventListener("resize", updateBounds);
    editor.on("transaction", updateBounds);

    return () => {
      window.cancelAnimationFrame(animationFrame);
      resizeObserver?.disconnect();
      editor.view.dom.removeEventListener("scroll", updateBounds, true);
      resolvedScrollTarget?.removeEventListener("scroll", updateBounds);
      window.removeEventListener("scroll", updateBounds);
      window.removeEventListener("resize", updateBounds);
      editor.off("transaction", updateBounds);
    };
  }, [editor, overlayRoot, scrollTarget]);

  return bounds;
};

const ActiveRichTextTableMenu = ({
  editor,
  scrollTarget,
}: {
  editor: Editor;
  scrollTarget: HTMLElement | null;
}) => {
  const restoreEditorFocusRef = useRef(false);
  const overlayRoot = getRichTextOverlayRoot(editor);
  const tableBounds = useActiveTableBounds(editor, overlayRoot, scrollTarget);

  if (!tableBounds || typeof document === "undefined") return null;
  const rowActions: TableAction[] = [
    {
      icon: InsertRowUpIcon,
      label: "Add row above",
      onSelect: () => {
        editor.chain().focus().addRowBefore().run();
      },
    },
    {
      icon: InsertRowDownIcon,
      label: "Add row below",
      onSelect: () => {
        editor.chain().focus().addRowAfter().run();
      },
    },
    {
      icon: InsertColumnLeftIcon,
      label: "Add column left",
      onSelect: () => {
        editor.chain().focus().addColumnBefore().run();
      },
    },
    {
      icon: InsertColumnRightIcon,
      label: "Add column right",
      onSelect: () => {
        editor.chain().focus().addColumnAfter().run();
      },
    },
  ];
  const cellActions: TableAction[] = [
    {
      disabled: !editor.can().chain().focus().mergeCells().run(),
      icon: LayoutTable01Icon,
      label: "Merge selected cells",
      onSelect: () => {
        editor.chain().focus().mergeCells().run();
      },
    },
    {
      disabled: !editor.can().chain().focus().splitCell().run(),
      icon: LayoutTable01Icon,
      label: "Split cell",
      onSelect: () => {
        editor.chain().focus().splitCell().run();
      },
    },
    {
      icon: LayoutTable01Icon,
      label: "Toggle header row",
      onSelect: () => {
        editor.chain().focus().toggleHeaderRow().run();
      },
    },
  ];
  const deleteSelectionActions: TableAction[] = [
    {
      icon: DeleteRowIcon,
      label: "Delete row",
      onSelect: () => {
        editor.chain().focus().deleteRow().run();
      },
    },
    {
      icon: DeleteColumnIcon,
      label: "Delete column",
      onSelect: () => {
        editor.chain().focus().deleteColumn().run();
      },
    },
  ];
  const deleteTableAction: TableAction = {
    danger: true,
    icon: DeleteIcon,
    label: "Delete table",
    onSelect: () => {
      editor.chain().focus().deleteTable().run();
    },
  };

  const menuOpensToLeft = tableBounds.viewportLeft >= 288;
  const tableActionLeft = tableBounds.left - 24;
  const dialogBoundary =
    overlayRoot.closest<HTMLElement>('[role="dialog"]') ?? undefined;
  const menuPortalContainer = dialogBoundary ?? document.body;
  const handleActionSelect = () => {
    restoreEditorFocusRef.current = true;
  };

  return createPortal(
    <>
      <div
        className="not-prose absolute z-50"
        style={{ left: tableActionLeft, top: tableBounds.top + 2 }}
      >
        <Menu>
          <Menu.Button>
            <Button
              aria-label="Table actions"
              asIcon
              className="h-8 w-6 p-0 hover:bg-transparent focus-visible:bg-transparent active:bg-transparent"
              color="tertiary"
              size="sm"
              type="button"
              variant="naked"
            >
              <MoreVerticalIcon className="h-5" />
            </Button>
          </Menu.Button>
          <Menu.Items
            align="start"
            avoidCollisions
            className="max-h-[var(--radix-dropdown-menu-content-available-height)] min-w-56 overflow-y-auto"
            collisionBoundary={dialogBoundary}
            collisionPadding={12}
            hideWhenDetached
            onCloseAutoFocus={(event) => {
              if (!restoreEditorFocusRef.current) return;
              restoreEditorFocusRef.current = false;
              event.preventDefault();
              editor.commands.focus();
            }}
            portalContainer={menuPortalContainer}
            side={menuOpensToLeft ? "left" : "right"}
            sideOffset={4}
            sticky="always"
          >
            <Menu.Group>
              {rowActions.map((action) => (
                <TableActionItem
                  action={action}
                  key={action.label}
                  onActionSelect={handleActionSelect}
                />
              ))}
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              {cellActions.map((action) => (
                <TableActionItem
                  action={action}
                  key={action.label}
                  onActionSelect={handleActionSelect}
                />
              ))}
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              {deleteSelectionActions.map((action) => (
                <TableActionItem
                  action={action}
                  key={action.label}
                  onActionSelect={handleActionSelect}
                />
              ))}
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <TableActionItem
                action={deleteTableAction}
                onActionSelect={handleActionSelect}
              />
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </div>
      <AddTableEdge bounds={tableBounds} direction="column" editor={editor} />
      <AddTableEdge bounds={tableBounds} direction="row" editor={editor} />
    </>,
    overlayRoot,
  );
};

export const RichTextTableMenu = ({
  editor,
  scrollTarget,
}: {
  editor: Editor | null;
  scrollTarget: HTMLElement | null;
}) => {
  const tableSelection = useEditorState({
    editor,
    selector: ({ editor: currentEditor }) =>
      currentEditor?.isEditable && currentEditor.isActive("table")
        ? currentEditor.state.selection.from
        : null,
  });

  if (!editor || tableSelection === null) return null;

  return (
    <ActiveRichTextTableMenu editor={editor} scrollTarget={scrollTarget} />
  );
};
