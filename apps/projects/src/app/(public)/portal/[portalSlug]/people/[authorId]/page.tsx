import { notFound } from "next/navigation";
import { PublicPortalAuthorProfilePage } from "@/modules/public-portal";
import {
  getPublicContributorComments,
  getPublicContributorOrNotFound,
  getPublicPortalOrNotFound,
} from "@/modules/public-portal/query";
import { getPublicPortalViewer } from "@/modules/public-portal/viewer";

const PAGE_SIZE = 20;
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

type PageProps = {
  params: Promise<{ authorId: string; portalSlug: string }>;
  searchParams: Promise<{ tab?: string | string[] }>;
};

export default async function PublicPortalAuthorRoute({
  params,
  searchParams,
}: PageProps) {
  const { authorId, portalSlug } = await params;
  const query = await searchParams;

  if (!UUID_PATTERN.test(authorId)) notFound();

  const initialTab = query.tab === "comments" ? "comments" : "feedback";
  const [portal, viewer, contributor, initialComments] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, {
      authorId,
      page: 1,
      pageSize: PAGE_SIZE,
      sort: "newest",
    }),
    getPublicPortalViewer(portalSlug),
    getPublicContributorOrNotFound(portalSlug, authorId),
    initialTab === "comments"
      ? getPublicContributorComments(portalSlug, authorId, 1, PAGE_SIZE)
      : Promise.resolve(null),
  ]);

  return (
    <PublicPortalAuthorProfilePage
      authorId={authorId}
      contributor={contributor}
      initialComments={initialComments}
      initialTab={initialTab}
      portal={portal}
      viewer={viewer}
    />
  );
}
