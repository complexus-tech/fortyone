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
import { Button } from "ui";
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
    element?.closest<HTMLElement>(".tableWrapper") ??
    element?.closest<HTMLElement>("table") ??
    null
  );
};

type TableBounds = {
  bottom: number;
  height: number;
  left: number;
  right: number;
  top: number;
  width: number;
};

const toTableBounds = (rect: DOMRect): TableBounds => ({
  bottom: rect.bottom,
  height: rect.height,
  left: rect.left,
  right: rect.right,
  top: rect.top,
  width: rect.width,
});

const tableBoundsAreEqual = (
  current: TableBounds | null,
  next: TableBounds | null,
) =>
  current?.bottom === next?.bottom &&
  current?.height === next?.height &&
  current?.left === next?.left &&
  current?.right === next?.right &&
  current?.top === next?.top &&
  current?.width === next?.width;

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

const TableActionButton = ({
  action,
  close,
}: {
  action: TableAction;
  close: () => void;
}) => (
  <button
    className={cn(
      "hover:bg-state-hover focus-visible:bg-state-hover flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left outline-none disabled:pointer-events-none disabled:opacity-45",
      { "text-danger dark:!text-danger": action.danger },
    )}
    disabled={action.disabled}
    onClick={() => {
      action.onSelect();
      close();
    }}
    onMouseDown={(event) => {
      event.preventDefault();
    }}
    role="menuitem"
    type="button"
  >
    <TableActionIcon action={action} />
    <span>{action.label}</span>
  </button>
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
      "bg-surface-muted/70 hover:bg-state-hover focus-visible:ring-ring dark:bg-surface-muted/70 fixed z-40 flex items-center justify-center rounded-xl transition-colors outline-none focus-visible:ring-1",
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
  scrollTarget: HTMLElement | null,
) => {
  const [bounds, setBounds] = useState<TableBounds | null>(null);

  useEffect(() => {
    const updateBounds = () => {
      const activeTable = getActiveTableElement(editor);
      if (!activeTable) {
        setBounds((current) => (current === null ? current : null));
        return;
      }

      const tableRect = activeTable.getBoundingClientRect();
      const viewportRect = scrollTarget?.getBoundingClientRect();
      const isVisible =
        !viewportRect ||
        (tableRect.bottom > viewportRect.top &&
          tableRect.top < viewportRect.bottom &&
          tableRect.right > viewportRect.left &&
          tableRect.left < viewportRect.right);
      const nextBounds = isVisible ? toTableBounds(tableRect) : null;

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
    scrollTarget?.addEventListener("scroll", updateBounds, { passive: true });
    window.addEventListener("resize", updateBounds);
    editor.on("transaction", updateBounds);

    return () => {
      window.cancelAnimationFrame(animationFrame);
      resizeObserver?.disconnect();
      scrollTarget?.removeEventListener("scroll", updateBounds);
      window.removeEventListener("resize", updateBounds);
      editor.off("transaction", updateBounds);
    };
  }, [editor, scrollTarget]);

  return bounds;
};

const ActiveDocumentTableMenu = ({
  editor,
  scrollTarget,
}: {
  editor: Editor;
  scrollTarget: HTMLElement | null;
}) => {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const tableBounds = useActiveTableBounds(editor, scrollTarget);

  useEffect(() => {
    if (!open) return;
    const closeOnOutsidePointer = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("pointerdown", closeOnOutsidePointer);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnOutsidePointer);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  if (!tableBounds || typeof document === "undefined") return null;

  const closeMenu = () => {
    setOpen(false);
  };
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

  const menuOpensToLeft = tableBounds.left >= 288;
  const tableActionLeft = Math.max(0, tableBounds.left - 24);

  return createPortal(
    <>
      <div
        className="fixed z-50"
        ref={rootRef}
        style={{ left: tableActionLeft, top: tableBounds.top + 2 }}
      >
        <Button
          aria-expanded={open}
          aria-haspopup="menu"
          aria-label="Table actions"
          asIcon
          className="h-8 w-6 p-0 hover:bg-transparent focus-visible:bg-transparent active:bg-transparent"
          color="tertiary"
          onClick={() => {
            setOpen((current) => !current);
          }}
          onMouseDown={(event) => {
            event.preventDefault();
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <MoreVerticalIcon className="h-5" />
        </Button>
        {open ? (
          <div
            className={cn(
              "border-border/70 bg-surface-elevated dark:border-border-strong/80 absolute top-0 z-50 min-w-60 rounded-xl border-[0.5px] p-1.5 shadow-xl",
              menuOpensToLeft ? "right-full mr-2" : "left-full ml-2",
            )}
            role="menu"
          >
            <div>
              {rowActions.map((action) => (
                <TableActionButton
                  action={action}
                  close={closeMenu}
                  key={action.label}
                />
              ))}
            </div>
            <div className="border-border-strong/80 my-1.5 border-t" />
            <div>
              {cellActions.map((action) => (
                <TableActionButton
                  action={action}
                  close={closeMenu}
                  key={action.label}
                />
              ))}
            </div>
            <div className="border-border-strong/80 my-1.5 border-t" />
            <div>
              {deleteSelectionActions.map((action) => (
                <TableActionButton
                  action={action}
                  close={closeMenu}
                  key={action.label}
                />
              ))}
            </div>
            <div className="border-border-strong/80 my-1.5 border-t" />
            <div>
              <TableActionButton action={deleteTableAction} close={closeMenu} />
            </div>
          </div>
        ) : null}
      </div>
      <AddTableEdge bounds={tableBounds} direction="column" editor={editor} />
      <AddTableEdge bounds={tableBounds} direction="row" editor={editor} />
    </>,
    document.body,
  );
};

export const DocumentTableMenu = ({
  editor,
  scrollTarget,
}: {
  editor: Editor | null;
  scrollTarget: HTMLElement | null;
}) => {
  const tableSelection = useEditorState({
    editor,
    selector: ({ editor: currentEditor }) =>
      currentEditor?.isActive("table")
        ? currentEditor.state.selection.from
        : null,
  });

  if (!editor || tableSelection === null) return null;

  return (
    <ActiveDocumentTableMenu editor={editor} scrollTarget={scrollTarget} />
  );
};
