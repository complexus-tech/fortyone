export default function SharedDocumentNotFound() {
  return (
    <main className="flex min-h-dvh items-center justify-center px-6">
      <div className="max-w-md text-center">
        <h1 className="mb-3 text-2xl font-semibold">Document unavailable</h1>
        <p className="text-text-muted">
          This link may have been revoked, or the document is no longer
          available.
        </p>
      </div>
    </main>
  );
}
