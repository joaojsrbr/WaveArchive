import { describe, expect, it } from 'vitest';
import type { OwnedEcho, Sonata } from '../types';
import { buildSonataProgress } from './sonata';

const sonata: Sonata = {
  id: 7,
  name: 'Celestial Light',
  iconPath: '',
  twoPiece: 'Aumenta Spectro DMG.',
  fivePiece: 'Aumenta Spectro DMG após usar Intro Skill.',
  gameVersion: '3.6.1',
};

function echo(id: number): OwnedEcho {
  return {
    id,
    echoId: id,
    echoName: `Echo ${id}`,
    iconPath: '',
    cost: 1,
    mainStat: '',
    substatsJson: '[]',
    level: 25,
    sonataId: sonata.id,
    sonataName: sonata.name,
    characterName: '',
    locked: false,
    favorite: false,
    note: '',
  };
}

describe('buildSonataProgress', () => {
  it('mantém as descrições oficiais e ativa o efeito de duas peças', () => {
    const [progress] = buildSonataProgress([echo(1), echo(2)], [sonata]);

    expect(progress.count).toBe(2);
    expect(progress.tiers).toEqual([
      {
        pieces: 2,
        description: sonata.twoPiece,
        active: true,
        missing: 0,
      },
      {
        pieces: 5,
        description: sonata.fivePiece,
        active: false,
        missing: 3,
      },
    ]);
    expect(progress).toMatchObject({ hasActiveEffect: true, piecesWithoutActiveEffect: 0 });
  });

  it('ativa os dois efeitos com cinco peças', () => {
    const progress = buildSonataProgress([1, 2, 3, 4, 5].map(echo), [sonata])[0];

    expect(progress.tiers.every((tier) => tier.active)).toBe(true);
  });

  it('identifica uma peÃ§a que ainda nÃ£o ativa efeito', () => {
    const progress = buildSonataProgress([echo(1)], [sonata])[0];

    expect(progress).toMatchObject({
      count: 1,
      hasActiveEffect: false,
      piecesWithoutActiveEffect: 1,
    });
  });
});
