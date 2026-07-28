import type { Character } from "../types";

export function isRoverCharacter(character?: Pick<Character, "name">): boolean {
  return character?.name.trim().toLocaleLowerCase("en-US").startsWith("rover:") ?? false;
}
