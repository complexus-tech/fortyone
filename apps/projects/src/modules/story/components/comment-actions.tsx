import {
  CheckIcon,
  DeleteIcon,
  EditIcon,
  LinkIcon,
  MoreHorizontalIcon,
} from "icons";
import { Menu } from "ui";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";

export const CommentActions = ({
  commentId,
  storyId,
  isOwner,
  isRoot,
  disabled,
  resolved,
  onEdit,
  onDelete,
  onResolve,
}: {
  commentId: string;
  storyId: string;
  isOwner: boolean;
  isRoot: boolean;
  disabled: boolean;
  resolved: boolean;
  onEdit: () => void;
  onDelete: () => void;
  onResolve: () => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const copyLink = async () => {
    const url = new URL(
      withWorkspace(`/work/${storyId}`),
      window.location.origin,
    );
    url.searchParams.set("comment", commentId);
    try {
      await navigator.clipboard.writeText(url.toString());
      toast.success("Comment link copied");
    } catch {
      toast.error("Could not copy the comment link");
    }
  };
  return (
    <Menu>
      <Menu.Button>
        <button
          aria-label="Comment actions"
          className="text-text-muted hover:bg-surface-muted focus-visible:ring-primary ml-1 flex size-7 items-center justify-center rounded-md opacity-60 outline-none hover:opacity-100 focus-visible:opacity-100 focus-visible:ring-2 disabled:opacity-30"
          disabled={disabled}
          type="button"
        >
          <MoreHorizontalIcon className="size-4" />
        </button>
      </Menu.Button>
      <Menu.Items align="start" className="min-w-44">
        <Menu.Group>
          <Menu.Item
            onSelect={() => {
              void copyLink();
            }}
          >
            <LinkIcon className="size-4" /> Copy link
          </Menu.Item>
          {isRoot ? (
            <Menu.Item onSelect={onResolve}>
              <CheckIcon className="size-4" />{" "}
              {resolved ? "Reopen thread" : "Resolve thread"}
            </Menu.Item>
          ) : null}
        </Menu.Group>
        {isOwner ? (
          <>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item onSelect={onEdit}>
                <EditIcon className="size-4" /> Edit
              </Menu.Item>
              <Menu.Item className="text-danger" onSelect={onDelete}>
                <DeleteIcon className="text-danger size-4" /> Delete
              </Menu.Item>
            </Menu.Group>
          </>
        ) : null}
      </Menu.Items>
    </Menu>
  );
};
