import { DocumentsPage } from "@/modules/teams/documents";
import { getDocuments } from "@/modules/teams/documents/queries/get-documents";

export default async function Page() {
  const documents = await getDocuments();

  return <DocumentsPage documents={documents} />;
}
