import type { Comment, Member } from "@/types";
import { FORMER_USER_NAME, isFormerUser } from "@/lib/former-user";

export function resolveCommentAuthor(
  { userId, user }: Pick<Comment, "userId" | "user">,
  members: readonly Member[],
) {
  if (isFormerUser(userId)) {
    return { name: FORMER_USER_NAME, avatarUrl: undefined };
  }
  const person =
    user?.id === userId ? user : members.find((member) => member.id === userId);
  return {
    name: person?.username || person?.fullName || "Unknown",
    avatarUrl: person?.avatarUrl ?? undefined,
  };
}
