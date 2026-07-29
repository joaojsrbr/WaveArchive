import { useEffect, useMemo, useState, type CSSProperties } from 'react';
import {
  BookOpen,
  Box,
  Check,
  ChevronRight,
  Copy,
  ExternalLink,
  Globe2,
  Link2,
  ListFilter,
  Plus,
  RotateCcw,
  Save,
  Search,
  Sparkles,
  Star,
  Trash2,
  Unlink,
  UserRound,
  UsersRound,
  X,
} from 'lucide-react';
import {
  deleteTeam,
  duplicateTeam,
  getCharacter,
  listAllCharacterGuides,
  listBuilds,
  listCharacters,
  listTeams,
  restoreTeam,
  saveTeam,
} from './lib/backend';
import { readOpenTarget } from './lib/navigation';
import { useContextualShortcuts } from './lib/contextualShortcuts';
import { buildOfficialTeamSynergy, type OfficialSynergyItem } from './lib/teamSynergy';
import { LibraryFilterBar } from './LibraryFilterBar';
import type { Build, Character, CharacterGuide, CharacterProfile, Team, TeamMember } from './types';

type SourceTab = 'teams' | 'presets';
type TeamPreset = {
  key: string;
  name: string;
  source: string;
  characterIds: number[];
  characters: Character[];
};

const elements = [
  [0, 'Todos'],
  [1, 'Glacio'],
  [2, 'Fusion'],
  [3, 'Electro'],
  [4, 'Aero'],
  [5, 'Spectro'],
  [6, 'Havoc'],
] as const;

export function TeamsPage({
  version,
  onError,
  onOpenBuild,
}: {
  version: string;
  onError: (message: string) => void;
  onOpenBuild: (build: Build) => void;
}) {
  const [teams, setTeams] = useState<Team[]>([]);
  const [builds, setBuilds] = useState<Build[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [guides, setGuides] = useState<CharacterGuide[]>([]);
  const [draft, setDraft] = useState<Team>(() => emptyTeam(version));
  const [activeSlot, setActiveSlot] = useState(0);
  const [profiles, setProfiles] = useState<Map<number, CharacterProfile>>(new Map());
  const [sourceTab, setSourceTab] = useState<SourceTab>('teams');
  const [sourceQuery, setSourceQuery] = useState('');
  const [characterQuery, setCharacterQuery] = useState('');
  const [element, setElement] = useState(0);
  const [rarity, setRarity] = useState(0);
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [weaponType, setWeaponType] = useState(0);
  const [account, setAccount] = useState<'all' | 'owned' | 'missing'>('all');
  const [sort, setSort] = useState<'api' | 'name' | 'rarity' | 'level'>('api');
  const [deletedID, setDeletedID] = useState<number>();
  const [saving, setSaving] = useState(false);

  async function load(selectFirst = false) {
    try {
      const [nextTeams, nextCharacters, nextGuides, nextBuilds] = await Promise.all([
        listTeams(),
        listCharacters({
          query: '',
          element: 0,
          rarity: 0,
          ownedOnly: false,
          favorites: false,
          sort: 'api',
        }),
        listAllCharacterGuides(),
        listBuilds(),
      ]);
      setTeams(nextTeams);
      setCharacters(nextCharacters);
      setGuides(nextGuides);
      setBuilds(nextBuilds);
      const target = readOpenTarget('team');
      const selected = target ? nextTeams.find((item) => item.id === target.id) : undefined;
      if (selected) setDraft(cloneTeam(selected));
      else if (selectFirst && nextTeams[0]) setDraft(cloneTeam(nextTeams[0]));
      onError('');
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  useEffect(() => {
    void load(true);
  }, []);

  useEffect(() => {
    const characterIDs = [
      ...new Set(draft.members.map((member) => member.characterId).filter(Boolean)),
    ];
    if (!characterIDs.length) {
      setProfiles(new Map());
      return;
    }
    let cancelled = false;
    void Promise.all(
      characterIDs.map((characterID) =>
        getCharacter(characterID)
          .then((item) => [characterID, item] as const)
          .catch(() => undefined)
      )
    ).then((items) => {
      if (cancelled) return;
      setProfiles(
        new Map(
          items.filter((item): item is readonly [number, CharacterProfile] => item !== undefined)
        )
      );
    });
    return () => {
      cancelled = true;
    };
  }, [draft.members]);

  const characterMap = useMemo(
    () => new Map(characters.map((character) => [character.id, character])),
    [characters]
  );
  const buildMap = useMemo(() => new Map(builds.map((build) => [build.id, build])), [builds]);

  const selectedCharacterIDs = useMemo(
    () => new Set(draft.members.map((member) => member.characterId).filter(Boolean)),
    [draft.members]
  );

  const presets = useMemo<TeamPreset[]>(() => {
    const result: TeamPreset[] = [];
    for (const guide of guides) {
      (guide.teams || []).forEach((ids, index) => {
        const characterIds = ids.slice(0, 3);
        const resolved = characterIds
          .map((id) => characterMap.get(id))
          .filter(Boolean) as Character[];
        if (characterIds.length !== 3 || resolved.length !== 3 || new Set(characterIds).size !== 3)
          return;
        result.push({
          key: `${guide.id}-${index}`,
          name: guide.name || `Preset sincronizado ${String(index + 1).padStart(2, '0')}`,
          source: guide.source || 'guide-server.aki-game.net',
          characterIds,
          characters: resolved,
        });
      });
    }
    return result;
  }, [characterMap, guides]);

  const filteredCharacters = useMemo(() => {
    const list = characters.filter((character) => {
      const query = characterQuery.trim().toLocaleLowerCase();
      return (
        (!query || `${character.name} ${character.nickname}`.toLocaleLowerCase().includes(query)) &&
        (!element || character.elementCode === element) &&
        (!rarity || character.rarity === rarity) &&
        (!weaponType || character.weaponTypeCode === weaponType) &&
        (account === 'all' || (account === 'owned' ? character.owned : !character.owned)) &&
        (!favoritesOnly || character.favorite || selectedCharacterIDs.has(character.id))
      );
    });

    return [...list].sort((left, right) => {
      if (sort === 'name') return left.name.localeCompare(right.name);
      if (sort === 'rarity') return right.rarity - left.rarity || left.apiOrder - right.apiOrder;
      if (sort === 'level')
        return (right.level || 0) - (left.level || 0) || right.rarity - left.rarity;
      return left.apiOrder - right.apiOrder;
    });
  }, [
    account,
    characterQuery,
    characters,
    element,
    favoritesOnly,
    rarity,
    selectedCharacterIDs,
    sort,
    weaponType,
  ]);

  const characterFiltersActive = Boolean(
    characterQuery ||
    element ||
    rarity ||
    weaponType ||
    account !== 'all' ||
    favoritesOnly ||
    sort !== 'api'
  );

  function resetCharacterFilters() {
    setCharacterQuery('');
    setElement(0);
    setRarity(0);
    setWeaponType(0);
    setAccount('all');
    setFavoritesOnly(false);
    setSort('api');
  }

  const filteredTeams = useMemo(() => {
    const query = sourceQuery.trim().toLocaleLowerCase();
    return teams.filter((team) => !query || team.name.toLocaleLowerCase().includes(query));
  }, [sourceQuery, teams]);

  const filteredPresets = useMemo(() => {
    const query = sourceQuery.trim().toLocaleLowerCase();
    return presets.filter(
      (preset) => !query || `${preset.name} ${preset.source}`.toLocaleLowerCase().includes(query)
    );
  }, [presets, sourceQuery]);

  function create() {
    setDraft(emptyTeam(version));
    setActiveSlot(0);
  }

  function selectCharacter(character: Character) {
    const duplicateIndex = draft.members.findIndex(
      (member, index) => index !== activeSlot && member.characterId === character.id
    );
    if (duplicateIndex >= 0) {
      setActiveSlot(duplicateIndex);
      return;
    }
    const members = draft.members.map((member, index) =>
      index === activeSlot ? memberFromCharacter(character, index) : member
    );
    setDraft({ ...draft, members });
    const nextEmpty = members.findIndex(
      (member, index) => index > activeSlot && !member.characterId
    );
    if (nextEmpty >= 0) setActiveSlot(nextEmpty);
  }

  function removeMember(index: number) {
    const members = draft.members.map((member, position) =>
      position === index ? emptyMember(index) : member
    );
    setDraft({ ...draft, members });
    setActiveSlot(index);
  }

  function linkBuild(buildID: number | undefined) {
    const current = draft.members[activeSlot];
    if (!current?.characterId) return;
    const build = buildID ? buildMap.get(buildID) : undefined;
    if (build && build.characterId !== current.characterId) return;
    const members = draft.members.map((member, index) =>
      index === activeSlot
        ? {
            ...member,
            buildId: build?.id,
            buildName: build?.name || '',
          }
        : member
    );
    setDraft({ ...draft, members });
  }

  function applyPreset(preset: TeamPreset) {
    setDraft({
      ...emptyTeam(version),
      name: preset.name,
      members: preset.characters.map(memberFromCharacter),
      notes: `Fonte sincronizada: ${preset.source}`,
    });
    setActiveSlot(0);
  }

  async function submit(): Promise<boolean> {
    if (!canSave(draft)) return false;
    setSaving(true);
    try {
      const saved = await saveTeam(draft);
      setDraft(cloneTeam(saved));
      await load();
      onError('');
      return true;
    } catch (cause) {
      onError(messageFrom(cause));
      return false;
    } finally {
      setSaving(false);
    }
  }

  async function remove(team: Team) {
    try {
      await deleteTeam(team.id);
      setDeletedID(team.id);
      if (draft.id === team.id) create();
      await load();
    } catch (cause) {
      onError(team.locked ? 'Desbloqueie a equipe antes de excluí-la.' : messageFrom(cause));
    }
  }

  async function undo() {
    if (!deletedID) return;
    try {
      await restoreTeam(deletedID);
      setDeletedID(undefined);
      await load(true);
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  const activeCharacter = characterMap.get(draft.members[activeSlot]?.characterId);
  const profile = activeCharacter ? profiles.get(activeCharacter.id) : undefined;
  const activeMember = draft.members[activeSlot];
  const activeBuild = activeMember?.buildId ? buildMap.get(activeMember.buildId) : undefined;
  const compatibleBuilds = activeCharacter
    ? builds.filter((build) => build.characterId === activeCharacter.id)
    : [];
  const linkedBuildUnavailable = Boolean(activeMember?.buildId && !activeBuild);
  const teamSynergy = useMemo(
    () => buildOfficialTeamSynergy(draft.members, profiles),
    [draft.members, profiles]
  );
  const shortcutFeedback = useContextualShortcuts({
    canSave: canSave(draft) && !saving,
    onNew: create,
    onSave: submit,
    newMessage: 'Nova equipe criada.',
    savedMessage: 'Equipe salva.',
    invalidMessage: 'Adicione pelo menos um personagem antes de salvar.',
  });

  return (
    <div className="teamWorkspace">
      <aside className="teamSources" aria-label="Equipes e presets">
        <button className="newTeamButton" onClick={create}>
          <Plus size={16} />
          Nova equipe
        </button>
        <div className="sourceTabs" role="tablist">
          <button
            role="tab"
            aria-selected={sourceTab === 'teams'}
            className={sourceTab === 'teams' ? 'active' : ''}
            onClick={() => setSourceTab('teams')}
          >
            <UsersRound size={16} />
            Equipes
          </button>
          <button
            role="tab"
            aria-selected={sourceTab === 'presets'}
            className={sourceTab === 'presets' ? 'active' : ''}
            onClick={() => setSourceTab('presets')}
          >
            <BookOpen size={16} />
            Presets
          </button>
        </div>
        <label className="sourceSearch">
          <Search size={16} />
          <span className="srOnly">Pesquisar {sourceTab === 'teams' ? 'equipes' : 'presets'}</span>
          <input
            value={sourceQuery}
            onChange={(event) => setSourceQuery(event.target.value)}
            placeholder={`Buscar ${sourceTab === 'teams' ? 'equipes' : 'presets'}…`}
          />
          <ListFilter size={15} />
        </label>

        <div className="sourceList">
          {sourceTab === 'teams'
            ? filteredTeams.map((team) => (
                <article
                  className={team.id === draft.id ? 'sourceRow active' : 'sourceRow'}
                  key={team.id}
                >
                  <button
                    className="sourceRowMain"
                    onClick={() => {
                      setDraft(cloneTeam(team));
                      setActiveSlot(0);
                    }}
                  >
                    <span className="sourceRowTitle">
                      {team.favorite && <Star size={13} fill="currentColor" />}
                      {team.name}
                    </span>
                    <small>
                      <UserRound size={12} />
                      Dados do usuário
                    </small>
                    <span className="sourceAvatars">
                      {team.members.map((member) => (
                        <Portrait
                          key={member.slot}
                          src={member.characterIcon}
                          name={member.characterName}
                        />
                      ))}
                    </span>
                  </button>
                  <div className="sourceRowActions">
                    <button
                      title="Duplicar"
                      aria-label={`Duplicar ${team.name}`}
                      onClick={() =>
                        void duplicateTeam(team.id)
                          .then(() => load())
                          .catch((cause) => onError(messageFrom(cause)))
                      }
                    >
                      <Copy size={15} />
                    </button>
                    <button
                      title="Excluir"
                      aria-label={`Excluir ${team.name}`}
                      onClick={() => void remove(team)}
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                </article>
              ))
            : filteredPresets.map((preset) => (
                <button
                  className="sourceRow presetRow"
                  key={preset.key}
                  onClick={() => applyPreset(preset)}
                >
                  <span className="sourceRowTitle">{preset.name}</span>
                  <small>
                    <Globe2 size={12} />
                    {sourceLabel(preset.source)}
                  </small>
                  <span className="sourceAvatars">
                    {preset.characters.map((character) => (
                      <Portrait key={character.id} src={character.iconPath} name={character.name} />
                    ))}
                  </span>
                  <ChevronRight size={16} className="sourceChevron" />
                </button>
              ))}
          {((sourceTab === 'teams' && filteredTeams.length === 0) ||
            (sourceTab === 'presets' && filteredPresets.length === 0)) && (
            <div className="sourceEmpty">
              <span>
                {sourceTab === 'teams' ? 'Nenhuma equipe salva.' : 'Nenhum preset sincronizado.'}
              </span>
              <small>
                {sourceTab === 'teams'
                  ? 'Crie sua primeira composição.'
                  : 'Sincronize guias para exibir presets com fonte.'}
              </small>
            </div>
          )}
        </div>
      </aside>

      <section className="teamComposer">
        <header className="composerHeader">
          <div>
            <span className="sectionLabel">COMPOSIÇÃO DA EQUIPE</span>
            <input
              aria-label="Nome da equipe"
              maxLength={80}
              value={draft.name}
              onChange={(event) => setDraft({ ...draft, name: event.target.value })}
            />
          </div>
          <span>{draft.members.filter((member) => member.characterId).length}/3 personagens</span>
        </header>

        <div className="teamStage">
          {draft.members.map((member, index) => {
            const character = characterMap.get(member.characterId);
            return (
              <button
                className={`teamStageSlot element-${character?.elementCode || 0}${activeSlot === index ? ' active' : ''}`}
                key={index}
                onClick={() => setActiveSlot(index)}
                onContextMenu={(event) => {
                  if (!character) return;
                  event.preventDefault();
                  removeMember(index);
                }}
                title={character ? 'Clique direito para remover do slot' : undefined}
              >
                <span className="slotNumber">{String(index + 1).padStart(2, '0')}</span>
                {character ? (
                  <>
                    <CharacterArt character={character} />
                    <span className="slotIdentity">
                      <strong>{character.name}</strong>
                      <small>
                        <span className={`elementDot element-${character.elementCode}`} />
                        {character.element}
                        <i />
                        {character.weaponType}
                      </small>
                      <span className="rarityStars" aria-label={`${character.rarity} estrelas`}>
                        {'★'.repeat(character.rarity)}
                      </span>
                      {member.buildId && (
                        <span
                          className={
                            buildMap.has(member.buildId)
                              ? 'slotBuildBadge'
                              : 'slotBuildBadge unavailable'
                          }
                        >
                          <Link2 size={12} />
                          {buildMap.get(member.buildId)?.name || 'Build indisponível'}
                        </span>
                      )}
                    </span>
                    <span className="selectedMark">
                      <Check size={15} />
                      Selecionado
                    </span>
                  </>
                ) : (
                  <span className="emptySlotContent">
                    <Plus size={24} />
                    <strong>Selecionar personagem</strong>
                    <small>Escolha na biblioteca abaixo</small>
                  </span>
                )}
              </button>
            );
          })}
        </div>

        <TeamSynergySummary
          items={teamSynergy}
          selectedCount={draft.members.filter((member) => member.characterId).length}
          version={version}
          onSelectMember={(name) => {
            const index = draft.members.findIndex((member) => member.characterName === name);
            if (index >= 0) setActiveSlot(index);
          }}
        />

        <div className="composerActions">
          <div></div>
          <button onClick={create}>
            <Trash2 size={16} />
            Limpar
          </button>
          <button
            className="saveTeamButton"
            disabled={!canSave(draft) || saving}
            onClick={() => void submit()}
          >
            <Save size={16} />
            {saving ? 'Salvando…' : 'Salvar equipe'}
          </button>
        </div>
      </section>

      <aside className="characterInspector" aria-label="Dados do personagem selecionado">
        {activeCharacter && profile ? (
          <>
            <header>
              <Portrait src={activeCharacter.iconPath} name={activeCharacter.name} />
              <div>
                <h2>{activeCharacter.name}</h2>
                <span className="rarityStars">{'★'.repeat(activeCharacter.rarity)}</span>
              </div>
            </header>
            <div className="inspectorFacts">
              <span>
                <i className={`elementDot element-${activeCharacter.elementCode}`} />
                {activeCharacter.element}
              </span>
              <span>{activeCharacter.weaponType}</span>
            </div>
            <section className="teamBuildLink" aria-label="Build vinculada ao slot">
              <header>
                <div>
                  <span className="sectionLabel">BUILD DO SLOT</span>
                  <strong>
                    {compatibleBuilds.length}{' '}
                    {compatibleBuilds.length === 1 ? 'compatível' : 'compatíveis'}
                  </strong>
                </div>
              </header>
              <label>
                <span>Build de {activeCharacter.name}</span>
                <select
                  aria-label={`Build de ${activeCharacter.name}`}
                  value={activeMember?.buildId || ''}
                  onChange={(event) =>
                    linkBuild(event.target.value ? Number(event.target.value) : undefined)
                  }
                >
                  <option value="">Sem Build vinculada</option>
                  {linkedBuildUnavailable && activeMember?.buildId && (
                    <option value={activeMember.buildId}>Build indisponível</option>
                  )}
                  {compatibleBuilds.map((build) => (
                    <option value={build.id} key={build.id}>
                      {build.name}
                    </option>
                  ))}
                </select>
              </label>

              {activeBuild ? (
                <article className="linkedBuildCard">
                  <div className="linkedBuildIdentity">
                    <span>
                      <Link2 size={15} />
                    </span>
                    <div>
                      <strong>{activeBuild.name}</strong>
                      <small>{activeBuild.weaponName || 'Sem arma definida'}</small>
                    </div>
                  </div>
                  <div className="linkedBuildMetrics">
                    <span>
                      <small>CUSTO</small>
                      <strong>{buildEchoCost(activeBuild)}/12</strong>
                    </span>
                    <span>
                      <small>ECHOES</small>
                      <strong>{activeBuild.echoes.length}/5</strong>
                    </span>
                  </div>
                  {buildSonatas(activeBuild).length > 0 && (
                    <div className="linkedBuildSonatas" aria-label="Sonatas da Build">
                      {buildSonatas(activeBuild).map((sonata) => (
                        <span key={sonata}>{sonata}</span>
                      ))}
                    </div>
                  )}
                  <div className="linkedBuildActions">
                    <button type="button" onClick={() => onOpenBuild(activeBuild)}>
                      <ExternalLink size={14} />
                      Abrir Build
                    </button>
                    <button type="button" onClick={() => linkBuild(undefined)}>
                      <Unlink size={14} />
                      Desvincular
                    </button>
                  </div>
                </article>
              ) : linkedBuildUnavailable ? (
                <div className="linkedBuildUnavailable" role="status">
                  <Box size={18} />
                  <div>
                    <strong>Build indisponível</strong>
                    <small>
                      O personagem foi preservado. Escolha outra Build compatível ou remova o
                      vínculo.
                    </small>
                  </div>
                  <button type="button" onClick={() => linkBuild(undefined)}>
                    <Unlink size={14} />
                    Remover vínculo
                  </button>
                </div>
              ) : compatibleBuilds.length === 0 ? (
                <div className="teamBuildEmpty">
                  <Box size={18} />
                  <span>Nenhuma Build salva para {activeCharacter.name}.</span>
                </div>
              ) : null}
            </section>
            <section className="apiTags">
              <header>
                <div>
                  <span className="sectionLabel">TAGS DA API</span>
                  <strong>Nanoka {activeCharacter.gameVersion || version}</strong>
                </div>
              </header>
              {(profile.extras?.tags || []).length > 0 ? (
                <ul>
                  {profile.extras.tags.map((tag) => (
                    <li
                      key={tag.id}
                      style={
                        {
                          '--tag-color': tag.color ? `#${tag.color.replace('#', '')}` : undefined,
                        } as CSSProperties
                      }
                    >
                      {tag.iconPath?.startsWith('/cache/') ? (
                        <img src={tag.iconPath} alt="" />
                      ) : (
                        <span className="tagIndicator" />
                      )}
                      <div>
                        <strong>{tag.name}</strong>
                        <small>{tag.description}</small>
                      </div>
                    </li>
                  ))}
                </ul>
              ) : (
                <p>Não disponível para este personagem.</p>
              )}
            </section>
            <div className="inspectorActions">
              <button
                onClick={() =>
                  document
                    .querySelector<HTMLInputElement>('.characterLibrary .catalogFilterSearch input')
                    ?.focus()
                }
              >
                <RotateCcw size={16} />
                Substituir
              </button>
              <button className="danger" onClick={() => removeMember(activeSlot)}>
                <X size={16} />
                Remover do slot
              </button>
            </div>
          </>
        ) : (
          <div className="inspectorEmpty">
            <UserRound size={30} />
            <strong>Slot vazio</strong>
            <p>Escolha um personagem na biblioteca para ver os dados sincronizados.</p>
          </div>
        )}
      </aside>

      <section className="characterLibrary">
        <LibraryFilterBar
          title="Encontrar personagem"
          resultLabel={`${filteredCharacters.length} de ${characters.length}`}
          query={characterQuery}
          placeholder="Buscar por nome ou apelido…"
          sortValue={sort}
          sortLabel="Ordenar personagens"
          sortOptions={[
            { value: 'api', label: 'Lançamento' },
            { value: 'name', label: 'Nome A–Z' },
            { value: 'rarity', label: 'Maior raridade' },
            { value: 'level', label: 'Nível na conta' },
          ]}
          active={characterFiltersActive}
          onQueryChange={setCharacterQuery}
          onSortChange={(value) => setSort(value as typeof sort)}
          onReset={resetCharacterFilters}
        >
          <div className="catalogFacet catalogFacetWide">
            <span>Atributo</span>
            <div className="catalogChipRail">
              {elements.map(([value, label]) => (
                <button
                  type="button"
                  className={
                    element === value
                      ? `catalogChip active element-${value}`
                      : `catalogChip element-${value}`
                  }
                  onClick={() => setElement(value)}
                  key={value}
                >
                  {value > 0 && <i />}
                  {label}
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
                  className={rarity === item.value ? 'catalogChip active' : 'catalogChip'}
                  onClick={() => setRarity(item.value)}
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
              {[
                { value: 'all', label: 'Todos' },
                { value: 'owned', label: 'Possuídos' },
                { value: 'missing', label: 'Não possuídos' },
              ].map((item) => (
                <button
                  type="button"
                  className={account === item.value ? 'catalogChip active' : 'catalogChip'}
                  onClick={() => setAccount(item.value as typeof account)}
                  key={item.value}
                >
                  {item.label}
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
        </LibraryFilterBar>
        <div className="characterFilmstrip">
          {filteredCharacters.map((character) => {
            const used = draft.members.some((member) => member.characterId === character.id);
            return (
              <button
                className={`libraryCharacter element-${character.elementCode}${used ? ' used' : ''}`}
                key={character.id}
                onClick={() => selectCharacter(character)}
              >
                <CharacterArt character={character} compact />
                <span className="libraryIdentity">
                  <strong>{character.name}</strong>
                  <small>
                    {character.element} · {character.weaponType}
                  </small>
                </span>
                {used && (
                  <span className="librarySelected">
                    <Check size={13} />
                    Na equipe
                  </span>
                )}
              </button>
            );
          })}
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
          Equipe excluída.<button onClick={() => void undo()}>Desfazer</button>
          <button aria-label="Fechar" onClick={() => setDeletedID(undefined)}>
            <X size={15} />
          </button>
        </div>
      )}
    </div>
  );
}

function TeamSynergySummary({
  items,
  selectedCount,
  version,
  onSelectMember,
}: {
  items: OfficialSynergyItem[];
  selectedCount: number;
  version: string;
  onSelectMember: (name: string) => void;
}) {
  const shared = items.filter((item) => item.members.length > 1);
  const individual = items.filter((item) => item.members.length === 1);
  return (
    <section className="teamSynergy" aria-labelledby="team-synergy-title">
      <header>
        <div>
          <span className="sectionLabel">SINERGIA OFICIAL</span>
          <h3 id="team-synergy-title">Funções da composição</h3>
        </div>
        <small>Tags da API · Dados {version || 'sincronizados'}</small>
      </header>
      {selectedCount < 3 ? (
        <p className="teamSynergyEmpty">
          Complete os três slots para consolidar as funções oficiais da equipe.
        </p>
      ) : items.length === 0 ? (
        <p className="teamSynergyEmpty">
          A fonte sincronizada não forneceu tags oficiais para esta composição.
        </p>
      ) : (
        <div className="teamSynergyGroups">
          <SynergyGroup
            title="Funções compartilhadas"
            items={shared}
            empty="Nenhuma função oficial aparece em mais de um personagem."
            onSelectMember={onSelectMember}
            variant="shared"
          />
          <SynergyGroup
            title="Funções individuais"
            items={individual}
            empty="Todas as funções oficiais desta composição são compartilhadas."
            onSelectMember={onSelectMember}
            variant="individual"
          />
        </div>
      )}
      <footer>
        <Sparkles size={13} />
        Consolidação determinística; nenhuma pontuação ou recomendação foi gerada.
      </footer>
    </section>
  );
}

function SynergyGroup({
  title,
  items,
  empty,
  onSelectMember,
  variant,
}: {
  title: string;
  items: OfficialSynergyItem[];
  empty: string;
  onSelectMember: (name: string) => void;
  variant: 'shared' | 'individual';
}) {
  return (
    <div className={`teamSynergyGroup ${variant}`}>
      <header>
        <strong>{title}</strong>
        <small>{items.length}</small>
      </header>
      {items.length ? (
        <div>
          {items.map((item) => (
            <article
              key={item.key}
              style={
                {
                  '--synergy-color': item.tag.color
                    ? `#${item.tag.color.replace('#', '')}`
                    : undefined,
                } as CSSProperties
              }
            >
              {item.tag.iconPath?.startsWith('/cache/') ? (
                <img src={item.tag.iconPath} alt="" />
              ) : (
                <span />
              )}
              <div>
                <strong>{item.tag.name}</strong>
                {item.tag.description && <p>{item.tag.description}</p>}
                <small className="synergySources">
                  Fonte:
                  {item.members.map((member) => (
                    <button type="button" onClick={() => onSelectMember(member)} key={member}>
                      {member}
                    </button>
                  ))}
                </small>
              </div>
            </article>
          ))}
        </div>
      ) : (
        <p>{empty}</p>
      )}
    </div>
  );
}

function CharacterArt({ character, compact = false }: { character: Character; compact?: boolean }) {
  const src = character.backgroundPath?.startsWith('/cache/')
    ? character.backgroundPath
    : character.iconPath;
  return (
    <span className={compact ? 'characterArt compact' : 'characterArt'}>
      {src?.startsWith('/cache/') ? (
        <img
          src={src}
          alt=""
          loading="lazy"
          onError={(event) => {
            event.currentTarget.style.display = 'none';
          }}
        />
      ) : (
        <span>{initials(character.name)}</span>
      )}
    </span>
  );
}

function Portrait({ src, name }: { src: string; name: string }) {
  return (
    <span className="sourcePortrait">
      {src?.startsWith('/cache/') ? <img src={src} alt="" loading="lazy" /> : initials(name)}
    </span>
  );
}

function emptyMember(index: number): TeamMember {
  return {
    slot: index + 1,
    characterId: 0,
    characterName: '',
    characterIcon: '',
    buildName: '',
    role: '',
    customRole: '',
  };
}

function emptyTeam(version: string): Team {
  return {
    id: 0,
    name: 'Nova equipe',
    members: [emptyMember(0), emptyMember(1), emptyMember(2)],
    notes: '',
    favorite: false,
    locked: false,
    gameVersion: version,
    createdAt: '',
    updatedAt: '',
  };
}

function memberFromCharacter(character: Character, index: number): TeamMember {
  return {
    slot: index + 1,
    characterId: character.id,
    characterName: character.name,
    characterIcon: character.iconPath,
    buildName: '',
    role: '',
    customRole: '',
  };
}

function cloneTeam(team: Team): Team {
  const members = [0, 1, 2].map((index) =>
    team.members[index] ? { ...team.members[index] } : emptyMember(index)
  );
  return { ...team, members };
}

function canSave(team: Team) {
  const ids = team.members.map((member) => member.characterId).filter(Boolean);
  return team.name.trim().length > 0 && ids.length === 3 && new Set(ids).size === 3;
}

function sourceLabel(value: string) {
  try {
    return new URL(value).hostname;
  } catch {
    return value || 'Fonte sincronizada';
  }
}

function initials(value: string) {
  return (
    value
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0])
      .join('')
      .toUpperCase() || '?'
  );
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}

function buildEchoCost(build: Build) {
  return build.echoes.reduce((total, echo) => total + echo.cost, 0);
}

function buildSonatas(build: Build) {
  const counts = new Map<string, number>();
  build.echoes.forEach((echo) => {
    if (!echo.sonataName) return;
    counts.set(echo.sonataName, (counts.get(echo.sonataName) || 0) + 1);
  });
  return [...counts.entries()]
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
    .map(([name, count]) => `${name} · ${count}`);
}
