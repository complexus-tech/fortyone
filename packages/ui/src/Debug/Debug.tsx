export const Debug = ({ data }: { data: any }) => {
  if (process.env.NODE_ENV === "production") {
    return null;
  }
  return <pre className="dark:text-white">{JSON.stringify(data, null, 2)}</pre>;
};
