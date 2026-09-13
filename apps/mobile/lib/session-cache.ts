type SessionCacheReset = () => Promise<void>;

const resets = new Set<SessionCacheReset>();

export const registerSessionCacheReset = (reset: SessionCacheReset) => {
  resets.add(reset);
  return () => {
    resets.delete(reset);
  };
};

export const resetSessionCache = async () => {
  await Promise.all([...resets].map((reset) => reset()));
};
