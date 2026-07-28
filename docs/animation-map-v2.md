# WaveArchive Animation Map v2

## Global

- Tier: 2, JS-enhanced where it adds orientation; CSS elsewhere.
- Route change: old content fades 80 ms, new content enters 280 ms with a
  maximum 12 px vertical transform.
- Lists: filtering and sorting use FLIP-style position transforms.
- Modals and drawers: opacity plus horizontal/vertical transform only.
- Reduced motion: remove spatial transforms and retain immediate opacity/state
  changes.
- Signature: selecting a list entity visually transfers it into the active
  workspace without moving keyboard focus.

## Page patterns

### Overview

- Summary rows reveal once on entry.
- Recent activity updates in place without replaying page animation.

### Catalogs

Applies to Characters, Weapons, Echoes, and Sonata Effects.

- Filter results reposition rather than disappear abruptly.
- List/grid mode preserves the selected item and scroll position.
- Hover adds only local image emphasis and action visibility.

### Character detail

- Official artwork and identity appear together on entry.
- Tab content crossfades in place; tab underline moves using transform.
- Forte and skill-tree dependencies reveal only after their tab is selected.

### Teams

- Saved team list is stable and sortable.
- Applying an official/user preset fills the three slots in sequence.
- Selecting a character from the library transfers the artwork to the chosen
  slot and updates only fields supported by synchronized data.
- Reordering members uses direct manipulation plus keyboard alternatives.

### Builds, Comparator, and Calculator

- Selection panes remain static.
- Result changes use number/content replacement, not decorative count-up.
- Comparison differences reveal by row.

### Rotations

- Timeline focus follows the selected action.
- Playback motion is optional and never blocks editing.

### Planner, Materials, Account, and Convene

- Progress changes animate through transform/clip-path.
- Lists use stable rows and preserve scroll position.

### Assistant and Settings

- Conversation/status changes use restrained opacity transitions.
- Forms never animate validation errors away from their fields.
