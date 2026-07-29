import type { OwnedEcho, Sonata } from '../types';

export type SonataTier = {
  pieces: number;
  description: string;
  active: boolean;
  missing: number;
};

export type SonataProgressItem = {
  sonata: Sonata;
  count: number;
  tiers: SonataTier[];
  hasActiveEffect: boolean;
  piecesWithoutActiveEffect: number;
};

export function buildSonataProgress(
  echoes: readonly OwnedEcho[],
  sonatas: readonly Sonata[]
): SonataProgressItem[] {
  return sonatas
    .map((sonata) => {
      const count = echoes.filter((echo) => echo.sonataId === sonata.id).length;
      const tiers = [
        { pieces: 2, description: sonata.twoPiece },
        { pieces: 5, description: sonata.fivePiece },
      ]
        .filter((tier) => tier.description.trim())
        .map((tier) => ({
          ...tier,
          active: count >= tier.pieces,
          missing: Math.max(0, tier.pieces - count),
        }));
      const hasActiveEffect = tiers.some((tier) => tier.active);
      return {
        sonata,
        count,
        tiers,
        hasActiveEffect,
        piecesWithoutActiveEffect: hasActiveEffect ? 0 : count,
      };
    })
    .filter((item) => item.count > 0)
    .sort(
      (left, right) => right.count - left.count || left.sonata.name.localeCompare(right.sonata.name)
    );
}
