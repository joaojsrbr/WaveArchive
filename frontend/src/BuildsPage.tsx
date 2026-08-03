import { useEffect, useMemo, useState } from 'react';
import {
  Check,
  ChevronRight,
  Copy,
  ImageDown,
  Lock,
  Plus,
  Save,
  Search,
  Star,
  Swords,
  Trash2,
  UserRound,
  Waves,
  X,
} from 'lucide-react';
import {
  deleteBuild,
  duplicateBuild,
  getCharacter,
  listBuilds,
  listCharacters,
  listEchoes,
  listOwnedEchoes,
  listSonatas,
  listWeapons,
  restoreBuild,
  saveBuild,
  saveOwnedEcho,
} from './lib/backend';
import { readOpenTarget } from './lib/navigation';
import { useContextualShortcuts } from './lib/contextualShortcuts';
import { isRoverCharacter } from './lib/characters';
import { buildSonataProgress, type SonataProgressItem } from './lib/sonata';
import { FilterField, FilterRange } from './AdvancedFilters';
import { LibraryFilterBar } from './LibraryFilterBar';
import { BuildExportCard } from './BuildExportCard';
import type { Build, Character, Echo, OwnedEcho, Sonata, Weapon } from './types';

type Target = { kind: 'character' } | { kind: 'weapon' } | { kind: 'echo'; slot: number };
type EchoSubstatChoice = { key: string; label: string; values: readonly string[] };
type ParsedEchoSubstat = { key: string; value: string };
type EchoLibrarySource = 'catalog' | 'inventory';

const buildSkillFields: ReadonlyArray<{
  key:
    'normalAttackLevel' | 'resonanceSkillLevel' | 'forteLevel' | 'liberationLevel' | 'introLevel';
  label: string;
}> = [
  { key: 'normalAttackLevel', label: 'Ataque Normal' },
  { key: 'resonanceSkillLevel', label: 'Habilidade' },
  { key: 'forteLevel', label: 'Forte' },
  { key: 'liberationLevel', label: 'Liberação' },
  { key: 'introLevel', label: 'Intro' },
];

const characterFilter = {
  query: '',
  element: 0,
  rarity: 0,
  ownedOnly: false,
  favorites: false,
  sort: 'api' as const,
};
const weaponFilter = {
  query: '',
  type: 0,
  rarity: 0,
  ownedOnly: false,
  favorites: false,
  sort: 'id' as const,
};
const echoFilter = { query: '', cost: 0, sonataId: 0, ownedOnly: false, sort: 'id' as const };
const commonPercentRolls = [
  '6.4%',
  '7.1%',
  '7.9%',
  '8.6%',
  '9.4%',
  '10.1%',
  '10.9%',
  '11.6%',
] as const;
const echoSubstatChoices: readonly EchoSubstatChoice[] = [
  { key: 'hp-flat', label: 'HP', values: ['320', '360', '390', '430', '470', '510', '540', '580'] },
  { key: 'atk-flat', label: 'ATK', values: ['30', '40', '50', '60'] },
  { key: 'def-flat', label: 'DEF', values: ['40', '50', '60', '70'] },
  { key: 'hp-percent', label: 'HP%', values: commonPercentRolls },
  { key: 'atk-percent', label: 'ATK%', values: commonPercentRolls },
  {
    key: 'def-percent',
    label: 'DEF%',
    values: ['8.1%', '9.0%', '10.0%', '10.9%', '11.8%', '12.8%', '13.8%', '14.7%'],
  },
  {
    key: 'crit-rate',
    label: 'CRIT Rate',
    values: ['6.3%', '6.9%', '7.5%', '8.1%', '8.7%', '9.3%', '9.9%', '10.5%'],
  },
  {
    key: 'crit-dmg',
    label: 'CRIT DMG',
    values: ['12.6%', '13.8%', '15.0%', '16.2%', '17.4%', '18.6%', '19.8%', '21.0%'],
  },
  {
    key: 'energy-regen',
    label: 'Energy Regen',
    values: ['6.8%', '7.6%', '8.4%', '9.2%', '10.0%', '10.8%', '11.6%', '12.4%'],
  },
  { key: 'basic-dmg', label: 'Basic Attack DMG', values: commonPercentRolls },
  { key: 'heavy-dmg', label: 'Heavy Attack DMG', values: commonPercentRolls },
  { key: 'skill-dmg', label: 'Resonance Skill DMG', values: commonPercentRolls },
  { key: 'liberation-dmg', label: 'Resonance Liberation DMG', values: commonPercentRolls },
] as const;

export function BuildsPage({
  version,
  onError,
}: {
  version: string;
  onError: (message: string) => void;
}) {
  const [builds, setBuilds] = useState<Build[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [weapons, setWeapons] = useState<Weapon[]>([]);
  const [echoes, setEchoes] = useState<Echo[]>([]);
  const [ownedEchoes, setOwnedEchoes] = useState<OwnedEcho[]>([]);
  const [sonatas, setSonatas] = useState<Sonata[]>([]);
  const [draft, setDraft] = useState<Build>(() => emptyBuild(version));
  const [target, setTarget] = useState<Target>({ kind: 'character' });
  const [query, setQuery] = useState('');
  const [sourceQuery, setSourceQuery] = useState('');
  const [costFilter, setCostFilter] = useState(0);
  const [elementFilter, setElementFilter] = useState(0);
  const [rarityFilter, setRarityFilter] = useState(0);
  const [ownershipFilter, setOwnershipFilter] = useState<'all' | 'owned' | 'missing'>('all');
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [characterSort, setCharacterSort] = useState<'api' | 'name' | 'rarity' | 'level'>('api');
  const [weaponRarityFilter, setWeaponRarityFilter] = useState(0);
  const [weaponSubStatFilter, setWeaponSubStatFilter] = useState('');
  const [weaponOwnershipFilter, setWeaponOwnershipFilter] = useState<'all' | 'owned' | 'missing'>(
    'all'
  );
  const [weaponMinAtk, setWeaponMinAtk] = useState(0);
  const [weaponMaxAtk, setWeaponMaxAtk] = useState(0);
  const [weaponFavoritesOnly, setWeaponFavoritesOnly] = useState(false);
  const [weaponSort, setWeaponSort] = useState<'rarity' | 'atk' | 'name' | 'api'>('rarity');
  const [sonataFilter, setSonataFilter] = useState(0);
  const [echoClassFilter, setEchoClassFilter] = useState('');
  const [echoTypeFilter, setEchoTypeFilter] = useState('');
  const [echoPlaceFilter, setEchoPlaceFilter] = useState('');
  const [echoRarityFilter, setEchoRarityFilter] = useState(0);
  const [echoMinOwned, setEchoMinOwned] = useState(0);
  const [echoFavoritesOnly, setEchoFavoritesOnly] = useState(false);
  const [echoSort, setEchoSort] = useState<'api' | 'name' | 'cost'>('api');
  const [echoLibrarySource, setEchoLibrarySource] = useState<EchoLibrarySource>('catalog');
  const [signatureWeaponID, setSignatureWeaponID] = useState<number>();
  const [saving, setSaving] = useState(false);
  const [deletedID, setDeletedID] = useState<number>();
  const [exportOpen, setExportOpen] = useState(false);

  async function load(selectID?: number) {
    try {
      const [nextBuilds, nextCharacters, nextWeapons, nextEchoes, nextOwnedEchoes, nextSonatas] =
        await Promise.all([
          listBuilds(),
          listCharacters(characterFilter),
          listWeapons(weaponFilter),
          listEchoes(echoFilter),
          listOwnedEchoes(),
          listSonatas(),
        ]);
      setBuilds(nextBuilds);
      setCharacters(nextCharacters);
      setWeapons(nextWeapons);
      setEchoes(nextEchoes);
      setOwnedEchoes(nextOwnedEchoes);
      setSonatas(nextSonatas);
      if (selectID) {
        const selected = nextBuilds.find((item) => item.id === selectID);
        if (selected) setDraft(normalizeBuild(selected));
      }
      onError('');
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  useEffect(() => {
    const target = readOpenTarget('build');
    void load(target?.id);
  }, []);
  useEffect(() => {
    let cancelled = false;
    if (!draft.characterId) {
      setSignatureWeaponID(undefined);
      return;
    }
    void getCharacter(draft.characterId)
      .then((profile) => {
        if (!cancelled) setSignatureWeaponID(profile.signatureWeapon?.id);
      })
      .catch(() => {
        if (!cancelled) setSignatureWeaponID(undefined);
      });
    return () => {
      cancelled = true;
    };
  }, [draft.characterId]);

  const character = characters.find((item) => item.id === draft.characterId);
  const weapon = weapons.find((item) => item.id === draft.weaponId);
  const roverSelected = isRoverCharacter(character);
  const totalCost = draft.echoes.reduce((sum, echo) => sum + echo.cost, 0);
  const activeEcho = target.kind === 'echo' ? draft.echoes[target.slot] : undefined;
  const canSave = Boolean(draft.name.trim() && draft.characterId && totalCost <= 12 && !saving);

  const library = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase('pt-BR');
    if (target.kind === 'character') {
      return characters
        .filter(
          (item) =>
            (!needle || item.name.toLocaleLowerCase('pt-BR').includes(needle)) &&
            (!elementFilter || item.elementCode === elementFilter) &&
            (!rarityFilter || item.rarity === rarityFilter) &&
            (ownershipFilter === 'all' || item.owned === (ownershipFilter === 'owned')) &&
            (!favoritesOnly || item.favorite)
        )
        .sort((left, right) => {
          if (characterSort === 'name') return left.name.localeCompare(right.name);
          if (characterSort === 'rarity')
            return right.rarity - left.rarity || left.apiOrder - right.apiOrder;
          if (characterSort === 'level')
            return right.level - left.level || left.apiOrder - right.apiOrder;
          return left.apiOrder - right.apiOrder;
        });
    }
    if (target.kind === 'weapon') {
      return weapons
        .filter(
          (item) =>
            (!character || item.typeCode === character.weaponTypeCode) &&
            (!needle || item.name.toLocaleLowerCase('pt-BR').includes(needle)) &&
            (!weaponRarityFilter || item.rarity === weaponRarityFilter) &&
            (!weaponSubStatFilter || item.subStat === weaponSubStatFilter) &&
            (weaponOwnershipFilter === 'all' ||
              item.owned === (weaponOwnershipFilter === 'owned')) &&
            (!weaponMinAtk || item.baseAtk >= weaponMinAtk) &&
            (!weaponMaxAtk || item.baseAtk <= weaponMaxAtk) &&
            (!weaponFavoritesOnly || item.favorite)
        )
        .sort((left, right) => {
          if (left.id === signatureWeaponID) return -1;
          if (right.id === signatureWeaponID) return 1;
          if (weaponSort === 'name') return left.name.localeCompare(right.name);
          if (weaponSort === 'atk')
            return right.baseAtk - left.baseAtk || right.rarity - left.rarity;
          if (weaponSort === 'api') return left.id - right.id;
          return right.rarity - left.rarity || right.baseAtk - left.baseAtk;
        });
    }
    return echoes
      .filter(
        (item) =>
          (!costFilter || item.cost === costFilter) &&
          (!sonataFilter || parseIDs(item.sonataIdsJson).includes(sonataFilter)) &&
          (!echoClassFilter || item.class === echoClassFilter) &&
          (!echoTypeFilter || item.type === echoTypeFilter) &&
          (!echoPlaceFilter || item.place === echoPlaceFilter) &&
          (!echoRarityFilter || parseIDs(item.raritiesJson).includes(echoRarityFilter)) &&
          (!echoMinOwned || item.ownedCount >= echoMinOwned) &&
          (!echoFavoritesOnly || item.favorite) &&
          (!needle || item.name.toLocaleLowerCase('pt-BR').includes(needle))
      )
      .sort((left, right) => {
        if (echoSort === 'name') return left.name.localeCompare(right.name);
        if (echoSort === 'cost') return right.cost - left.cost || left.id - right.id;
        return left.id - right.id;
      });
  }, [
    target,
    query,
    costFilter,
    characters,
    weapons,
    echoes,
    character,
    elementFilter,
    rarityFilter,
    ownershipFilter,
    favoritesOnly,
    characterSort,
    weaponRarityFilter,
    weaponSubStatFilter,
    weaponOwnershipFilter,
    weaponMinAtk,
    weaponMaxAtk,
    weaponFavoritesOnly,
    weaponSort,
    signatureWeaponID,
    sonataFilter,
    echoClassFilter,
    echoTypeFilter,
    echoPlaceFilter,
    echoRarityFilter,
    echoMinOwned,
    echoFavoritesOnly,
    echoSort,
  ]);
  const ownedEchoLibrary = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase('pt-BR');
    return ownedEchoes
      .filter(
        (item) =>
          (!needle ||
            `${item.echoName} ${item.mainStat} ${item.sonataName} ${item.note}`
              .toLocaleLowerCase('pt-BR')
              .includes(needle)) &&
          (!costFilter || item.cost === costFilter) &&
          (!sonataFilter || item.sonataId === sonataFilter) &&
          (!echoFavoritesOnly || item.favorite)
      )
      .sort((left, right) => {
        if (echoSort === 'name') return left.echoName.localeCompare(right.echoName);
        if (echoSort === 'cost') return right.cost - left.cost || right.id - left.id;
        return right.id - left.id;
      });
  }, [costFilter, echoFavoritesOnly, echoSort, ownedEchoes, query, sonataFilter]);
  const activeLibrary =
    target.kind === 'echo' && echoLibrarySource === 'inventory' ? ownedEchoLibrary : library;
  const activeLibraryTitle =
    target.kind === 'echo' && echoLibrarySource === 'inventory'
      ? 'Echoes do inventário'
      : libraryTitle(target);
  const sonataProgress = useMemo(
    () => buildSonataProgress(draft.echoes, sonatas),
    [draft.echoes, sonatas]
  );
  const visibleBuilds = builds.filter(
    (item) =>
      !sourceQuery.trim() ||
      item.name.toLocaleLowerCase('pt-BR').includes(sourceQuery.trim().toLocaleLowerCase('pt-BR'))
  );
  const echoClasses = [...new Set(echoes.map((item) => item.class).filter(Boolean))].sort();
  const echoTypes = [...new Set(echoes.map((item) => item.type).filter(Boolean))].sort();
  const echoPlaces = [...new Set(echoes.map((item) => item.place).filter(Boolean))].sort();
  const weaponSubStats = [...new Set(weapons.map((item) => item.subStat).filter(Boolean))].sort();

  const libraryFiltersActive = Boolean(
    query ||
    (target.kind === 'character' &&
      (elementFilter ||
        rarityFilter ||
        ownershipFilter !== 'all' ||
        favoritesOnly ||
        characterSort !== 'api')) ||
    (target.kind === 'weapon' &&
      (weaponRarityFilter ||
        weaponSubStatFilter ||
        weaponOwnershipFilter !== 'all' ||
        weaponFavoritesOnly ||
        weaponSort !== 'rarity' ||
        weaponMinAtk ||
        weaponMaxAtk)) ||
    (target.kind === 'echo' &&
      (costFilter ||
        sonataFilter ||
        (echoLibrarySource === 'catalog' &&
          (echoClassFilter ||
            echoTypeFilter ||
            echoPlaceFilter ||
            echoRarityFilter ||
            echoMinOwned)) ||
        echoFavoritesOnly ||
        echoSort !== 'api'))
  );

  function resetLibraryFilters() {
    setQuery('');
    if (target.kind === 'character') {
      setElementFilter(0);
      setRarityFilter(0);
      setOwnershipFilter('all');
      setFavoritesOnly(false);
      setCharacterSort('api');
    } else if (target.kind === 'weapon') {
      setWeaponRarityFilter(0);
      setWeaponSubStatFilter('');
      setWeaponOwnershipFilter('all');
      setWeaponFavoritesOnly(false);
      setWeaponSort('rarity');
      setWeaponMinAtk(0);
      setWeaponMaxAtk(0);
    } else {
      setCostFilter(0);
      setSonataFilter(0);
      setEchoClassFilter('');
      setEchoTypeFilter('');
      setEchoPlaceFilter('');
      setEchoRarityFilter(0);
      setEchoMinOwned(0);
      setEchoFavoritesOnly(false);
      setEchoSort('api');
    }
  }

  function changeEchoLibrarySource(source: EchoLibrarySource) {
    setEchoLibrarySource(source);
    setQuery('');
    setEchoClassFilter('');
    setEchoTypeFilter('');
    setEchoPlaceFilter('');
    setEchoRarityFilter(0);
    setEchoMinOwned(0);
  }

  function newBuild() {
    setDraft(emptyBuild(version));
    setTarget({ kind: 'character' });
    setQuery('');
  }

  function openBuild(build: Build) {
    setDraft(normalizeBuild(build));
    setTarget({ kind: 'character' });
    setQuery('');
  }

  function selectCharacter(item: Character) {
    const incompatible = weapon && weapon.typeCode !== item.weaponTypeCode;
    setDraft({
      ...draft,
      characterId: item.id,
      characterName: item.name,
      characterIcon: item.iconPath,
      characterLevel: item.owned ? item.level : draft.characterLevel,
      sequence: item.owned ? item.sequence : draft.sequence,
      weaponId: incompatible ? undefined : draft.weaponId,
      weaponName: incompatible ? '' : draft.weaponName,
      weaponIcon: incompatible ? '' : draft.weaponIcon,
    });
    setTarget({ kind: 'weapon' });
    setQuery('');
  }

  function selectWeapon(item: Weapon) {
    setDraft({
      ...draft,
      weaponId: item.id,
      weaponName: item.name,
      weaponIcon: item.iconPath,
      weaponLevel: item.owned ? item.level : draft.weaponLevel,
      weaponRank: item.owned ? item.rank : draft.weaponRank,
    });
    setTarget({ kind: 'echo', slot: firstEmptySlot(draft.echoes) });
    setQuery('');
  }

  function selectEcho(item: Echo) {
    if (target.kind !== 'echo') return;
    const previous = draft.echoes[target.slot];
    const nextCost = totalCost - (previous?.cost ?? 0) + item.cost;
    if (nextCost > 12) {
      onError(`Este Echo ultrapassa o limite: ${nextCost}/12 de custo.`);
      return;
    }
    const effectiveSlot = previous ? target.slot : Math.min(target.slot, draft.echoes.length);
    const next = [...draft.echoes];
    next[effectiveSlot] = ownedFromCatalog(item, sonatas);
    setDraft({ ...draft, echoes: compactSlots(next) });
    setTarget({ kind: 'echo', slot: effectiveSlot });
    onError('');
  }

  function selectOwnedEcho(item: OwnedEcho) {
    if (target.kind !== 'echo') return;
    const usedInSlot = draft.echoes.findIndex(
      (echo, index) => index !== target.slot && echo.id === item.id
    );
    if (usedInSlot >= 0) {
      onError(`Esta peça já está no slot ${usedInSlot + 1} da build.`);
      return;
    }
    const previous = draft.echoes[target.slot];
    const nextCost = totalCost - (previous?.cost ?? 0) + item.cost;
    if (nextCost > 12) {
      onError(`Este Echo ultrapassa o limite: ${nextCost}/12 de custo.`);
      return;
    }
    const effectiveSlot = previous ? target.slot : Math.min(target.slot, draft.echoes.length);
    const next = [...draft.echoes];
    next[effectiveSlot] = { ...item };
    setDraft({ ...draft, echoes: compactSlots(next) });
    setTarget({ kind: 'echo', slot: effectiveSlot });
    onError('');
  }

  function updateEcho(patch: Partial<OwnedEcho>) {
    if (target.kind !== 'echo' || !activeEcho) return;
    const next = [...draft.echoes];
    next[target.slot] = { ...activeEcho, ...patch };
    setDraft({ ...draft, echoes: next });
  }

  function removeEcho(slot: number) {
    const next = [...draft.echoes];
    next.splice(slot, 1);
    setDraft({ ...draft, echoes: next });
    setTarget({ kind: 'echo', slot: Math.min(slot, 4) });
  }

  async function submit(): Promise<boolean> {
    if (!canSave) return false;
    setSaving(true);
    try {
      const savedEchoes: OwnedEcho[] = [];
      for (const echo of draft.echoes) {
        savedEchoes.push(await saveOwnedEcho({ ...echo, characterId: draft.characterId }));
      }
      const saved = await saveBuild({ ...draft, echoes: savedEchoes, rotationId: undefined });
      await load(saved.id);
      onError('');
      return true;
    } catch (cause) {
      onError(messageFrom(cause));
      return false;
    } finally {
      setSaving(false);
    }
  }

  async function openExport() {
    if (await submit()) setExportOpen(true);
  }

  async function duplicate(id: number) {
    try {
      const copy = await duplicateBuild(id);
      await load(copy.id);
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  async function remove(id: number) {
    try {
      await deleteBuild(id);
      setDeletedID(id);
      newBuild();
      await load();
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  async function undo() {
    if (!deletedID) return;
    try {
      await restoreBuild(deletedID);
      setDeletedID(undefined);
      await load(deletedID);
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  const shortcutFeedback = useContextualShortcuts({
    canSave,
    onNew: newBuild,
    onSave: submit,
    newMessage: 'Nova build criada.',
    savedMessage: 'Build salva.',
    invalidMessage: 'Selecione um personagem e respeite o limite de custo antes de salvar.',
  });

  return (
    <div className="buildWorkspace">
      <aside className="buildSources">
        <div className="buildSourceHeading">
          <span className="sectionLabel">ARQUIVO DE BUILDS</span>
          <strong>{builds.length} salvas</strong>
        </div>
        <button className="newTeamButton buildNewTop" onClick={newBuild}>
          <Plus size={16} />
          Nova build
        </button>
        <label className="sourceSearch">
          <Search size={15} />
          <input
            value={sourceQuery}
            onChange={(event) => setSourceQuery(event.target.value)}
            placeholder="Buscar builds..."
          />
        </label>
        <div className="sourceList">
          {visibleBuilds.length ? (
            visibleBuilds.map((build) => (
              <article
                className={`sourceRow ${draft.id === build.id ? 'active' : ''}`}
                key={build.id}
              >
                <button className="sourceRowMain" onClick={() => openBuild(build)}>
                  <span className="sourceRowTitle">
                    {build.favorite && <Star size={11} />} {build.name}
                  </span>
                  <small>
                    {build.characterName || 'Sem personagem'} · {build.echoes.length} Echoes
                  </small>
                  <span className="sourceAvatars">
                    <Portrait path={build.characterIcon} name={build.characterName} />
                    {build.echoes.slice(0, 4).map((echo) => (
                      <Portrait
                        path={echo.iconPath}
                        name={echo.echoName}
                        key={echo.id || echo.echoId}
                      />
                    ))}
                  </span>
                </button>
                <div className="sourceRowActions">
                  <button title="Duplicar" onClick={() => void duplicate(build.id)}>
                    <Copy size={13} />
                  </button>
                  <button title="Excluir" onClick={() => void remove(build.id)}>
                    <Trash2 size={13} />
                  </button>
                </div>
              </article>
            ))
          ) : (
            <div className="sourceEmpty">
              <strong>{builds.length ? 'Nenhuma build encontrada' : 'Nenhuma build salva'}</strong>
              <small>{builds.length ? 'Tente outro nome.' : 'Monte a primeira composição.'}</small>
            </div>
          )}
        </div>
      </aside>

      <main className="buildComposer">
        <header className="composerHeader">
          <div>
            <span className="sectionLabel">COMPOSIÇÃO DA BUILD</span>
            <input
              value={draft.name}
              onChange={(event) => setDraft({ ...draft, name: event.target.value })}
              aria-label="Nome da build"
            />
          </div>
          <div className={`echoBudget ${totalCost > 12 ? 'over' : ''}`}>
            <span>CUSTO DOS ECHOES</span>
            <strong>
              {totalCost}
              <small>/12</small>
            </strong>
          </div>
        </header>

        <section className="buildCore">
          <button
            className={`buildCoreCard ${target.kind === 'character' ? 'active' : ''}`}
            onClick={() => {
              setTarget({ kind: 'character' });
              setQuery('');
            }}
          >
            <span>PERSONAGEM</span>
            {character ? (
              <>
                <Portrait path={character.iconPath} name={character.name} large />
                <div>
                  <strong>{character.name}</strong>
                  <small>
                    {character.element} · {character.weaponType}
                  </small>
                  <em>
                    Nv. {draft.characterLevel} · S{draft.sequence}
                  </em>
                </div>
              </>
            ) : (
              <>
                <UserRound size={34} />
                <div>
                  <strong>Selecionar personagem</strong>
                  <small>Escolha na biblioteca abaixo</small>
                </div>
              </>
            )}
            <ChevronRight size={17} />
          </button>
          <button
            className={`buildCoreCard ${target.kind === 'weapon' ? 'active' : ''}`}
            onClick={() => {
              setTarget({ kind: 'weapon' });
              setQuery('');
            }}
          >
            <span>ARMA</span>
            {weapon ? (
              <>
                <Portrait path={weapon.iconPath} name={weapon.name} large />
                <div>
                  <strong>{weapon.name}</strong>
                  <small>
                    {weapon.type} · ATK {weapon.baseAtk}
                  </small>
                  <em>
                    Nv. {draft.weaponLevel} · R{draft.weaponRank}
                  </em>
                </div>
              </>
            ) : (
              <>
                <Swords size={34} />
                <div>
                  <strong>Selecionar arma</strong>
                  <small>
                    {character
                      ? `Somente ${character.weaponType}`
                      : 'Escolha um personagem primeiro'}
                  </small>
                </div>
              </>
            )}
            <ChevronRight size={17} />
          </button>
        </section>

        <section className="buildSkillEditor" aria-labelledby="build-skills-title">
          <header>
            <div>
              <span className="sectionLabel">NÍVEIS DE HABILIDADE</span>
              <h2 id="build-skills-title">Progressão da build</h2>
            </div>
            <small>Use 0 quando a habilidade ainda não foi informada.</small>
          </header>
          <div>
            {buildSkillFields.map(({ key, label }) => (
              <label key={key}>
                <span>{label}</span>
                <input
                  type="number"
                  min={0}
                  max={10}
                  value={draft[key]}
                  onChange={(event) =>
                    setDraft({ ...draft, [key]: clamp(Number(event.target.value), 0, 10) })
                  }
                />
              </label>
            ))}
          </div>
        </section>

        <section className="echoLoadout">
          <header>
            <div>
              <span className="sectionLabel">FORMAÇÃO DE ECHOES</span>
              <h2>Slots da build</h2>
            </div>
            <p>Escolha até 5 Echoes respeitando o custo máximo de 12.</p>
          </header>
          <div className="echoSlotsStage">
            {Array.from({ length: 5 }, (_, slot) => {
              const echo = draft.echoes[slot];
              const active = target.kind === 'echo' && target.slot === slot;
              return (
                <button
                  className={`echoBuildSlot ${active ? 'active' : ''} ${echo ? 'filled' : ''}`}
                  onClick={() => {
                    setTarget({ kind: 'echo', slot });
                    setQuery('');
                  }}
                  key={slot}
                >
                  <span className="echoSlotNumber">{String(slot + 1).padStart(2, '0')}</span>
                  {echo ? (
                    <>
                      <div className="echoSlotArt">
                        <Asset path={echo.iconPath} name={echo.echoName} />
                      </div>
                      <div className="echoSlotInfo">
                        <strong>{echo.echoName}</strong>
                        <small>
                          Custo {echo.cost} · +{echo.level}
                        </small>
                        <em>{echo.mainStat || 'Atributo principal pendente'}</em>
                      </div>
                    </>
                  ) : (
                    <div className="emptyEchoSlot">
                      <Plus size={22} />
                      <strong>Adicionar Echo</strong>
                      <small>Slot disponível</small>
                    </div>
                  )}
                  {active && (
                    <span className="selectedMark">
                      <Check size={12} />
                      Selecionado
                    </span>
                  )}
                </button>
              );
            })}
          </div>
          <SonataSummary items={sonataProgress} />
        </section>

        <footer className="composerActions buildComposerActions">
          <div className="buildFlags">
            <label>
              <input
                type="checkbox"
                checked={draft.favorite}
                onChange={(event) => setDraft({ ...draft, favorite: event.target.checked })}
              />
              <Star size={13} />
              Favorita
            </label>
            <label>
              <input
                type="checkbox"
                checked={draft.locked}
                onChange={(event) => setDraft({ ...draft, locked: event.target.checked })}
              />
              <Lock size={13} />
              Bloqueada
            </label>
          </div>
          <button onClick={newBuild}>
            <X size={15} />
            Limpar
          </button>
          <button
            disabled={!canSave}
            onClick={() => void openExport()}
            title="Salvar os dados atuais e exportar o card em PNG"
          >
            <ImageDown size={15} />
            Salvar e exportar
          </button>
          <button className="saveTeamButton" disabled={!canSave} onClick={() => void submit()}>
            <Save size={15} />
            {saving ? 'Salvando...' : 'Salvar build'}
          </button>
        </footer>
      </main>

      <aside className="buildInspector">
        {target.kind === 'character' ? (
          <CharacterInspector character={character} draft={draft} setDraft={setDraft} />
        ) : target.kind === 'weapon' ? (
          <WeaponInspector weapon={weapon} draft={draft} setDraft={setDraft} />
        ) : activeEcho ? (
          <EchoInspector
            echo={activeEcho}
            sonatas={sonatas}
            onChange={updateEcho}
            onRemove={() => removeEcho(target.slot)}
          />
        ) : (
          <div className="inspectorEmpty">
            <Waves size={30} />
            <strong>Slot de Echo vazio</strong>
            <p>Escolha uma peça na biblioteca para configurar seus atributos.</p>
          </div>
        )}
      </aside>

      <section className="buildLibrary">
        <LibraryFilterBar
          className="buildCatalogFilters"
          contentClassName="buildFilterGrid"
          title={activeLibraryTitle}
          resultLabel={`${activeLibrary.length} resultados`}
          query={query}
          placeholder={`Buscar ${activeLibraryTitle.toLocaleLowerCase('pt-BR')}...`}
          sortValue={
            target.kind === 'character'
              ? characterSort
              : target.kind === 'weapon'
                ? weaponSort
                : echoSort
          }
          sortLabel={
            target.kind === 'character'
              ? 'Ordenar personagens'
              : target.kind === 'weapon'
                ? 'Ordenar armas'
                : 'Ordenar Echoes'
          }
          sortOptions={
            target.kind === 'character'
              ? [
                  { value: 'api', label: 'Lançamento' },
                  { value: 'name', label: 'Nome A–Z' },
                  { value: 'rarity', label: 'Maior raridade' },
                  { value: 'level', label: 'Nível na conta' },
                ]
              : target.kind === 'weapon'
                ? [
                    { value: 'rarity', label: 'Raridade' },
                    { value: 'atk', label: 'ATK base' },
                    { value: 'name', label: 'Nome A–Z' },
                    { value: 'api', label: 'ID da API' },
                  ]
                : [
                    {
                      value: 'api',
                      label: echoLibrarySource === 'inventory' ? 'Mais recentes' : 'ID da API',
                    },
                    { value: 'name', label: 'Nome A–Z' },
                    { value: 'cost', label: 'Maior custo' },
                  ]
          }
          active={libraryFiltersActive}
          onQueryChange={setQuery}
          onSortChange={(value) => {
            if (target.kind === 'character') {
              setCharacterSort(value as typeof characterSort);
            } else if (target.kind === 'weapon') {
              setWeaponSort(value as typeof weaponSort);
            } else {
              setEchoSort(value as typeof echoSort);
            }
          }}
          onReset={resetLibraryFilters}
        >
          {target.kind === 'character' && (
            <>
              <div className="catalogFacet catalogFacetWide">
                <span>Atributo</span>
                <div className="catalogChipRail">
                  {(
                    [
                      [0, 'Todos'],
                      [1, 'Glacio'],
                      [2, 'Fusion'],
                      [3, 'Electro'],
                      [4, 'Aero'],
                      [5, 'Spectro'],
                      [6, 'Havoc'],
                    ] as const
                  ).map(([id, name]) => (
                    <button
                      type="button"
                      className={
                        elementFilter === id
                          ? `catalogChip active element-${id}`
                          : `catalogChip element-${id}`
                      }
                      onClick={() => setElementFilter(id)}
                      key={id}
                    >
                      {id > 0 && <i />}
                      {name}
                    </button>
                  ))}
                </div>
              </div>
              <div className="catalogFacet">
                <span>Raridade</span>
                <div className="catalogChipRail">
                  {[
                    { value: 0, label: 'Todas' },
                    { value: 5, label: '5★' },
                    { value: 4, label: '4★' },
                  ].map((item) => (
                    <button
                      type="button"
                      className={rarityFilter === item.value ? 'catalogChip active' : 'catalogChip'}
                      onClick={() => setRarityFilter(item.value)}
                      key={item.value}
                    >
                      {item.label}
                    </button>
                  ))}
                </div>
              </div>
              <div className="catalogFacet">
                <span>Conta</span>
                <div className="catalogChipRail">
                  {(
                    [
                      ['all', 'Todos'],
                      ['owned', 'Possuídos'],
                      ['missing', 'Não possuídos'],
                    ] as const
                  ).map(([value, label]) => (
                    <button
                      type="button"
                      className={ownershipFilter === value ? 'catalogChip active' : 'catalogChip'}
                      onClick={() => setOwnershipFilter(value)}
                      key={value}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
              <button
                type="button"
                className={favoritesOnly ? 'catalogToggle active' : 'catalogToggle'}
                onClick={() => setFavoritesOnly((current) => !current)}
                aria-pressed={favoritesOnly}
              >
                <Star size={14} fill={favoritesOnly ? 'currentColor' : 'none'} />
                Somente favoritos
              </button>
            </>
          )}
          {target.kind === 'weapon' && (
            <>
              <div className="catalogFacet">
                <span>Raridade</span>
                <div className="catalogChipRail">
                  {[
                    { value: 0, label: 'Todas' },
                    { value: 5, label: '5★' },
                    { value: 4, label: '4★' },
                    { value: 3, label: '3★' },
                  ].map((item) => (
                    <button
                      type="button"
                      className={
                        weaponRarityFilter === item.value ? 'catalogChip active' : 'catalogChip'
                      }
                      onClick={() => setWeaponRarityFilter(item.value)}
                      key={item.value}
                    >
                      {item.label}
                    </button>
                  ))}
                </div>
              </div>
              <FilterField label="Subatributo">
                <select
                  value={weaponSubStatFilter}
                  onChange={(event) => setWeaponSubStatFilter(event.target.value)}
                >
                  <option value="">Todos</option>
                  {weaponSubStats.map((item) => (
                    <option value={item} key={item}>
                      {item}
                    </option>
                  ))}
                </select>
              </FilterField>
              <div className="catalogFacet">
                <span>Conta</span>
                <div className="catalogChipRail">
                  {(
                    [
                      ['all', 'Todas'],
                      ['owned', 'Possuídas'],
                      ['missing', 'Não possuídas'],
                    ] as const
                  ).map(([value, label]) => (
                    <button
                      type="button"
                      className={
                        weaponOwnershipFilter === value ? 'catalogChip active' : 'catalogChip'
                      }
                      onClick={() => setWeaponOwnershipFilter(value)}
                      key={value}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
              <FilterRange
                label="ATK base"
                min={1}
                max={700}
                minValue={weaponMinAtk}
                maxValue={weaponMaxAtk}
                onMinChange={setWeaponMinAtk}
                onMaxChange={setWeaponMaxAtk}
              />
              <span className="signaturePriority">
                <Star size={13} />
                {roverSelected
                  ? 'Arma recomendada sempre primeiro'
                  : 'Arma assinatura sempre primeiro'}
              </span>
              <button
                type="button"
                className={weaponFavoritesOnly ? 'catalogToggle active' : 'catalogToggle'}
                onClick={() => setWeaponFavoritesOnly((current) => !current)}
                aria-pressed={weaponFavoritesOnly}
              >
                <Star size={14} fill={weaponFavoritesOnly ? 'currentColor' : 'none'} />
                Somente favoritas
              </button>
            </>
          )}
          {target.kind === 'echo' && (
            <>
              <div className="catalogFacet">
                <span>Origem</span>
                <div className="catalogChipRail">
                  {(
                    [
                      ['catalog', 'Catálogo'],
                      ['inventory', 'Inventário'],
                    ] as const
                  ).map(([value, label]) => (
                    <button
                      type="button"
                      className={echoLibrarySource === value ? 'catalogChip active' : 'catalogChip'}
                      onClick={() => changeEchoLibrarySource(value)}
                      key={value}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
              <div className="catalogFacet">
                <span>Custo</span>
                <div className="catalogChipRail">
                  {[
                    { value: 0, label: 'Todos' },
                    { value: 1, label: 'Custo 1' },
                    { value: 3, label: 'Custo 3' },
                    { value: 4, label: 'Custo 4' },
                  ].map((item) => (
                    <button
                      type="button"
                      className={costFilter === item.value ? 'catalogChip active' : 'catalogChip'}
                      onClick={() => setCostFilter(item.value)}
                      key={item.value}
                    >
                      {item.label}
                    </button>
                  ))}
                </div>
              </div>
              <div className="sonataFilterGroup">
                <span>Sonata</span>
                <div className="sonataChipRail" role="group" aria-label="Filtrar por Sonata">
                  <button
                    className={sonataFilter === 0 ? 'active' : ''}
                    onClick={() => setSonataFilter(0)}
                  >
                    Todas
                  </button>
                  {sonatas.map((item) => (
                    <button
                      className={sonataFilter === item.id ? 'active' : ''}
                      onClick={() => setSonataFilter(item.id)}
                      title={item.name}
                      key={item.id}
                    >
                      {item.iconPath?.startsWith('/cache/') && <img src={item.iconPath} alt="" />}
                      <span>{item.name}</span>
                    </button>
                  ))}
                </div>
              </div>
              {echoLibrarySource === 'catalog' && (
                <>
                  <FilterField label="Classe">
                    <select
                      value={echoClassFilter}
                      onChange={(event) => setEchoClassFilter(event.target.value)}
                    >
                      <option value="">Todas</option>
                      {echoClasses.map((item) => (
                        <option value={item} key={item}>
                          {item}
                        </option>
                      ))}
                    </select>
                  </FilterField>
                  <FilterField label="Tipo">
                    <select
                      value={echoTypeFilter}
                      onChange={(event) => setEchoTypeFilter(event.target.value)}
                    >
                      <option value="">Todos</option>
                      {echoTypes.map((item) => (
                        <option value={item} key={item}>
                          {item}
                        </option>
                      ))}
                    </select>
                  </FilterField>
                  <FilterField label="Local">
                    <select
                      value={echoPlaceFilter}
                      onChange={(event) => setEchoPlaceFilter(event.target.value)}
                    >
                      <option value="">Todos</option>
                      {echoPlaces.map((item) => (
                        <option value={item} key={item}>
                          {item}
                        </option>
                      ))}
                    </select>
                  </FilterField>
                  <FilterField label="Raridade">
                    <select
                      value={echoRarityFilter}
                      onChange={(event) => setEchoRarityFilter(Number(event.target.value))}
                    >
                      <option value={0}>Todas</option>
                      <option value={5}>5 estrelas</option>
                      <option value={4}>4 estrelas</option>
                      <option value={3}>3 estrelas</option>
                      <option value={2}>2 estrelas</option>
                    </select>
                  </FilterField>
                  <FilterField label="Quantidade mínima">
                    <input
                      type="number"
                      min={0}
                      value={echoMinOwned || ''}
                      placeholder="0"
                      onChange={(event) => setEchoMinOwned(Number(event.target.value))}
                    />
                  </FilterField>
                </>
              )}
              <button
                type="button"
                className={echoFavoritesOnly ? 'catalogToggle active' : 'catalogToggle'}
                onClick={() => setEchoFavoritesOnly((current) => !current)}
                aria-pressed={echoFavoritesOnly}
              >
                <Star size={14} fill={echoFavoritesOnly ? 'currentColor' : 'none'} />
                Somente favoritos
              </button>
            </>
          )}
        </LibraryFilterBar>
        <div className="buildLibraryGrid">
          {target.kind === 'character' &&
            (library as Character[]).map((item) => (
              <LibraryCard
                key={item.id}
                name={item.name}
                path={item.iconPath}
                meta={`${item.element} · ${item.weaponType}`}
                selected={item.id === draft.characterId}
                onClick={() => selectCharacter(item)}
              />
            ))}
          {target.kind === 'weapon' &&
            (library as Weapon[]).map((item) => (
              <LibraryCard
                key={item.id}
                name={item.name}
                path={item.iconPath}
                meta={`${item.type} · ${item.rarity}★ · ATK ${item.baseAtk}`}
                badge={
                  item.id === signatureWeaponID
                    ? roverSelected
                      ? 'RECOMENDADA'
                      : 'ASSINATURA'
                    : undefined
                }
                selected={item.id === draft.weaponId}
                disabled={!character}
                onClick={() => selectWeapon(item)}
              />
            ))}
          {target.kind === 'echo' &&
            echoLibrarySource === 'catalog' &&
            (library as Echo[]).map((item) => {
              const current = activeEcho;
              const unavailable = totalCost - (current?.cost ?? 0) + item.cost > 12;
              return (
                <LibraryCard
                  key={item.id}
                  name={item.name}
                  path={item.iconPath}
                  meta={`Custo ${item.cost} · ${item.class}`}
                  selected={current?.echoId === item.id}
                  disabled={unavailable}
                  onClick={() => selectEcho(item)}
                />
              );
            })}
          {target.kind === 'echo' &&
            echoLibrarySource === 'inventory' &&
            ownedEchoLibrary.map((item) => {
              const current = activeEcho;
              const unavailable = totalCost - (current?.cost ?? 0) + item.cost > 12;
              const usedElsewhere = draft.echoes.some(
                (echo, index) => index !== target.slot && echo.id === item.id
              );
              const otherBuild = builds.find(
                (build) => build.id !== draft.id && build.echoes.some((echo) => echo.id === item.id)
              );
              return (
                <LibraryCard
                  key={item.id}
                  name={item.echoName}
                  path={item.iconPath}
                  meta={`Custo ${item.cost} · +${item.level} · ${item.mainStat || 'Sem atributo principal'}`}
                  badge={
                    usedElsewhere
                      ? 'JÁ NA BUILD'
                      : otherBuild
                        ? `EM ${otherBuild.name}`
                        : item.sonataName || 'INVENTÁRIO'
                  }
                  selected={current?.id === item.id}
                  disabled={unavailable || usedElsewhere}
                  onClick={() => selectOwnedEcho(item)}
                />
              );
            })}
          {target.kind === 'echo' && activeLibrary.length === 0 && (
            <div className="buildLibraryEmpty">
              <Waves size={24} />
              <strong>
                {echoLibrarySource === 'inventory'
                  ? 'Nenhum Echo no inventário'
                  : 'Nenhum Echo encontrado'}
              </strong>
              <small>
                {echoLibrarySource === 'inventory'
                  ? 'Cadastre peças na aba Echoes para reutilizá-las nas Builds.'
                  : 'Ajuste os filtros para voltar ao catálogo.'}
              </small>
            </div>
          )}
        </div>
      </section>

      {shortcutFeedback && (
        <div className={`shortcutToast ${shortcutFeedback.tone}`} role="status" aria-live="polite">
          <kbd>{shortcutFeedback.message.includes('Nova') ? 'Ctrl+N' : 'Ctrl+S'}</kbd>
          {shortcutFeedback.message}
        </div>
      )}

      {deletedID && (
        <div className="undoToast">
          Build excluída.<button onClick={() => void undo()}>Desfazer</button>
          <button onClick={() => setDeletedID(undefined)} aria-label="Fechar">
            ×
          </button>
        </div>
      )}
      {exportOpen && draft.id > 0 && (
        <BuildExportCard
          build={draft}
          character={character}
          sonatas={sonatas}
          onClose={() => setExportOpen(false)}
          onError={onError}
        />
      )}
    </div>
  );
}

function SonataSummary({ items }: { items: SonataProgressItem[] }) {
  return (
    <section className="sonataSummary" aria-labelledby="sonata-summary-title">
      <header>
        <div>
          <span className="sectionLabel">EFEITOS DE SONATA</span>
          <h3 id="sonata-summary-title">Conjuntos da build</h3>
        </div>
        <small>{items.length ? `${items.length} em uso` : 'Nenhum conjunto'}</small>
      </header>
      {items.length ? (
        <div className="sonataSummaryGrid">
          {items.map(({ sonata, count, tiers, hasActiveEffect, piecesWithoutActiveEffect }) => {
            const nextTier = tiers.find((tier) => count < tier.pieces);
            return (
              <article className="sonataSummaryCard" key={sonata.id}>
                <div className="sonataSummaryIdentity">
                  {sonata.iconPath?.startsWith('/cache/') ? (
                    <img src={sonata.iconPath} alt="" />
                  ) : (
                    <Waves size={22} aria-hidden="true" />
                  )}
                  <div>
                    <strong>{sonata.name}</strong>
                    <small>
                      {count} {count === 1 ? 'peça equipada' : 'peças equipadas'}
                    </small>
                  </div>
                  <span className={nextTier ? 'incomplete' : 'active'}>
                    {nextTier ? `Faltam ${nextTier.pieces - count}` : 'Completo'}
                  </span>
                </div>
                <div className="sonataTierList">
                  {tiers.length ? (
                    tiers.map((tier) => (
                      <div
                        className={tier.active ? 'sonataTier active' : 'sonataTier'}
                        key={tier.pieces}
                      >
                        <span>{tier.pieces} peças</span>
                        <strong>{tier.active ? 'Ativo' : `${tier.missing} restantes`}</strong>
                        <p>{tier.description}</p>
                      </div>
                    ))
                  ) : (
                    <p className="sonataDescriptionUnavailable">
                      A fonte sincronizada não forneceu uma descrição para este conjunto.
                    </p>
                  )}
                </div>
                {!hasActiveEffect && piecesWithoutActiveEffect > 0 && (
                  <p className="sonataNoActiveEffect">
                    {piecesWithoutActiveEffect}{' '}
                    {piecesWithoutActiveEffect === 1
                      ? 'peça ainda não ativa'
                      : 'peças ainda não ativam'}{' '}
                    um efeito deste conjunto.
                  </p>
                )}
                <small className="sonataDataSource">
                  Descrição oficial · Dados {sonata.gameVersion || 'sincronizados'}
                </small>
              </article>
            );
          })}
        </div>
      ) : (
        <p className="sonataSummaryEmpty">
          Defina a Sonata dos Echoes para ver os efeitos oficiais ativos e o progresso do conjunto.
        </p>
      )}
    </section>
  );
}

function CharacterInspector({
  character,
  draft,
  setDraft,
}: {
  character?: Character;
  draft: Build;
  setDraft: (build: Build) => void;
}) {
  if (!character)
    return (
      <div className="inspectorEmpty">
        <UserRound size={30} />
        <strong>Nenhum personagem</strong>
        <p>Selecione o personagem que receberá esta build.</p>
      </div>
    );
  return (
    <div className="buildInspectorContent">
      <header>
        <Portrait path={character.iconPath} name={character.name} large />
        <div>
          <span className="sectionLabel">PERSONAGEM</span>
          <h2>{character.name}</h2>
          <p>
            {character.element} · {character.weaponType}
          </p>
        </div>
      </header>
      <div className="inspectorForm">
        <label>
          Nível
          <input
            type="number"
            min={1}
            max={90}
            value={draft.characterLevel}
            onChange={(event) =>
              setDraft({ ...draft, characterLevel: clamp(Number(event.target.value), 1, 90) })
            }
          />
        </label>
        <label>
          Sequência
          <select
            value={draft.sequence}
            onChange={(event) => setDraft({ ...draft, sequence: Number(event.target.value) })}
          >
            {Array.from({ length: 7 }, (_, value) => (
              <option value={value} key={value}>
                S{value}
              </option>
            ))}
          </select>
        </label>
      </div>
      <div className="inspectorSummary">
        <span>RARIDADE</span>
        <strong>{'★'.repeat(character.rarity)}</strong>
        <span>VERSÃO</span>
        <strong>{character.gameVersion}</strong>
      </div>
    </div>
  );
}

function WeaponInspector({
  weapon,
  draft,
  setDraft,
}: {
  weapon?: Weapon;
  draft: Build;
  setDraft: (build: Build) => void;
}) {
  if (!weapon)
    return (
      <div className="inspectorEmpty">
        <Swords size={30} />
        <strong>Nenhuma arma</strong>
        <p>Selecione uma arma compatível com o personagem.</p>
      </div>
    );
  return (
    <div className="buildInspectorContent">
      <header>
        <Portrait path={weapon.iconPath} name={weapon.name} large />
        <div>
          <span className="sectionLabel">ARMA</span>
          <h2>{weapon.name}</h2>
          <p>
            {weapon.type} · {weapon.rarity} estrelas
          </p>
        </div>
      </header>
      <div className="inspectorForm">
        <label>
          Nível
          <input
            type="number"
            min={1}
            max={90}
            value={draft.weaponLevel}
            onChange={(event) =>
              setDraft({ ...draft, weaponLevel: clamp(Number(event.target.value), 1, 90) })
            }
          />
        </label>
        <label>
          Rank
          <select
            value={draft.weaponRank}
            onChange={(event) => setDraft({ ...draft, weaponRank: Number(event.target.value) })}
          >
            {[1, 2, 3, 4, 5].map((rank) => (
              <option value={rank} key={rank}>
                R{rank}
              </option>
            ))}
          </select>
        </label>
      </div>
      <div className="inspectorSummary">
        <span>ATK BASE</span>
        <strong>{weapon.baseAtk}</strong>
        <span>ATRIBUTO</span>
        <strong>{weapon.subStat || '—'}</strong>
      </div>
    </div>
  );
}

function EchoInspector({
  echo,
  sonatas,
  onChange,
  onRemove,
}: {
  echo: OwnedEcho;
  sonatas: Sonata[];
  onChange: (patch: Partial<OwnedEcho>) => void;
  onRemove: () => void;
}) {
  const substats = parseEchoSubstats(echo.substatsJson);
  const unlockedSlots = Math.min(5, Math.floor(echo.level / 5));
  const usedKeys = new Set(substats.map((item) => item.key).filter(Boolean));

  function updateSubstat(index: number, key: string, value?: string) {
    const choice = echoSubstatChoices.find((item) => item.key === key);
    const nextValue =
      choice && value && choice.values.includes(value) ? value : (choice?.values[0] ?? '');
    const next = Array.from({ length: 5 }, (_, position) =>
      position === index
        ? { key, value: key ? nextValue : '' }
        : (substats[position] ?? { key: '', value: '' })
    );
    onChange({ substatsJson: JSON.stringify(next.map(formatEchoSubstat)) });
  }
  function updateLevel(level: number) {
    const nextLevel = clamp(level, 0, 25);
    const nextUnlockedSlots = Math.min(5, Math.floor(nextLevel / 5));
    onChange({
      level: nextLevel,
      substatsJson: JSON.stringify(
        Array.from({ length: 5 }, (_, index) =>
          index < nextUnlockedSlots
            ? formatEchoSubstat(substats[index] ?? { key: '', value: '' })
            : ''
        )
      ),
    });
  }
  return (
    <div className="buildInspectorContent echoInspectorContent">
      <header className="echoInspectorHero">
        <Portrait path={echo.iconPath} name={echo.echoName} large />
        <div>
          <span className="sectionLabel">ECHO · CUSTO {echo.cost}</span>
          <h2>{echo.echoName}</h2>
          <p>{echo.sonataName || 'Sonata não definida'}</p>
        </div>
        <strong className="echoLevelTag">+{echo.level}</strong>
      </header>
      <div className="inspectorForm">
        <label>
          Nível
          <input
            type="number"
            min={0}
            max={25}
            value={echo.level}
            onChange={(event) => updateLevel(Number(event.target.value))}
          />
        </label>
        <label>
          Sonata
          <select
            value={echo.sonataId ?? 0}
            onChange={(event) =>
              onChange({
                sonataId: Number(event.target.value) || undefined,
                sonataName:
                  sonatas.find((item) => item.id === Number(event.target.value))?.name ?? '',
              })
            }
          >
            <option value={0}>Não definida</option>
            {sonatas.map((item) => (
              <option value={item.id} key={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </label>
        <label className="wide echoMainStat">
          Atributo principal
          <input
            value={echo.mainStat}
            onChange={(event) => onChange({ mainStat: event.target.value })}
            placeholder="Ex.: CRIT Rate 22%"
          />
        </label>
      </div>
      <section className="echoSubstats">
        <header>
          <div>
            <span className="sectionLabel">SUBATRIBUTOS</span>
            <strong>
              {substats.filter((item) => item.key).length}/{unlockedSlots} sintonizados
            </strong>
          </div>
          <small>Um slot é liberado a cada +5 níveis.</small>
        </header>
        {Array.from({ length: 5 }, (_, index) => {
          const entry = substats[index] ?? { key: '', value: '' };
          const choice = echoSubstatChoices.find((item) => item.key === entry.key);
          const locked = index >= unlockedSlots;
          return (
            <div className={`echoSubstatRow ${locked ? 'locked' : ''}`} key={index}>
              <span>{String(index + 1).padStart(2, '0')}</span>
              <label>
                <em>Atributo</em>
                <select
                  value={locked ? '' : entry.key}
                  disabled={locked}
                  onChange={(event) => updateSubstat(index, event.target.value)}
                >
                  <option value="">
                    {locked ? `Disponível no +${(index + 1) * 5}` : 'Selecionar'}
                  </option>
                  {echoSubstatChoices.map((item) => (
                    <option
                      value={item.key}
                      disabled={item.key !== entry.key && usedKeys.has(item.key)}
                      key={item.key}
                    >
                      {item.label}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <em>Valor</em>
                <select
                  value={entry.value}
                  disabled={locked || !choice}
                  onChange={(event) => updateSubstat(index, entry.key, event.target.value)}
                >
                  {!choice && <option value="">—</option>}
                  {choice?.values.map((value) => (
                    <option value={value} key={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>
              {!locked && entry.key && (
                <button title="Limpar subatributo" onClick={() => updateSubstat(index, '')}>
                  <X size={13} />
                </button>
              )}
            </div>
          );
        })}
      </section>
      <button className="removeEchoButton" onClick={onRemove}>
        <Trash2 size={14} />
        Remover deste slot
      </button>
    </div>
  );
}

function LibraryCard({
  name,
  path,
  meta,
  badge,
  selected,
  disabled,
  onClick,
}: {
  name: string;
  path: string;
  meta: string;
  badge?: string;
  selected: boolean;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      className={`buildLibraryCard ${selected ? 'selected' : ''}`}
      disabled={disabled}
      onClick={onClick}
    >
      <div>
        <Asset path={path} name={name} />
      </div>
      <span>
        <strong>{name}</strong>
        <small>{meta}</small>
      </span>
      {badge && <em>{badge}</em>}
      {selected && <Check size={14} />}
    </button>
  );
}

function Portrait({ path, name, large }: { path: string; name: string; large?: boolean }) {
  return (
    <span className={`sourcePortrait ${large ? 'large' : ''}`}>
      {path?.startsWith('/cache/') ? <img src={path} alt="" /> : initials(name)}
    </span>
  );
}

function Asset({ path, name }: { path: string; name: string }) {
  return path?.startsWith('/cache/') ? (
    <img src={path} alt="" loading="lazy" />
  ) : (
    <span>{initials(name)}</span>
  );
}

function emptyBuild(version: string): Build {
  return {
    id: 0,
    name: 'Nova build',
    characterId: 0,
    characterName: '',
    characterIcon: '',
    characterLevel: 90,
    sequence: 0,
    weaponId: undefined,
    weaponName: '',
    weaponIcon: '',
    weaponLevel: 90,
    weaponRank: 1,
    normalAttackLevel: 0,
    resonanceSkillLevel: 0,
    forteLevel: 0,
    liberationLevel: 0,
    introLevel: 0,
    echoes: [],
    targetEnemyId: undefined,
    rotationId: undefined,
    conditions: '',
    notes: '',
    favorite: false,
    locked: false,
    gameVersion: version,
    createdAt: '',
    updatedAt: '',
  };
}

function normalizeBuild(build: Build): Build {
  return {
    ...build,
    normalAttackLevel: build.normalAttackLevel ?? 0,
    resonanceSkillLevel: build.resonanceSkillLevel ?? 0,
    forteLevel: build.forteLevel ?? 0,
    liberationLevel: build.liberationLevel ?? 0,
    introLevel: build.introLevel ?? 0,
    echoes: build.echoes ?? [],
    rotationId: undefined,
  };
}

function ownedFromCatalog(echo: Echo, sonatas: Sonata[]): OwnedEcho {
  const sonataID = firstNumber(echo.sonataIdsJson);
  return {
    id: 0,
    echoId: echo.id,
    echoName: echo.name,
    iconPath: echo.iconPath,
    cost: echo.cost,
    mainStat: '',
    substatsJson: JSON.stringify(['', '', '', '', '']),
    level: 25,
    sonataId: sonataID,
    sonataName: sonatas.find((item) => item.id === sonataID)?.name ?? '',
    characterId: undefined,
    characterName: '',
    locked: false,
    favorite: false,
    note: '',
  };
}

function firstNumber(value: string): number | undefined {
  try {
    const parsed = JSON.parse(value);
    const first = Array.isArray(parsed) ? Number(parsed[0]) : 0;
    return first || undefined;
  } catch {
    return undefined;
  }
}

function parseIDs(value: string): number[] {
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed.map(Number).filter(Boolean) : [];
  } catch {
    return [];
  }
}

function compactSlots(items: OwnedEcho[]): OwnedEcho[] {
  return items.filter(Boolean).slice(0, 5);
}

function firstEmptySlot(items: OwnedEcho[]): number {
  return Math.min(items.length, 4);
}

function parseSubstats(value: string): string[] {
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed.map(String).slice(0, 5) : [];
  } catch {
    return [];
  }
}

function parseEchoSubstats(value: string): ParsedEchoSubstat[] {
  return parseSubstats(value).map((raw) => {
    const normalized = raw.trim();
    const choice = [...echoSubstatChoices]
      .sort((left, right) => right.label.length - left.label.length)
      .find(
        (item) =>
          normalized.startsWith(`${item.label} `) &&
          item.values.includes(normalized.slice(item.label.length + 1))
      );
    return choice
      ? { key: choice.key, value: normalized.slice(choice.label.length + 1) }
      : { key: '', value: '' };
  });
}

function formatEchoSubstat(item: ParsedEchoSubstat): string {
  const choice = echoSubstatChoices.find((option) => option.key === item.key);
  return choice && item.value ? `${choice.label} ${item.value}` : '';
}

function libraryTitle(target: Target) {
  return target.kind === 'character'
    ? 'Personagens'
    : target.kind === 'weapon'
      ? 'Armas compatíveis'
      : 'Echoes';
}

function initials(name: string) {
  return name
    ? name
        .split(/\s+/)
        .slice(0, 2)
        .map((part) => part[0])
        .join('')
        .toUpperCase()
    : '—';
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, Number.isFinite(value) ? value : min));
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}
