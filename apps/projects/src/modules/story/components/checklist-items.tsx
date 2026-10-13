import type { DragEndEvent } from "@dnd-kit/core";
import {
  closestCenter,
  DndContext,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import type { useChecklistMutation } from "../hooks/checklist";
import type { StoryChecklistItem } from "../types/checklist";
import { ChecklistItem } from "./checklist-item";

export const ChecklistItems = ({
  id,
  items,
  canEdit,
  mutation,
  storyTerm,
}: {
  id: string;
  items: StoryChecklistItem[];
  canEdit: boolean;
  mutation: ReturnType<typeof useChecklistMutation>;
  storyTerm: string;
}) => {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    if (!canEdit || mutation.isPending || !over || active.id === over.id)
      return;
    const from = items.findIndex(({ id }) => id === active.id);
    const to = items.findIndex(({ id }) => id === over.id);
    if (from < 0 || to < 0) return;
    mutation.mutate({
      type: "reorder",
      itemIds: arrayMove(items, from, to).map(({ id }) => id),
    });
  };

  if (!items.length) return null;
  return (
    <DndContext
      collisionDetection={closestCenter}
      id={id}
      onDragEnd={handleDragEnd}
      sensors={sensors}
    >
      <SortableContext
        items={items.map(({ id }) => id)}
        strategy={verticalListSortingStrategy}
      >
        <ul aria-label="Checklist items" className="mt-2">
          {items.map((item) => (
            <ChecklistItem
              canEdit={canEdit}
              isPending={mutation.isPending}
              item={item}
              key={item.id}
              onConvert={() => {
                mutation.mutate({ type: "convert", itemId: item.id });
              }}
              onDelete={() => {
                mutation.mutate({ type: "delete", itemId: item.id });
              }}
              onUpdate={(updates, onSuccess) => {
                mutation.mutate(
                  { type: "update", itemId: item.id, updates },
                  { onSuccess },
                );
              }}
              storyTerm={storyTerm}
            />
          ))}
        </ul>
      </SortableContext>
    </DndContext>
  );
};
