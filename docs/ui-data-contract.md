# WaveArchive UI Data Contract

The redesigned UI must distinguish synchronized source data from user-created
data. It must never infer a label, relationship, score, or recommendation that
is not represented by one of these sources.

## Nanoka character API

Allowed character presentation data:

- id, name, nickname, description
- rarity, element, weapon type
- icon and background artwork
- character information
- tags with name, description, icon, and color
- stats and weakness values
- Forte descriptions, features, inputs, and instructions
- skill tree nodes, parents, branches, conditions, descriptions, and costs
- resonance chains
- ascension, level, and skill materials
- recommended weapon identifiers

## Official guide API

Allowed preset data:

- guide id and name
- source and language
- synchronized team member identifiers
- like count as source metadata, not as a quality score
- raw guide data only when the corresponding field is explicitly mapped

Every preset must visibly identify its source. If the guide cannot be resolved
to three local characters, it is incomplete and must not silently substitute
other characters.

## User data

Allowed user-created data:

- owned/favorite state and account progress
- saved teams and their name, notes, order, favorite/locked state
- manual member role
- saved builds and build links
- rotations, buffs, planner goals, and account records

User-created data must be labeled as such when shown beside synchronized API
data.

## Missing data

- Render “Não disponível” or omit the section.
- Never fabricate a role, build name, synergy, buff, stat, recommendation, tag,
  rotation, or material relationship.
- Derived values are allowed only when the formula and inputs already exist in
  the application and the result is labeled as calculated.
