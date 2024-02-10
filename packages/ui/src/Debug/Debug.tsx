export const Debug = ({ data }: { data: any }) => {
  return <pre className="dark:text-white">{JSON.stringify(data, null, 2)}</pre>;
};
