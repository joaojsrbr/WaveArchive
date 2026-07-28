import type { Character, CharacterFilter, CharacterProfile, Team } from '../types';

const elementNames: Record<number, string> = {
  1: 'Glacio',
  2: 'Fusion',
  3: 'Electro',
  4: 'Aero',
  5: 'Spectro',
  6: 'Havoc',
};
const weaponNames: Record<number, string> = {
  1: 'Broadblade',
  2: 'Sword',
  3: 'Pistols',
  4: 'Gauntlets',
  5: 'Rectifier',
};

const sourceCharacters = [
  [1210, 'Aemeath', 'Aemeath', 5, 2, 2, 17],
  [1606, 'Roccia', 'Roccia', 5, 6, 4, 58],
  [1206, 'Brant', 'Brant', 5, 2, 2, 13],
  [1409, 'Cartethyia', 'Cartethyia', 5, 4, 2, 37],
  [1503, 'Verina', 'Nature Calling', 5, 5, 5, 44],
  [1102, 'Sanhua', 'Snow Waltz', 4, 1, 2, 0],
  [1304, 'Jinhsi', "Jinhsi's Title", 5, 5, 1, 23],
  [1302, 'Yinlin', 'Lightning of Execution', 5, 3, 5, 21],
  [1203, 'Encore', 'Wooly-Counting Game', 5, 2, 5, 10],
] as const;

export const previewCharacters: Character[] = sourceCharacters.map(
  ([id, name, nickname, rarity, element, weapon, apiOrder]) => ({
    id,
    name,
    nickname,
    rarity,
    elementCode: element,
    element: elementNames[element],
    weaponTypeCode: weapon,
    weaponType: weaponNames[weapon],
    iconPath: `/cache/characters/${id}/icon.webp`,
    backgroundPath: `/cache/characters/${id}/background.webp`,
    owned: false,
    level: 1,
    sequence: 0,
    favorite: false,
    gameVersion: '3.6.1',
    apiOrder,
  })
);

const rocciaTags = [
  {
    id: 3,
    name: 'Concerto Efficiency',
    description: 'Quick concerto energy regeneration',
    iconPath: '',
    color: 'ff4040',
  },
  {
    id: 5,
    name: 'Heavy Attack DMG',
    description: 'Deals higher Heavy Attack DMG',
    iconPath: '',
    color: 'ffde73',
  },
  {
    id: 8,
    name: 'Traction',
    description: 'Pulls targets within range towards a specific position',
    iconPath: '',
    color: '77adff',
  },
  {
    id: 14,
    name: 'Havoc DMG Amplification',
    description: 'Provides Havoc DMG Amplification for a specific Resonator in the team',
    iconPath: '',
    color: 'ff7777',
  },
  {
    id: 20,
    name: 'Basic Attack DMG Amplification',
    description: 'Provides Basic Attack DMG Amplification for a specific Resonator in the team',
    iconPath: '',
    color: 'ff7777',
  },
];

const aemeathSkills = [
  ['1', 'Normal Attack', 'Infinity Calibration'],
  ['2', 'Resonance Skill', 'Shared Voyage'],
  ['3', 'Resonance Liberation', 'Towards the Daybreak'],
  ['4', 'Inherent Skill', 'Before All Sounds'],
  ['5', 'Inherent Skill', 'Between the Stars'],
  ['6', 'Intro Skill', 'Overture of Departure'],
  ['7', 'Forte Circuit', 'To Sculpt the Silence'],
  ['8', 'Outro Skill', 'Silent Protection'],
  ['9', '', 'Crit. Rate+'],
  ['10', '', 'ATK+'],
  ['11', '', 'ATK+'],
  ['12', '', 'Crit. Rate+'],
  ['13', '', 'Crit. Rate+'],
  ['14', '', 'ATK+'],
  ['15', '', 'ATK+'],
  ['16', '', 'Crit. Rate+'],
  ['17', 'Tune Break', 'Unlanded Melody'],
] as const;

const aemeathTree = [
  ['1', 2, 1, [], [], 0],
  ['2', 2, 2, [], [], 0],
  ['3', 2, 3, [], [], 0],
  ['4', 3, 2, [7], [], 0],
  ['5', 3, 3, [4], [121001, 121002], 0],
  ['6', 2, 4, [], [], 0],
  ['7', 1, 1, [], [121001, 121002], 0],
  ['8', 3, 1, [], [121001, 121002], 0],
  ['9', 4, 1, [1], [], 500006],
  ['10', 4, 1, [2], [], 500005],
  ['11', 4, 1, [3], [], 500005],
  ['12', 4, 1, [6], [], 500006],
  ['13', 4, 2, [9], [], 500008],
  ['14', 4, 2, [10], [], 500007],
  ['15', 4, 2, [11], [], 500007],
  ['16', 4, 2, [12], [], 500008],
  ['17', 3, 1, [], [121001, 121002], 0],
] as const;

let previewTeams: Team[] = [];

export function listPreviewCharacters(filter: CharacterFilter) {
  const query = filter.query.trim().toLocaleLowerCase();
  const result = previewCharacters.filter(
    (character) =>
      (!query || `${character.name} ${character.nickname}`.toLocaleLowerCase().includes(query)) &&
      (!filter.element || character.elementCode === filter.element) &&
      (!filter.rarity || character.rarity === filter.rarity) &&
      (!filter.ownedOnly || character.owned) &&
      (!filter.favorites || character.favorite)
  );
  return [...result].sort((left, right) =>
    filter.sort === 'api' ? left.apiOrder - right.apiOrder : left.name.localeCompare(right.name)
  );
}

export function getPreviewCharacter(id: number): CharacterProfile {
  const character = previewCharacters.find((item) => item.id === id);
  if (!character) throw new Error('Personagem não encontrado.');
  if (id === 1210) {
    return {
      character,
      description: '',
      birthday: '',
      gender: '',
      region: '',
      faction: '',
      talentName: '',
      talentDescription: '',
      skills: aemeathSkills.map(([nodeId, type, name], index) => ({
        nodeId,
        type,
        name,
        description: '',
        iconPath: '',
        levelsJson: 'null',
        sortOrder: index + 1,
      })),
      chains: [],
      progression: {
        ascensions: [],
        skills: aemeathSkills.map(([nodeId, type, name]) => ({
          nodeId,
          nodeType: 0,
          type,
          name,
          iconPath: '',
          maxLevel: ['1', '2', '3', '6', '7'].includes(nodeId) ? 10 : 1,
          unlockCosts: [],
          levelCosts: [],
          values:
            nodeId === '1'
              ? [
                  {
                    name: 'Basic Attack - Aemeath Stage 1 DMG',
                    values: [
                      '23.31%',
                      '25.23%',
                      '27.14%',
                      '29.81%',
                      '31.73%',
                      '33.92%',
                      '36.98%',
                      '40.04%',
                      '43.10%',
                      '46.35%',
                    ],
                  },
                ]
              : [],
        })),
        levelExp: [],
        stats: [],
      },
      extras: {
        tags: [],
        stories: [],
        goods: [],
        forte: {
          iconPath: '',
          descriptions: [],
          features: [],
          actions: [
            {
              name: 'Forte Gauge',
              description: 'Restore Synchronization Rate through attacks.',
              inputs: [],
              images: [],
            },
            {
              name: 'Instructions',
              description: 'Form Switch: {0}',
              inputs: ['技能1'],
              images: [],
            },
          ],
        },
        weakness: { buildUp: 0, buildUpMax: 0, totalBonus: 0, breakRatio: 0, mastery: 0 },
        skillBranches: [
          {
            id: 121001,
            name: 'Resonance Mode - Tune Rupture',
            description:
              'When in Resonance Mode - Tune Rupture, Aemeath can inflict Tune Rupture - Shifting on the target, and Seraphic Duet can deal additional Tune Rupture DMG.',
            iconPath: '',
          },
          {
            id: 121002,
            name: 'Resonance Mode - Fusion Burst',
            description:
              'When in Resonance Mode - Fusion Burst, the required number of stacks to trigger Fusion Burst reduces. Aemeath can inflict Fusion Burst on the target, and Seraphic Duet can trigger Fusion Burst based on its max stack limit without removing its stacks.',
            iconPath: '',
          },
        ],
        skillTree: aemeathTree.map(
          ([nodeId, nodeType, coordinate, parentNodes, branchIds, unlockCondition]) => ({
            nodeId,
            nodeType,
            coordinate,
            parentNodes: [...parentNodes],
            branchIds: [...branchIds],
            unlockCondition,
          })
        ),
      },
    };
  }
  return {
    character,
    description: '',
    birthday: '',
    gender: '',
    region: '',
    faction: '',
    talentName: '',
    talentDescription: '',
    skills: [],
    chains: [],
    progression: { ascensions: [], skills: [], levelExp: [], stats: [] },
    extras: {
      tags: id === 1606 ? rocciaTags : [],
      stories: [],
      goods: [],
      forte: { iconPath: '', descriptions: [], features: [], actions: [] },
      weakness: { buildUp: 0, buildUpMax: 0, totalBonus: 0, breakRatio: 0, mastery: 0 },
      skillBranches: [],
      skillTree: [],
    },
  };
}

export function listPreviewTeams() {
  return previewTeams.map((team) => ({
    ...team,
    members: team.members.map((member) => ({ ...member })),
  }));
}

export function savePreviewTeam(team: Team) {
  const now = new Date().toISOString();
  const saved = {
    ...team,
    id: team.id || Date.now(),
    createdAt: team.createdAt || now,
    updatedAt: now,
  };
  previewTeams = [saved, ...previewTeams.filter((item) => item.id !== saved.id)];
  return saved;
}
