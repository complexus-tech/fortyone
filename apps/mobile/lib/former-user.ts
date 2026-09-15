// Reserved backend identity for retained workspace contributions after deletion.
export const FORMER_USER_ID = "ffffffff-ffff-4fff-8fff-ffffffffffff";
export const FORMER_USER_NAME = "Former user";
export const isFormerUser = (id?: string | null) => id === FORMER_USER_ID;
