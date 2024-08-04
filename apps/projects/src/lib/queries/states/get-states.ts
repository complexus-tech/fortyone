import { get } from "@/lib/http";
import { State } from "@/types/states";

export const getStates = async () => {
  const states = await get<State[]>("states");
  return states;
};
