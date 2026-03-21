import { slugify } from "@/utils";

export const GITHUB_BRANCH_FORMATS = [
  "username/identifier-title",
  "identifier-title",
  "identifier/title",
] as const;

export type GitHubBranchFormat = (typeof GITHUB_BRANCH_FORMATS)[number];

type BuildBranchNameInput = {
  format: string;
  username?: string | null;
  teamCode?: string | null;
  sequenceId: number;
  title: string;
};

const sanitizeSegment = (value?: string | null) => slugify(value ?? "");

export const buildGitBranchName = ({
  format,
  username,
  teamCode,
  sequenceId,
  title,
}: BuildBranchNameInput) => {
  const identifier = `${teamCode ?? ""}-${sequenceId}`.toLowerCase();
  const titleSlug = sanitizeSegment(title.slice(0, 32));
  const usernameSlug = sanitizeSegment(username);

  switch (format) {
    case "identifier-title":
      return `${identifier}-${titleSlug}`.replace(/-$/, "");
    case "identifier/title":
      return `${identifier}/${titleSlug}`.replace(/\/$/, "");
    case "username/identifier-title":
    default: {
      if (!usernameSlug) {
        return `${identifier}-${titleSlug}`.replace(/-$/, "");
      }
      return `${usernameSlug}/${identifier}-${titleSlug}`.replace(/-$/, "");
    }
  }
};
