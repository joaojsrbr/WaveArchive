import type { Character } from '../types';

export function isRoverCharacter(character?: Pick<Character, 'name'>): boolean {
  return character?.name.trim().toLocaleLowerCase('en-US').startsWith('rover:') ?? false;
}

const maleRoverIDs = new Set([1309, 1406, 1501, 1605]);
const femaleRoverIDs = new Set([1310, 1408, 1502, 1604]);

export function roverGender(
  character: Pick<Character, 'id' | 'gender'>
): 'male' | 'female' | 'unknown' {
  const value = character.gender?.trim().toLocaleLowerCase('en-US');
  if (value === 'male' || value === 'masculino') return 'male';
  if (value === 'female' || value === 'feminino') return 'female';
  if (maleRoverIDs.has(character.id)) return 'male';
  if (femaleRoverIDs.has(character.id)) return 'female';
  return 'unknown';
}
