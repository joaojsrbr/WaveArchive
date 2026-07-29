import { describe, expect, it } from 'vitest';
import type { CharacterTag, TeamMember } from '../types';
import { buildOfficialTeamSynergy } from './teamSynergy';

const support: CharacterTag = {
  id: 1,
  name: 'Support',
  description: 'Official support tag.',
  color: '74D3D0',
  iconPath: '',
};

const coordinated: CharacterTag = {
  id: 2,
  name: 'Coordinated Attack',
  description: 'Official coordinated attack tag.',
  color: 'DFBD78',
  iconPath: '',
};

const members = [member(1, 'Aemeath'), member(2, 'Zhezhi'), member(3, 'Shorekeeper')];

describe('buildOfficialTeamSynergy', () => {
  it('consolida tags oficiais repetidas e preserva suas fontes', () => {
    const profiles = new Map([
      [1, { extras: { tags: [coordinated] } }],
      [2, { extras: { tags: [coordinated, support] } }],
      [3, { extras: { tags: [support] } }],
    ]);

    const result = buildOfficialTeamSynergy(members, profiles);

    expect(result).toHaveLength(2);
    expect(result[0].members).toHaveLength(2);
    expect(result.find((item) => item.tag.id === coordinated.id)?.members).toEqual([
      'Aemeath',
      'Zhezhi',
    ]);
    expect(result.find((item) => item.tag.id === support.id)?.tag.description).toBe(
      support.description
    );
  });

  it('não cria funções quando a fonte não fornece tags', () => {
    expect(buildOfficialTeamSynergy(members, new Map())).toEqual([]);
  });
});

function member(characterId: number, characterName: string): TeamMember {
  return {
    slot: characterId,
    characterId,
    characterName,
    characterIcon: '',
    buildName: '',
    role: '',
    customRole: '',
  };
}
