type MayaMessageLimitInput = {
  isInternalUser: boolean;
  limit: number;
  totalMessages: number;
};

export const canSendMayaMessage = ({
  isInternalUser,
  limit,
  totalMessages,
}: MayaMessageLimitInput) => {
  if (isInternalUser) {
    return true;
  }

  return limit === Infinity || totalMessages < limit;
};

export const shouldShowMayaMessageLimit = (
  input: MayaMessageLimitInput,
) => {
  return !canSendMayaMessage(input);
};
