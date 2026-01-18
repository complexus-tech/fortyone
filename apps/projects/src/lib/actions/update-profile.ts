"use server";

import type { UpdateProfile } from "@/lib/actions/users/update";
import { updateProfile as updateProfileAction } from "@/lib/actions/users/update";

export async function updateProfile(updates: UpdateProfile) {
  return updateProfileAction(updates);
}
