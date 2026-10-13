import { useState } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { Checkbox } from "ui";
import { DragIcon } from "icons";
import { cn } from "lib";
import type { StoryChecklistItem } from "../types/checklist";
import { ChecklistItemActions } from "./checklist-item-actions";
import { ChecklistItemContent } from "./checklist-item-content";
import { ChecklistItemForm } from "./checklist-item-form";

export const ChecklistItem = ({
  item,
  canEdit,
  isPending,
  storyTerm,
  onUpdate,
  onDelete,
  onConvert,
}: {
  item: StoryChecklistItem;
  canEdit: boolean;
  isPending: boolean;
  storyTerm: string;
  onUpdate: (
    updates: { title?: string; completed?: boolean },
    onSuccess?: () => void,
  ) => void;
  onDelete: () => void;
  onConvert: () => void;
}) => {
  const [isEditing, setIsEditing] = useState(false);
  const {
    attributes,
    listeners,
    setActivatorNodeRef,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: item.id,
    disabled: !canEdit || isPending || isEditing,
  });
  const convertedStory = item.convertedStory;
  const isConverted = Boolean(item.convertedAt || convertedStory);

  return (
    <li
      className={cn("group flex min-h-11 items-start gap-2 rounded-lg py-1", {
        "bg-surface-elevated relative z-10 shadow-lg": isDragging,
      })}
      ref={setNodeRef}
      style={{
        transform: transform
          ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
          : undefined,
        transition,
      }}
    >
      {canEdit ? (
        <button
          {...attributes}
          {...listeners}
          aria-label={`Reorder ${item.title}`}
          className="text-text-muted hover:bg-state-hover focus-visible:ring-ring mt-1 flex size-7 shrink-0 cursor-grab touch-none items-center justify-center rounded focus-visible:ring-2 focus-visible:outline-none active:cursor-grabbing disabled:cursor-default disabled:opacity-40"
          disabled={isPending || isEditing}
          ref={setActivatorNodeRef}
          type="button"
        >
          <DragIcon className="text-text-muted h-5 w-auto" strokeWidth={3.5} />
        </button>
      ) : null}
      <Checkbox
        aria-label={`Complete ${item.title}`}
        checked={item.completed}
        className="mt-2"
        disabled={!canEdit || isPending || isConverted}
        onCheckedChange={(checked) => {
          onUpdate({ completed: checked === true });
        }}
      />
      {isEditing && canEdit ? (
        <ChecklistItemForm
          initialTitle={item.title}
          isPending={isPending}
          onCancel={() => {
            setIsEditing(false);
          }}
          onSave={(title) => {
            onUpdate({ title }, () => {
              setIsEditing(false);
            });
          }}
        />
      ) : (
        <ChecklistItemContent
          canEdit={canEdit}
          isConverted={isConverted}
          isPending={isPending}
          item={item}
          onEdit={() => {
            setIsEditing(true);
          }}
          storyTerm={storyTerm}
        />
      )}
      {isEditing && canEdit ? (
        <span aria-hidden className="w-[2.1rem] shrink-0" />
      ) : null}
      {canEdit && !isEditing ? (
        <ChecklistItemActions
          isConverted={isConverted}
          isPending={isPending}
          onConvert={onConvert}
          onDelete={onDelete}
          onEdit={() => {
            setIsEditing(true);
          }}
          storyTerm={storyTerm}
          title={item.title}
        />
      ) : null}
    </li>
  );
};
