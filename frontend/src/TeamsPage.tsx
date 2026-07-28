import { useEffect, useMemo, useState, type CSSProperties } from "react";
import {
  BookOpen, Check, ChevronRight, Copy, Filter, Globe2, ListFilter,
  Plus, RotateCcw, Save, Search, Star, Trash2, UserRound, UsersRound, X
} from "lucide-react";
import {
  deleteTeam, duplicateTeam, getCharacter, listAllCharacterGuides,
  listCharacters, listTeams, restoreTeam, saveTeam
} from "./lib/backend";
import type { Character, CharacterGuide, CharacterProfile, Team, TeamMember } from "./types";

type SourceTab = "teams" | "presets";
type TeamPreset = {
  key: string;
  name: string;
  source: string;
  characterIds: number[];
  characters: Character[];
};

const elements = [
  [0, "Todos"], [1, "Glacio"], [2, "Fusion"], [3, "Electro"],
  [4, "Aero"], [5, "Spectro"], [6, "Havoc"]
] as const;

export function TeamsPage({ version, onError }: {
  version: string;
  onError: (message: string) => void;
}) {
  const [teams, setTeams] = useState<Team[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [guides, setGuides] = useState<CharacterGuide[]>([]);
  const [draft, setDraft] = useState<Team>(() => emptyTeam(version));
  const [activeSlot, setActiveSlot] = useState(0);
  const [profile, setProfile] = useState<CharacterProfile>();
  const [sourceTab, setSourceTab] = useState<SourceTab>("teams");
  const [sourceQuery, setSourceQuery] = useState("");
  const [characterQuery, setCharacterQuery] = useState("");
  const [element, setElement] = useState(0);
  const [rarity, setRarity] = useState(0);
  const [ownedOnly, setOwnedOnly] = useState(false);
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [sort, setSort] = useState<"api" | "name" | "rarity" | "level">("api");
  const [deletedID, setDeletedID] = useState<number>();
  const [saving, setSaving] = useState(false);

  async function load(selectFirst = false) {
    try {
      const [nextTeams, nextCharacters, nextGuides] = await Promise.all([
        listTeams(),
        listCharacters({ query: "", element: 0, rarity: 0, ownedOnly: false, favorites: false, sort: "api" }),
        listAllCharacterGuides()
      ]);
      setTeams(nextTeams);
      setCharacters(nextCharacters);
      setGuides(nextGuides);
      if (selectFirst && nextTeams[0]) setDraft(cloneTeam(nextTeams[0]));
      onError("");
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  useEffect(() => { void load(true); }, []);

  useEffect(() => {
    const characterID = draft.members[activeSlot]?.characterId;
    if (!characterID) {
      setProfile(undefined);
      return;
    }
    void getCharacter(characterID).then(setProfile).catch(() => setProfile(undefined));
  }, [draft.members, activeSlot]);

  const characterMap = useMemo(
    () => new Map(characters.map((character) => [character.id, character])),
    [characters]
  );

  const selectedCharacterIDs = useMemo(
    () => new Set(draft.members.map((member) => member.characterId).filter(Boolean)),
    [draft.members]
  );

  const presets = useMemo<TeamPreset[]>(() => {
    const result: TeamPreset[] = [];
    for (const guide of guides) {
      (guide.teams || []).forEach((ids, index) => {
        const characterIds = ids.slice(0, 3);
        const resolved = characterIds.map((id) => characterMap.get(id)).filter(Boolean) as Character[];
        if (characterIds.length !== 3 || resolved.length !== 3 || new Set(characterIds).size !== 3) return;
        result.push({
          key: `${guide.id}-${index}`,
          name: guide.name || `Preset sincronizado ${String(index + 1).padStart(2, "0")}`,
          source: guide.source || "guide-server.aki-game.net",
          characterIds,
          characters: resolved
        });
      });
    }
    return result;
  }, [characterMap, guides]);

  const filteredCharacters = useMemo(() => {
    const list = characters.filter((character) => {
      const query = characterQuery.trim().toLocaleLowerCase();
      return (!query || `${character.name} ${character.nickname}`.toLocaleLowerCase().includes(query))
        && (!element || character.elementCode === element)
        && (!rarity || character.rarity === rarity)
        && (!ownedOnly || character.owned)
        && (!favoritesOnly || character.favorite || selectedCharacterIDs.has(character.id));
    });

    return [...list].sort((left, right) => {
      if (sort === "name") return left.name.localeCompare(right.name);
      if (sort === "rarity") return right.rarity - left.rarity || left.apiOrder - right.apiOrder;
      if (sort === "level") return (right.level || 0) - (left.level || 0) || right.rarity - left.rarity;
      return left.apiOrder - right.apiOrder;
    });
  }, [characterQuery, characters, element, rarity, ownedOnly, favoritesOnly, selectedCharacterIDs, sort]);

  const filteredTeams = useMemo(() => {
    const query = sourceQuery.trim().toLocaleLowerCase();
    return teams.filter((team) => !query || team.name.toLocaleLowerCase().includes(query));
  }, [sourceQuery, teams]);

  const filteredPresets = useMemo(() => {
    const query = sourceQuery.trim().toLocaleLowerCase();
    return presets.filter((preset) => !query || `${preset.name} ${preset.source}`.toLocaleLowerCase().includes(query));
  }, [presets, sourceQuery]);

  function create() {
    setDraft(emptyTeam(version));
    setActiveSlot(0);
  }

  function selectCharacter(character: Character) {
    const duplicateIndex = draft.members.findIndex((member, index) => index !== activeSlot && member.characterId === character.id);
    if (duplicateIndex >= 0) {
      setActiveSlot(duplicateIndex);
      return;
    }
    const members = draft.members.map((member, index) => index === activeSlot ? memberFromCharacter(character, index) : member);
    setDraft({ ...draft, members });
    const nextEmpty = members.findIndex((member, index) => index > activeSlot && !member.characterId);
    if (nextEmpty >= 0) setActiveSlot(nextEmpty);
  }

  function removeMember(index: number) {
    const members = draft.members.map((member, position) => position === index ? emptyMember(index) : member);
    setDraft({ ...draft, members });
    setActiveSlot(index);
  }

  function applyPreset(preset: TeamPreset) {
    setDraft({
      ...emptyTeam(version),
      name: preset.name,
      members: preset.characters.map(memberFromCharacter),
      notes: `Fonte sincronizada: ${preset.source}`
    });
    setActiveSlot(0);
  }

  async function submit() {
    if (!canSave(draft)) return;
    setSaving(true);
    try {
      const saved = await saveTeam(draft);
      setDraft(cloneTeam(saved));
      await load();
      onError("");
    } catch (cause) {
      onError(messageFrom(cause));
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
      onError(team.locked ? "Desbloqueie a equipe antes de excluí-la." : messageFrom(cause));
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

  return <div className="teamWorkspace">
    <aside className="teamSources" aria-label="Equipes e presets">
      <div className="sourceTabs" role="tablist">
        <button role="tab" aria-selected={sourceTab === "teams"} className={sourceTab === "teams" ? "active" : ""} onClick={() => setSourceTab("teams")}>
          <UsersRound size={16} />Equipes
        </button>
        <button role="tab" aria-selected={sourceTab === "presets"} className={sourceTab === "presets" ? "active" : ""} onClick={() => setSourceTab("presets")}>
          <BookOpen size={16} />Presets
        </button>
      </div>
      <label className="sourceSearch">
        <Search size={16} />
        <span className="srOnly">Pesquisar {sourceTab === "teams" ? "equipes" : "presets"}</span>
        <input value={sourceQuery} onChange={(event) => setSourceQuery(event.target.value)} placeholder={`Buscar ${sourceTab === "teams" ? "equipes" : "presets"}…`} />
        <ListFilter size={15} />
      </label>

      <div className="sourceList">
        {sourceTab === "teams" ? filteredTeams.map((team) => (
          <article className={team.id === draft.id ? "sourceRow active" : "sourceRow"} key={team.id}>
            <button className="sourceRowMain" onClick={() => { setDraft(cloneTeam(team)); setActiveSlot(0); }}>
              <span className="sourceRowTitle">{team.favorite && <Star size={13} fill="currentColor" />}{team.name}</span>
              <small><UserRound size={12} />Dados do usuário</small>
              <span className="sourceAvatars">{team.members.map((member) => <Portrait key={member.slot} src={member.characterIcon} name={member.characterName} />)}</span>
            </button>
            <div className="sourceRowActions">
              <button title="Duplicar" aria-label={`Duplicar ${team.name}`} onClick={() => void duplicateTeam(team.id).then(() => load()).catch((cause) => onError(messageFrom(cause)))}><Copy size={15} /></button>
              <button title="Excluir" aria-label={`Excluir ${team.name}`} onClick={() => void remove(team)}><Trash2 size={15} /></button>
            </div>
          </article>
        )) : filteredPresets.map((preset) => (
          <button className="sourceRow presetRow" key={preset.key} onClick={() => applyPreset(preset)}>
            <span className="sourceRowTitle">{preset.name}</span>
            <small><Globe2 size={12} />{sourceLabel(preset.source)}</small>
            <span className="sourceAvatars">{preset.characters.map((character) => <Portrait key={character.id} src={character.iconPath} name={character.name} />)}</span>
            <ChevronRight size={16} className="sourceChevron" />
          </button>
        ))}
        {((sourceTab === "teams" && filteredTeams.length === 0) || (sourceTab === "presets" && filteredPresets.length === 0)) && (
          <div className="sourceEmpty">
            <span>{sourceTab === "teams" ? "Nenhuma equipe salva." : "Nenhum preset sincronizado."}</span>
            <small>{sourceTab === "teams" ? "Crie sua primeira composição." : "Sincronize guias para exibir presets com fonte."}</small>
          </div>
        )}
      </div>
      <button className="newTeamButton" onClick={create}><Plus size={16} />Nova equipe</button>
    </aside>

    <section className="teamComposer">
      <header className="composerHeader">
        <div>
          <span className="sectionLabel">COMPOSIÇÃO DA EQUIPE</span>
          <input aria-label="Nome da equipe" maxLength={80} value={draft.name} onChange={(event) => setDraft({ ...draft, name: event.target.value })} />
        </div>
        <span>{draft.members.filter((member) => member.characterId).length}/3 personagens</span>
      </header>

      <div className="teamStage">
        {draft.members.map((member, index) => {
          const character = characterMap.get(member.characterId);
          return <button className={`teamStageSlot element-${character?.elementCode || 0}${activeSlot === index ? " active" : ""}`} key={index} onClick={() => setActiveSlot(index)}>
            <span className="slotNumber">{String(index + 1).padStart(2, "0")}</span>
            {character ? <>
              <CharacterArt character={character} />
              <span className="slotIdentity">
                <strong>{character.name}</strong>
                <small><span className={`elementDot element-${character.elementCode}`} />{character.element}<i />{character.weaponType}</small>
                <span className="rarityStars" aria-label={`${character.rarity} estrelas`}>{"★".repeat(character.rarity)}</span>
              </span>
              <span className="selectedMark"><Check size={15} />Selecionado</span>
            </> : <span className="emptySlotContent"><Plus size={24} /><strong>Selecionar personagem</strong><small>Escolha na biblioteca abaixo</small></span>}
          </button>;
        })}
      </div>

      <div className="composerActions">
        <div>
        </div>
        <button onClick={create}><Trash2 size={16} />Limpar</button>
        <button className="saveTeamButton" disabled={!canSave(draft) || saving} onClick={() => void submit()}><Save size={16} />{saving ? "Salvando…" : "Salvar equipe"}</button>
      </div>
    </section>

    <aside className="characterInspector" aria-label="Dados do personagem selecionado">
      {activeCharacter && profile ? <>
        <header>
          <Portrait src={activeCharacter.iconPath} name={activeCharacter.name} />
          <div><h2>{activeCharacter.name}</h2><span className="rarityStars">{"★".repeat(activeCharacter.rarity)}</span></div>
        </header>
        <div className="inspectorFacts">
          <span><i className={`elementDot element-${activeCharacter.elementCode}`} />{activeCharacter.element}</span>
          <span>{activeCharacter.weaponType}</span>
        </div>
        <section className="apiTags">
          <header><div><span className="sectionLabel">TAGS DA API</span><strong>Nanoka {activeCharacter.gameVersion || version}</strong></div></header>
          {(profile.extras?.tags || []).length > 0 ? <ul>{profile.extras.tags.map((tag) => (
            <li key={tag.id} style={{ "--tag-color": tag.color ? `#${tag.color.replace("#", "")}` : undefined } as CSSProperties}>
              {tag.iconPath?.startsWith("/cache/") ? <img src={tag.iconPath} alt="" /> : <span className="tagIndicator" />}
              <div><strong>{tag.name}</strong><small>{tag.description}</small></div>
            </li>
          ))}</ul> : <p>Não disponível para este personagem.</p>}
        </section>
        <div className="inspectorActions">
          <button onClick={() => document.querySelector<HTMLInputElement>(".characterLibrarySearch input")?.focus()}><RotateCcw size={16} />Substituir</button>
          <button className="danger" onClick={() => removeMember(activeSlot)}><X size={16} />Remover do slot</button>
        </div>
      </> : <div className="inspectorEmpty"><UserRound size={30} /><strong>Slot vazio</strong><p>Escolha um personagem na biblioteca para ver os dados sincronizados.</p></div>}
    </aside>

    <section className="characterLibrary">
      <div className="libraryFilters">
        <label className="characterLibrarySearch">
          <Search size={17} />
          <span className="srOnly">Buscar personagem</span>
          <input value={characterQuery} onChange={(event) => setCharacterQuery(event.target.value)} placeholder="Buscar personagem…" />
        </label>
        <div className="compactFilters">
          <div className="filterSelects">
            <Filter size={15} />
            <select aria-label="Filtrar por elemento" value={element} onChange={(event) => setElement(Number(event.target.value))}>
              {elements.map(([value, label]) => <option value={value} key={value}>{label}</option>)}
            </select>
            <select aria-label="Filtrar por raridade" value={rarity} onChange={(event) => setRarity(Number(event.target.value))}>
              <option value={0}>Todas raridades</option>
              <option value={5}>5 Estrelas</option>
              <option value={4}>4 Estrelas</option>
            </select>
            <select aria-label="Ordenar personagens" value={sort} onChange={(event) => setSort(event.target.value as any)}>
              <option value="api">Lançamento</option>
              <option value="name">Nome A–Z</option>
              <option value="rarity">Maior raridade</option>
              <option value="level">Nível na conta</option>
            </select>
          </div>
          <div className="filterChips">
            <button
              type="button"
              className={ownedOnly ? "filterChip active" : "filterChip"}
              onClick={() => setOwnedOnly(!ownedOnly)}
            >
              <UserRound size={13} />
              <span>Possuídos</span>
            </button>
            <button
              type="button"
              className={favoritesOnly ? "filterChip active" : "filterChip"}
              onClick={() => setFavoritesOnly(!favoritesOnly)}
            >
              <Star size={13} fill={favoritesOnly ? "currentColor" : "none"} />
              <span>Favoritos</span>
            </button>
            {(characterQuery || element !== 0 || rarity !== 0 || ownedOnly || favoritesOnly || sort !== "api") && (
              <button
                type="button"
                className="filterReset"
                onClick={() => {
                  setCharacterQuery("");
                  setElement(0);
                  setRarity(0);
                  setOwnedOnly(false);
                  setFavoritesOnly(false);
                  setSort("api");
                }}
                title="Limpar filtros"
              >
                <X size={13} />
              </button>
            )}
          </div>
        </div>
      </div>
      <div className="characterFilmstrip">
        {filteredCharacters.map((character) => {
          const used = draft.members.some((member) => member.characterId === character.id);
          return <button className={`libraryCharacter element-${character.elementCode}${used ? " used" : ""}`} key={character.id} onClick={() => selectCharacter(character)}>
            <CharacterArt character={character} compact />
            <span className="libraryIdentity"><strong>{character.name}</strong><small>{character.element} · {character.weaponType}</small></span>
            {used && <span className="librarySelected"><Check size={13} />Na equipe</span>}
          </button>;
        })}
      </div>
    </section>

    {deletedID && <div className="undoToast">Equipe excluída.<button onClick={() => void undo()}>Desfazer</button><button aria-label="Fechar" onClick={() => setDeletedID(undefined)}><X size={15} /></button></div>}
  </div>;
}

function CharacterArt({ character, compact = false }: { character: Character; compact?: boolean }) {
  const src = character.backgroundPath?.startsWith("/cache/") ? character.backgroundPath : character.iconPath;
  return <span className={compact ? "characterArt compact" : "characterArt"}>
    {src?.startsWith("/cache/") ? <img src={src} alt="" loading="lazy" onError={(event) => { event.currentTarget.style.display = "none"; }} /> : <span>{initials(character.name)}</span>}
  </span>;
}

function Portrait({ src, name }: { src: string; name: string }) {
  return <span className="sourcePortrait">{src?.startsWith("/cache/") ? <img src={src} alt="" loading="lazy" /> : initials(name)}</span>;
}

function emptyMember(index: number): TeamMember {
  return { slot: index + 1, characterId: 0, characterName: "", characterIcon: "", buildName: "", role: "", customRole: "" };
}

function emptyTeam(version: string): Team {
  return {
    id: 0,
    name: "Nova equipe",
    members: [emptyMember(0), emptyMember(1), emptyMember(2)],
    notes: "",
    favorite: false,
    locked: false,
    gameVersion: version,
    createdAt: "",
    updatedAt: ""
  };
}

function memberFromCharacter(character: Character, index: number): TeamMember {
  return {
    slot: index + 1,
    characterId: character.id,
    characterName: character.name,
    characterIcon: character.iconPath,
    buildName: "",
    role: "",
    customRole: ""
  };
}

function cloneTeam(team: Team): Team {
  const members = [0, 1, 2].map((index) => team.members[index] ? { ...team.members[index] } : emptyMember(index));
  return { ...team, members };
}

function canSave(team: Team) {
  const ids = team.members.map((member) => member.characterId).filter(Boolean);
  return team.name.trim().length > 0 && ids.length === 3 && new Set(ids).size === 3;
}

function sourceLabel(value: string) {
  try { return new URL(value).hostname; } catch { return value || "Fonte sincronizada"; }
}

function initials(value: string) {
  return value.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]).join("").toUpperCase() || "?";
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}
