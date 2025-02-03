import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { toast } from "sonner";
import { Button, Flex, TextEditor } from "ui";
import { cn } from "lib";
import { useCommentStoryMutation } from "@/modules/story/hooks/comment-mutation";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";

export const CommentInput = ({
  storyId,
  parentId,
  commentId,
  className,
  initialComment,
  onCancel,
}: {
  storyId: string;
  parentId?: string;
  commentId?: string;
  className?: string;
  initialComment?: string;
  onCancel?: () => void;
}) => {
  const { mutateAsync } = useCommentStoryMutation();
  const { mutateAsync: updateComment } = useUpdateCommentMutation();

  const getPlaceHolder = () => {
    if (commentId) {
      return "Edit comment...";
    }
    if (parentId) {
      return "Reply to comment...";
    }
    return "Leave a comment...";
  };

  const getButtonLabel = () => {
    if (commentId) {
      return "Update";
    }
    if (parentId) {
      return "Reply";
    }
    return "Comment";
  };

  const editor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({
        placeholder: getPlaceHolder(),
      }),
    ],
    content: initialComment ?? "",
    editable: true,
  });

  const handleComment = async () => {
    const comment = editor?.getHTML() ?? "";
    if (editor?.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }

    if (commentId) {
      // update comment
      await updateComment({
        commentId,
        payload: { content: comment },
        storyId,
      }).then(() => {
        editor?.commands.clearContent();
        onCancel?.();
      });
    } else {
      await mutateAsync({
        storyId,
        payload: { comment, parentId: parentId ?? null },
      }).then(() => {
        editor?.commands.clearContent();
        onCancel?.();
      });
    }
  };

  return (
    <Flex
      className={cn(
        "ml-1 min-h-[6rem] w-full rounded-lg border border-gray-50 bg-gray-50/40 px-4 pb-4 text-[0.95rem] shadow-sm transition-shadow duration-200 ease-linear focus-within:shadow-lg dark:border-dark-200/80 dark:bg-dark-200/50 dark:shadow-dark-300/50",
        className,
      )}
      direction="column"
      gap={2}
      justify="between"
    >
      <TextEditor
        className="prose-base font-normal leading-6 antialiased"
        editor={editor}
      />
      <Flex gap={2} justify="end">
        {onCancel ? (
          <Button
            className="px-3"
            color="tertiary"
            onClick={onCancel}
            size="sm"
            variant="outline"
          >
            Cancel
          </Button>
        ) : null}
        <Button
          className="px-3"
          color="tertiary"
          disabled={editor?.isEmpty}
          onClick={handleComment}
          size="sm"
          variant="outline"
        >
          {getButtonLabel()}
        </Button>
      </Flex>
    </Flex>
  );
};
