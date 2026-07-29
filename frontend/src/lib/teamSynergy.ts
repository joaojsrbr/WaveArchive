import type { CharacterTag, TeamMember } from '../types';

export type OfficialSynergyItem = {
  key: string;
  tag: CharacterTag;
  members: string[];
};

type TaggedProfile = {
  extras?: {
    tags?: CharacterTag[];
  };
};

export function buildOfficialTeamSynergy(
  members: readonly TeamMember[],
  profiles: ReadonlyMap<number, TaggedProfile>
) {
  const grouped = new Map<string, OfficialSynergyItem>();
  members.forEach((member) => {
    const profile = profiles.get(member.characterId);
    const seen = new Set<string>();
    (profile?.extras?.tags || []).forEach((tag) => {
      const key = String(tag.id || tag.name.trim().toLocaleLowerCase('pt-BR'));
      if (!tag.name.trim() || seen.has(key)) return;
      seen.add(key);
      const current = grouped.get(key);
      if (current) {
        if (!current.members.includes(member.characterName)) {
          current.members.push(member.characterName);
        }
      } else {
        grouped.set(key, { key, tag, members: [member.characterName] });
      }
    });
  });
  return [...grouped.values()].sort(
    (left, right) =>
      right.members.length - left.members.length || left.tag.name.localeCompare(right.tag.name)
  );
}
