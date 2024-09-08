export const JsonLd = ({ children }: { children: Record<string, any> }) => {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(children) }}
    />
  );
};
