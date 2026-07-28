import { useEffect, useMemo, useState } from 'react';
import {
  Check,
  ChevronRight,
  Copy,
  Lock,
  Plus,
  Save,
  Search,
  SlidersHorizontal,
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
  listSonatas,
  listWeapons,
  restoreBuild,
  saveBuild,
  saveOwnedEcho,
} from './lib/backend';
import { isRoverCharacter } from './lib/characters';
import type { Build, Character, Echo, OwnedEcho, Sonata, Weapon } from './types';

type Target = { kind: 'character' } | { kind: 'weapon' } | { kind: 'echo'; slot: number };
type EchoSubstatChoice = { key: string; label: string; values: readonly string[] };
type ParsedEchoSubstat = { key: string; value: string };

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
  const [sonatas, setSonatas] = useState<Sonata[]>([]);
  const [draft, setDraft] = useState<Build>(() => emptyBuild(version));
  const [target, setTarget] = useState<Target>({ kind: 'character' });
  const [query, setQuery] = useState('');
  const [sourceQuery, setSourceQuery] = useState('');
  const [costFilter, setCostFilter] = useState(0);
  const [elementFilter, setElementFilter] = useState(0);
  const [rarityFilter, setRarityFilter] = useState(0);
  const [weaponTypeFilter, setWeaponTypeFilter] = useState(0);
  const [ownershipFilter, setOwnershipFilter] = useState<'all' | 'owned' | 'missing'>('all');
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [characterSort, setCharacterSort] = useState<'api' | 'name' | 'rarity' | 'level'>('api');
  const [weaponRarityFilter, setWeaponRarityFilter] = useState(0);
  const [weaponOwnershipFilter, setWeaponOwnershipFilter] = useState<'all' | 'owned' | 'missing'>(
    'all'
  );
  const [weaponFavoritesOnly, setWeaponFavoritesOnly] = useState(false);
  const [weaponSort, setWeaponSort] = useState<'rarity' | 'atk' | 'name' | 'api'>('rarity');
  const [sonataFilter, setSonataFilter] = useState(0);
  const [echoClassFilter, setEchoClassFilter] = useState('');
  const [echoSort, setEchoSort] = useState<'api' | 'name' | 'cost'>('api');
  const [signatureWeaponID, setSignatureWeaponID] = useState<number>();
  const [saving, setSaving] = useState(false);
  const [deletedID, setDeletedID] = useState<number>();

  async function load(selectID?: number) {
    try {
      const [nextBuilds, nextCharacters, nextWeapons, nextEchoes, nextSonatas] = await Promise.all([
        listBuilds(),
        listCharacters(characterFilter),
        listWeapons(weaponFilter),
        listEchoes(echoFilter),
        listSonatas(),
      ]);
      setBuilds(nextBuilds);
      setCharacters(nextCharacters);
      setWeapons(nextWeapons);
      setEchoes(nextEchoes);
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
    void load();
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
            (!weaponTypeFilter || item.weaponTypeCode === weaponTypeFilter) &&
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
            (weaponOwnershipFilter === 'all' ||
              item.owned === (weaponOwnershipFilter === 'owned')) &&
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
    weaponTypeFilter,
    ownershipFilter,
    favoritesOnly,
    characterSort,
    weaponRarityFilter,
    weaponOwnershipFilter,
    weaponFavoritesOnly,
    weaponSort,
    signatureWeaponID,
    sonataFilter,
    echoClassFilter,
    echoSort,
  ]);
  const visibleBuilds = builds.filter(
    (item) =>
      !sourceQuery.trim() ||
      item.name.toLocaleLowerCase('pt-BR').includes(sourceQuery.trim().toLocaleLowerCase('pt-BR'))
  );
  const characterWeaponTypes = uniqueOptions(
    characters.map((item) => [item.weaponTypeCode, item.weaponType] as const)
  );
  const echoClasses = [...new Set(echoes.map((item) => item.class).filter(Boolean))].sort();

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

  async function submit() {
    if (!canSave) return;
    setSaving(true);
    try {
      const savedEchoes: OwnedEcho[] = [];
      for (const echo of draft.echoes) {
        savedEchoes.push(await saveOwnedEcho({ ...echo, characterId: draft.characterId }));
      }
      const saved = await saveBuild({ ...draft, echoes: savedEchoes, rotationId: undefined });
      await load(saved.id);
      onError('');
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setSaving(false);
    }
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
        <div className="libraryFilters">
          <div>
            <span className="sectionLabel">BIBLIOTECA</span>
            <strong>{libraryTitle(target)}</strong>
            <small>{library.length} resultados</small>
          </div>
          <label className="characterLibrarySearch">
            <Search size={15} />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={`Buscar ${libraryTitle(target).toLocaleLowerCase('pt-BR')}...`}
            />
          </label>
          <span className="advancedFilterLabel">
            <SlidersHorizontal size={14} />
            Filtros avançados
          </span>
        </div>
        <div className="advancedLibraryFilters">
          {target.kind === 'character' && (
            <>
              <label>
                Elemento
                <select
                  value={elementFilter}
                  onChange={(event) => setElementFilter(Number(event.target.value))}
                >
                  <option value={0}>Todos</option>
                  {[
                    [1, 'Glacio'],
                    [2, 'Fusion'],
                    [3, 'Electro'],
                    [4, 'Aero'],
                    [5, 'Spectro'],
                    [6, 'Havoc'],
                  ].map(([id, name]) => (
                    <option value={id} key={id}>
                      {name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Raridade
                <select
                  value={rarityFilter}
                  onChange={(event) => setRarityFilter(Number(event.target.value))}
                >
                  <option value={0}>Todas</option>
                  <option value={5}>5 estrelas</option>
                  <option value={4}>4 estrelas</option>
                </select>
              </label>
              <label>
                Tipo de arma
                <select
                  value={weaponTypeFilter}
                  onChange={(event) => setWeaponTypeFilter(Number(event.target.value))}
                >
                  <option value={0}>Todos</option>
                  {characterWeaponTypes.map(([id, name]) => (
                    <option value={id} key={id}>
                      {name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Conta
                <select
                  value={ownershipFilter}
                  onChange={(event) =>
                    setOwnershipFilter(event.target.value as typeof ownershipFilter)
                  }
                >
                  <option value="all">Todos</option>
                  <option value="owned">Possuídos</option>
                  <option value="missing">Não possuídos</option>
                </select>
              </label>
              <label>
                Ordenar
                <select
                  value={characterSort}
                  onChange={(event) => setCharacterSort(event.target.value as typeof characterSort)}
                >
                  <option value="api">Lançamento</option>
                  <option value="name">Nome</option>
                  <option value="rarity">Raridade</option>
                  <option value="level">Nível</option>
                </select>
              </label>
              <label className="filterCheck">
                <input
                  type="checkbox"
                  checked={favoritesOnly}
                  onChange={(event) => setFavoritesOnly(event.target.checked)}
                />
                Somente favoritos
              </label>
            </>
          )}
          {target.kind === 'weapon' && (
            <>
              <label>
                Raridade
                <select
                  value={weaponRarityFilter}
                  onChange={(event) => setWeaponRarityFilter(Number(event.target.value))}
                >
                  <option value={0}>Todas</option>
                  <option value={5}>5 estrelas</option>
                  <option value={4}>4 estrelas</option>
                  <option value={3}>3 estrelas</option>
                </select>
              </label>
              <label>
                Conta
                <select
                  value={weaponOwnershipFilter}
                  onChange={(event) =>
                    setWeaponOwnershipFilter(event.target.value as typeof weaponOwnershipFilter)
                  }
                >
                  <option value="all">Todas</option>
                  <option value="owned">Possuídas</option>
                  <option value="missing">Não possuídas</option>
                </select>
              </label>
              <label>
                Ordenar
                <select
                  value={weaponSort}
                  onChange={(event) => setWeaponSort(event.target.value as typeof weaponSort)}
                >
                  <option value="rarity">Raridade</option>
                  <option value="atk">ATK base</option>
                  <option value="name">Nome</option>
                  <option value="api">ID da API</option>
                </select>
              </label>
              <span className="signaturePriority">
                <Star size={13} />
                {roverSelected
                  ? 'Arma recomendada sempre primeiro'
                  : 'Arma assinatura sempre primeiro'}
              </span>
              <label className="filterCheck">
                <input
                  type="checkbox"
                  checked={weaponFavoritesOnly}
                  onChange={(event) => setWeaponFavoritesOnly(event.target.checked)}
                />
                Somente favoritas
              </label>
            </>
          )}
          {target.kind === 'echo' && (
            <>
              <label>
                Custo
                <select
                  value={costFilter}
                  onChange={(event) => setCostFilter(Number(event.target.value))}
                >
                  <option value={0}>Todos</option>
                  <option value={1}>Custo 1</option>
                  <option value={3}>Custo 3</option>
                  <option value={4}>Custo 4</option>
                </select>
              </label>
              <label>
                Classe
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
              </label>
              <label>
                Ordenar
                <select
                  value={echoSort}
                  onChange={(event) => setEchoSort(event.target.value as typeof echoSort)}
                >
                  <option value="api">ID da API</option>
                  <option value="name">Nome</option>
                  <option value="cost">Maior custo</option>
                </select>
              </label>
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
            </>
          )}
        </div>
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
        </div>
      </section>

      {deletedID && (
        <div className="undoToast">
          Build excluída.<button onClick={() => void undo()}>Desfazer</button>
          <button onClick={() => setDeletedID(undefined)} aria-label="Fechar">
            ×
          </button>
        </div>
      )}
    </div>
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
  return { ...build, echoes: build.echoes ?? [], rotationId: undefined };
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

function uniqueOptions(items: readonly (readonly [number, string])[]) {
  return [
    ...new Map(
      items.filter(([id, name]) => id > 0 && Boolean(name)).map((item) => [item[0], item])
    ).values(),
  ].sort((left, right) => left[1].localeCompare(right[1]));
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
