# WaveArchive Design System v2

## Product principle

WaveArchive is a local-first desktop archive and planning tool for Wuthering
Waves. The interface must make official data feel cinematic without hiding,
inventing, or embellishing facts.

## Brand metaphors

### Polished mineral archive

The product feels like layers of dark mineral glass: deep, precise, durable,
and legible. Surfaces separate content through spacing and tonal depth before
using borders or shadows.

### Resonance table

Characters, materials, weapons, and skills behave like pieces on a tactical
table. A visual connection is shown only when a real data relationship exists.

### Controlled gallery

Official character artwork can be cinematic, but controls and numeric data
remain restrained, readable, and predictable.

## Visual language

- Dark mineral base with graphite surfaces.
- Warm ivory for primary text and restrained sea-glass cyan for interaction.
- Element colors are data colors, never generic decoration.
- Pale gold is reserved for rarity and primary confirmation.
- Typography is compact and editorial; body copy remains 14–16 px.
- Lucide is the only interface icon family.
- Layout uses spacing and alignment first, separators second, tinted surfaces
  third, borders fourth, and shadows last.
- No dashed empty states, giant page titles, emoji, ASCII icons, invented
  metrics, nested card stacks, or permanent decorative motion.

## Motion language

- `ease-primary`: `cubic-bezier(.2,.8,.2,1)`
- `ease-reveal`: `cubic-bezier(.16,1,.3,1)`
- `ease-transition`: `cubic-bezier(.65,0,.35,1)`
- Direct controls respond in 140–180 ms.
- Panels and route content settle in 280–360 ms.
- Animate only transform, opacity, filter, and clip-path.
- Never use `transition: all`.
- All motion must degrade cleanly under `prefers-reduced-motion`.

## Signature interaction

When a user selects an entity from a list, the relevant workspace layers
reorganize through a FLIP-style transform. The selected artwork moves into its
destination while related official data updates in place. Keyboard focus must
remain anchored to the selected item.

## Shared shell

- Compact, collapsible navigation grouped by workflow rather than one long list.
- Persistent sync/source state that never competes with the primary action.
- Search and command access remain reachable from the shell.
- Content width follows the task: dense lists use the available viewport;
  reading views keep a comfortable line length.
- All pages share the same header height, focus ring, input system, buttons,
  tables, filters, empty states, feedback, and loading states.

## Teams workspace

- Saved teams appear as a list, not oversized standalone cards.
- A preset selector is available before manual composition.
- Presets must come from synchronized guide data or be explicitly saved by the
  user. Every preset displays its source.
- The character library is searchable and filterable.
- The active composition uses three slots in one shared visual stage.
- Character tags come from the synchronized API `CharacterExtras.tags`.
- Member identity uses API name, icon/background, rarity, element, weapon type,
  and tags. Missing values are shown as unavailable.
- Do not display invented build names, synergy scores, buffs, rotations, or
  role coverage.
- Saved user builds may be linked, but the UI must label them as user data.

## Accessibility and performance

- Normal text contrast is at least 4.5:1.
- Interactive component boundaries and focus indicators are at least 3:1.
- Selection is communicated by text/icon/shape as well as color.
- All core flows work with keyboard controls.
- Use native scrolling and preserve touch momentum.
- Official images use stable dimensions, lazy loading where appropriate, and a
  non-layout-shifting fallback.
