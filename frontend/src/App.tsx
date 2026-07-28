import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Bot, Calculator, ChevronDown, CircleDotDashed,
  ChevronRight, Database, Gauge, Grid3X3, List, PackageOpen,
  RefreshCw, Search, Settings, ShieldCheck, Sparkles, Star,
  Swords, UserRound, UsersRound, Waves, X
} from "lucide-react";
import { CharacterDetail } from "./CharacterDetail";
import { BuildsPage } from "./BuildsPage";
import { WeaponsPage } from "./WeaponsPage";
import { EchoesPage } from "./EchoesPage";
import { TeamsPage } from "./TeamsPage";
import { CalculatorPage } from "./CalculatorPage";
import { AssistantPage } from "./AssistantPage";
import { AccountPage, DashboardPage, SettingsPage } from "./WorkspacePages";
import { cancelSync, catalogStatus, getCharacter, listCharacters, restoreLatestSnapshot, syncCharacters, updateCharacterAccount } from "./lib/backend";
import { isRoverCharacter, roverGender } from "./lib/characters";
import type { CatalogStatus, Character, CharacterAccountUpdate, CharacterFilter, CharacterProfile } from "./types";

type PageID = "dashboard" | "characters" | "weapons" | "echoes" | "sonata" | "teams" | "builds" | "calculator" | "account" | "ai" | "settings";

const navItems: { id: PageID; label: string }[] = [
  { id: "dashboard", label: "Visão geral" }, { id: "characters", label: "Personagens" },
  { id: "weapons", label: "Armas" }, { id: "echoes", label: "Echoes" },
  { id: "sonata", label: "Sonata Effects" }, { id: "teams", label: "Equipes" },
  { id: "builds", label: "Builds" }, { id: "calculator", label: "Calculadora" },
  { id: "account", label: "Minha conta" }, { id: "ai", label: "Assistente IA" },
  { id: "settings", label: "Configurações" }
];

const navGroups: { label: string; items: PageID[] }[] = [
  { label: "Arquivo", items: ["characters", "weapons", "echoes", "sonata"] },
  { label: "Planejamento", items: ["teams", "builds"] },
  { label: "Análise", items: ["calculator"] },
  { label: "Conta", items: ["account", "ai", "settings"] }
];

const pageIcons: Record<PageID, typeof Gauge> = {
  dashboard: Gauge, characters: UserRound, weapons: Swords, echoes: Waves,
  sonata: Sparkles, teams: UsersRound, builds: ShieldCheck, calculator: Calculator,
  account: Database, ai: Bot, settings: Settings
};

const elements = [
  [0, "Todos"], [1, "Glacio"], [2, "Fusion"], [3, "Electro"],
  [4, "Aero"], [5, "Spectro"], [6, "Havoc"]
] as const;

const initialFilter: CharacterFilter = {
  query: "",
  element: 0,
  rarity: 0,
  ownedOnly: false,
  favorites: false,
  sort: "name"
};

type SyncProgress = { stage: string; progress: number };

export function App() {
  const [characters, setCharacters] = useState<Character[]>([]);
  const [status, setStatus] = useState<CatalogStatus>({ count: 0, version: "" });
  const [filter, setFilter] = useState<CharacterFilter>(readStoredFilter);
  const [view, setView] = useState<"grid" | "table">(readStoredView);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState("");
  const [selectedID, setSelectedID] = useState<number>();
  const [profile, setProfile] = useState<CharacterProfile>();
  const [profileLoading, setProfileLoading] = useState(false);
  const [syncProgress, setSyncProgress] = useState<SyncProgress>();
  const [page, setPage] = useState<PageID>(readStoredPage);
  const [openNav, setOpenNav] = useState<string | null>(null);

  const load = useCallback(async (nextFilter: CharacterFilter) => {
    try {
      setCharacters(await listCharacters(nextFilter));
      setError("");
    } catch (cause) {
      setError(messageFrom(cause));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(filter), 180);
    return () => window.clearTimeout(timer);
  }, [filter, load]);

  useEffect(() => {
    localStorage.setItem("wavearchive:catalog-filter", JSON.stringify(filter));
  }, [filter]);

  useEffect(() => {
    localStorage.setItem("wavearchive:catalog-view", view);
  }, [view]);

  useEffect(() => {
    localStorage.setItem("wavearchive:page", page);
  }, [page]);

  useEffect(() => {
    function shortcuts(event: KeyboardEvent) {
      if (event.ctrlKey && event.key.toLowerCase() === "k") {
        event.preventDefault();
        document.querySelector<HTMLInputElement>(".content .search input")?.focus();
      } else if (event.ctrlKey && event.key === ",") {
        event.preventDefault();
        setPage("settings");
      } else if (event.key === "Escape") {
        if (openNav) setOpenNav(null);
        else if (selectedID !== undefined) setSelectedID(undefined);
      }
    }
    window.addEventListener("keydown", shortcuts);
    return () => window.removeEventListener("keydown", shortcuts);
  }, [openNav, selectedID]);

  useEffect(() => {
    function closeNavigation(event: PointerEvent) {
      if (!(event.target as Element).closest(".navGroup")) setOpenNav(null);
    }
    window.addEventListener("pointerdown", closeNavigation);
    return () => window.removeEventListener("pointerdown", closeNavigation);
  }, []);

  useEffect(() => {
    void catalogStatus().then(setStatus).catch((cause) => setError(messageFrom(cause)));
  }, []);

  useEffect(() => {
    const unsubscribe = window.runtime?.EventsOn("catalog:sync", (payload) => {
      if (isSyncProgress(payload)) setSyncProgress(payload);
    });
    return () => unsubscribe?.();
  }, []);

  useEffect(() => {
    if (selectedID === undefined) {
      setProfile(undefined);
      return;
    }
    setProfileLoading(true);
    void getCharacter(selectedID)
      .then(setProfile)
      .catch((cause) => setError(messageFrom(cause)))
      .finally(() => setProfileLoading(false));
  }, [selectedID]);

  async function sync() {
    setSyncing(true);
    setSyncProgress({ stage: "detecting", progress: 5 });
    setError("");
    try {
      await syncCharacters();
      const nextStatus = await catalogStatus();
      setStatus(nextStatus);
      await load(filter);
    } catch (cause) {
      const message = messageFrom(cause);
      if (!message.toLowerCase().includes("cancel")) setError(message);
    } finally {
      setSyncing(false);
      window.setTimeout(() => setSyncProgress(undefined), 900);
    }
  }

  async function restore() {
    if (!window.confirm("Restaurar o snapshot mais recente? O estado atual será salvo antes da restauração.")) return;
    setLoading(true);
    setError("");
    try {
      await restoreLatestSnapshot();
      setStatus(await catalogStatus());
      setSelectedID(undefined);
      await load(filter);
    } catch (cause) {
      setError(messageFrom(cause));
    } finally {
      setLoading(false);
    }
  }

  async function saveAccount(update: CharacterAccountUpdate) {
    setError("");
    try {
      await updateCharacterAccount(update);
      await load(filter);
      if (selectedID === update.characterId) {
        setProfile(await getCharacter(update.characterId));
      }
    } catch (cause) {
      setError(messageFrom(cause));
      throw cause;
    }
  }

  async function toggleFavorite(character: Character) {
    await saveAccount({
      characterId: character.id,
      owned: character.owned,
      level: character.level || 1,
      sequence: character.sequence,
      favorite: !character.favorite
    });
  }

  const resultLabel = useMemo(() => {
    const roverCount=characters.filter(isRoverCharacter).length;
    const visibleCount=roverCount>1?characters.length-roverCount+1:characters.length;
    return `${visibleCount} ${visibleCount===1?"personagem":"personagens"}${roverCount>1?` · ${roverCount} formas Rover agrupadas`:""}`;
  },[characters]);

  return (
    <div className="shell">
      <a className="skipLink" href="#main-content">Pular para o conteúdo</a>
      <header className="topbar">
        <button className="brand" onClick={() => { setPage("dashboard"); setSelectedID(undefined); }} aria-label="Abrir visão geral">
          <CircleDotDashed className="brandMark" strokeWidth={1.5} />
          <span><strong>WAVE</strong>ARCHIVE</span>
        </button>
        <nav className="globalNav" aria-label="Navegação principal">
          <button className={page === "dashboard" ? "globalNavLink active" : "globalNavLink"} onClick={() => { setPage("dashboard"); setSelectedID(undefined); }}>
            Início
          </button>
          {navGroups.map((group) => (
            <div className={`${group.items.includes(page) ? "navGroup active" : "navGroup"}${openNav === group.label ? " open" : ""}`} key={group.label}>
              <button
                className="navGroupTrigger"
                aria-expanded={openNav === group.label}
                aria-haspopup="menu"
                onClick={() => setOpenNav((current) => current === group.label ? null : group.label)}
              >
                {group.label}<ChevronDown size={14} strokeWidth={1.5} />
              </button>
              {openNav === group.label && <div className="navMenu" role="menu">
                {group.items.map((id) => {
                  const Icon = pageIcons[id];
                  return <button key={id} className={id === page ? "active" : ""} onClick={() => {
                    setPage(id);
                    setSelectedID(undefined);
                    setOpenNav(null);
                  }}><Icon size={17} strokeWidth={1.5} /><span>{pageLabel(id)}</span></button>;
                })}
              </div>}
            </div>
          ))}
        </nav>
        <div className="topActions">
          <span className="offlineBadge"><i /> LOCAL-FIRST</span>
          {syncing && <button className="cancelButton" onClick={() => void cancelSync()}>Cancelar</button>}
          <button className="syncButton" onClick={() => void sync()} disabled={syncing}>
            <RefreshCw size={15} strokeWidth={1.5} className={syncing ? "spin" : ""} />
            {syncing ? `${stageLabel(syncProgress?.stage)} ${syncProgress?.progress ?? 0}%` : status.version ? "Sincronizado" : "Sincronizar"}
          </button>
        </div>
      </header>

      <main id="main-content" tabIndex={-1}>
        <div className="contextBar">
          <span>{pageLabel(page)}</span>
          {page === "characters" && profile && <><ChevronDown size={12} className="contextSeparator" /><strong>{profile.character.name}</strong></>}
          <span className="dataVersion"><Database size={13} />{status.version ? `Dados ${status.version}` : "Sem dados sincronizados"}</span>
        </div>

        <section
          className={
            page === "teams" || page === "builds"
              ? "content contentWorkspace"
              : "content"
          }
        >
          {error && <div className="errorBanner" role="alert"><strong>Não foi possível concluir.</strong>{error}</div>}
          {page === "characters" ? (selectedID !== undefined ? (
            profileLoading || !profile
              ? <Loading />
              : <CharacterDetail profile={profile} onBack={() => setSelectedID(undefined)} onSaveAccount={saveAccount} />
          ) : <>
          <div className="pageIntro">
            <div>
              <span className="eyebrow">ARQUIVO DE RESSONADORES</span>
              <h1>PERSONAGENS</h1>
              <p>Explore o elenco, filtre funções e organize quem já faz parte da sua conta.</p>
            </div>
            <div className="versionCard">
              <span>VERSÃO DOS DADOS</span>
              <strong>{status.version || "—"}</strong>
              <small>{status.lastSyncAt ? `Atualizado ${formatDate(status.lastSyncAt)}` : "Aguardando primeira sincronização"}</small>
              {status.count > 0 && <button onClick={() => void restore()}>Restaurar snapshot</button>}
            </div>
          </div>

          <div className="toolbar">
            <label className="search">
              <Search size={16} strokeWidth={1.5} aria-hidden="true" />
              <span className="srOnly">Pesquisar personagens</span>
              <input
                value={filter.query}
                onChange={(event) => setFilter({ ...filter, query: event.target.value })}
                placeholder="Buscar por nome ou apelido…"
              />
              <kbd>Ctrl K</kbd>
            </label>
            <select
              aria-label="Ordenar personagens"
              value={filter.sort}
              onChange={(event) => setFilter({ ...filter, sort: event.target.value as CharacterFilter["sort"] })}
            >
              <option value="name">Nome A–Z</option>
              <option value="api">Ordem da API (lançamento)</option>
              <option value="rarity">Maior raridade</option>
              <option value="element">Elemento</option>
              <option value="id">ID</option>
            </select>
            <div className="viewToggle" aria-label="Modo de visualização">
              <button className={view === "grid" ? "selected" : ""} onClick={() => setView("grid")} aria-label="Visualização em grade"><Grid3X3 size={17} strokeWidth={1.5} /></button>
              <button className={view === "table" ? "selected" : ""} onClick={() => setView("table")} aria-label="Visualização em tabela"><List size={17} strokeWidth={1.5} /></button>
            </div>
          </div>

          <div className="filterRow">
            <div className="chips" aria-label="Filtrar por elemento">
              {elements.map(([code, name]) => (
                <button
                  key={code}
                  className={filter.element === code ? `chip active element-${code}` : "chip"}
                  onClick={() => setFilter({ ...filter, element: code })}
                >
                  {code > 0 && <i className={`elementDot element-${code}`} aria-hidden="true" />}
                  {name}
                </button>
              ))}
            </div>
            <label className="check"><input type="checkbox" checked={filter.ownedOnly} onChange={(event) => setFilter({ ...filter, ownedOnly: event.target.checked })} />Somente possuídos</label>
            <label className="check"><input type="checkbox" checked={filter.favorites} onChange={(event) => setFilter({ ...filter, favorites: event.target.checked })} />Favoritos</label>
          </div>

          <div className="resultMeta">
            <strong>{resultLabel}</strong>
            {(filter.query || filter.element || filter.ownedOnly || filter.favorites) && (
              <button onClick={() => setFilter(initialFilter)}>Limpar filtros</button>
            )}
          </div>

          {loading ? <Loading /> : characters.length === 0 ? (
            <EmptyState hasCatalog={status.count > 0} syncing={syncing} onSync={sync} />
          ) : view === "grid" ? (
            <CharacterGrid characters={characters} onOpen={setSelectedID} onFavorite={toggleFavorite} />
          ) : (
            <CharacterTable characters={characters} onOpen={setSelectedID} />
          )}
          </>) : page === "dashboard" ? <DashboardPage onError={setError} onNavigate={setPage} /> : page === "weapons" ? <WeaponsPage version={status.version} onError={setError} /> : page === "echoes" ? <EchoesPage onError={setError} /> : page === "sonata" ? <EchoesPage mode="sonata" onError={setError} /> : page === "teams" ? <TeamsPage version={status.version} onError={setError} /> : page === "builds" ? <BuildsPage version={status.version} onError={setError} /> : page === "calculator" ? <CalculatorPage onError={setError} /> : page === "account" ? <AccountPage onError={setError} /> : page === "ai" ? <AssistantPage onError={setError} /> : page === "settings" ? <SettingsPage onError={setError} /> : <ComingSoon page={page} onNavigate={setPage} />}
        </section>
      </main>
    </div>
  );
}

function CharacterGrid({ characters, onOpen, onFavorite }: { characters: Character[]; onOpen: (id: number) => void; onFavorite: (character: Character) => Promise<void> }) {
  const [showRoverPicker,setShowRoverPicker]=useState(false);
  const items=groupCharacterCatalog(characters);
  const rover=characters.filter(isRoverCharacter);
  return <><div className="characterGrid">{items.map((item) => item.kind==="rover"?(
    <RoverGroupCard key="rover-group" variants={item.variants} onOpen={()=>setShowRoverPicker(true)}/>
  ):(
    <article
      className="characterCard"
      key={item.character.id}
      role="button"
      tabIndex={0}
      onClick={() => onOpen(item.character.id)}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") onOpen(item.character.id);
      }}
    >
      <div className={`portrait element-${item.character.elementCode}`}>
        {item.character.iconPath.startsWith("/cache/") && <img src={item.character.iconPath} alt="" loading="lazy" onError={(event) => { event.currentTarget.style.display = "none"; }} />}
        <span>{initials(item.character.name)}</span>
        <RarityStars value={item.character.rarity} />
      </div>
      <div className="cardBody">
        <div className="cardTitle">
          <div><h2>{item.character.name}</h2>{item.character.nickname && item.character.nickname !== item.character.name && <p>{item.character.nickname}</p>}</div>
          <button className="favorite" onClick={(event) => { event.stopPropagation(); void onFavorite(item.character).catch(() => undefined); }} aria-label={`Favoritar ${item.character.name}`}><Star size={17} strokeWidth={1.5} fill={item.character.favorite ? "currentColor" : "none"} /></button>
        </div>
        <div className="tags">
          <span><i className={`elementDot element-${item.character.elementCode}`} />{item.character.element}</span>
          <span>{item.character.weaponType}</span>
        </div>
        <footer><span>{item.character.owned ? `NV. ${item.character.level} · S${item.character.sequence}` : "NÃO REGISTRADO"}</span><ChevronRight size={17} aria-hidden="true" /></footer>
      </div>
    </article>
  ))}</div>{showRoverPicker&&<RoverVariantPicker variants={rover} onClose={()=>setShowRoverPicker(false)} onOpen={id=>{setShowRoverPicker(false);onOpen(id)}}/>}</>;
}

function CharacterTable({ characters, onOpen }: { characters: Character[]; onOpen: (id: number) => void }) {
  const [showRoverPicker,setShowRoverPicker]=useState(false);
  const items=groupCharacterCatalog(characters);
  const rover=characters.filter(isRoverCharacter);
  return <><div className="tableWrap"><table>
    <thead><tr><th>Personagem</th><th>Elemento</th><th>Raridade</th><th>Arma</th><th>Conta</th><th /></tr></thead>
    <tbody>{items.map((item) => item.kind==="rover"?<tr key="rover-group" className="roverTableRow">
      <td><span className="miniPortrait roverMiniPortrait">R</span><span><strong>Rover</strong><small>{item.variants.length} variantes oficiais</small></span></td>
      <td><span className="tableTag">4 ELEMENTOS</span></td><td><RarityStars value={5}/></td><td>Sword</td>
      <td>{item.variants.filter(character=>character.owned).length} registradas</td><td><button onClick={()=>setShowRoverPicker(true)} aria-label="Escolher variante do Rover"><ChevronRight size={18}/></button></td>
    </tr>:<tr key={item.character.id}>
      <td><span className={`miniPortrait element-${item.character.elementCode}`}>{initials(item.character.name)}</span><span><strong>{item.character.name}</strong><small>{item.character.nickname}</small></span></td>
      <td><span className="tableTag"><i className={`elementDot element-${item.character.elementCode}`} />{item.character.element}</span></td>
      <td><RarityStars value={item.character.rarity} /></td>
      <td>{item.character.weaponType}</td>
      <td>{item.character.owned ? `Nv. ${item.character.level} · S${item.character.sequence}` : "Não registrado"}</td>
      <td><button onClick={() => onOpen(item.character.id)} aria-label={`Abrir ${item.character.name}`}><ChevronRight size={18} /></button></td>
    </tr>)}</tbody>
  </table></div>{showRoverPicker&&<RoverVariantPicker variants={rover} onClose={()=>setShowRoverPicker(false)} onOpen={id=>{setShowRoverPicker(false);onOpen(id)}}/>}</>;
}

type CharacterCatalogItem={kind:"character";character:Character}|{kind:"rover";variants:Character[]};

function groupCharacterCatalog(characters:Character[]):CharacterCatalogItem[]{
  const rover=characters.filter(isRoverCharacter);
  if(rover.length<2)return characters.map(character=>({kind:"character",character}));
  let added=false;
  return characters.flatMap(character=>{
    if(!isRoverCharacter(character))return [{kind:"character",character} as CharacterCatalogItem];
    if(added)return [];
    added=true;
    return [{kind:"rover",variants:rover} as CharacterCatalogItem];
  });
}

function RoverGroupCard({variants,onOpen}:{variants:Character[];onOpen:()=>void}){
  const male=variants.find(character=>roverGender(character)==="male");
  const female=variants.find(character=>roverGender(character)==="female");
  const elements=Array.from(new Map(variants.map(character=>[character.elementCode,character])).values());
  return <article className="characterCard roverGroupCard" role="button" tabIndex={0} onClick={onOpen} onKeyDown={event=>{if(event.key==="Enter"||event.key===" ")onOpen()}}>
    <div className="roverPortraitPair">{[female,male].map((character,index)=><div key={character?.id??index} className={`element-${character?.elementCode??0}`}>{character?.iconPath.startsWith("/cache/")&&<img src={character.iconPath} alt="" loading="lazy"/>}</div>)}<span className="roverMark">R</span></div>
    <div className="cardBody"><div className="cardTitle"><div><h2>Rover</h2><p>Masculino e feminino · {variants.length} variantes</p></div></div>
      <div className="roverElementList">{elements.map(character=><span key={character.elementCode}><i className={`elementDot element-${character.elementCode}`}/>{character.element}</span>)}</div>
      <footer><span>ESCOLHER GÊNERO E ELEMENTO</span><ChevronRight size={17}/></footer>
    </div>
  </article>;
}

function RoverVariantPicker({variants,onClose,onOpen}:{variants:Character[];onClose:()=>void;onOpen:(id:number)=>void}){
  const groups=(["female","male"] as const).map(gender=>({gender,characters:variants.filter(character=>roverGender(character)===gender).sort((a,b)=>a.apiOrder-b.apiOrder)})).filter(group=>group.characters.length);
  return <div className="modalBackdrop roverPickerBackdrop" onMouseDown={onClose}><section className="roverPicker" role="dialog" aria-modal="true" aria-labelledby="rover-picker-title" onMouseDown={event=>event.stopPropagation()}>
    <header><div><span className="sectionLabel">PERSONAGEM MULTIFORMA</span><h2 id="rover-picker-title">Escolha o Rover</h2><p>Gênero e elemento mantêm IDs, progressão e habilidades independentes.</p></div><button onClick={onClose} aria-label="Fechar"><X size={19}/></button></header>
    <div className="roverGenderGroups">{groups.map(group=><section key={group.gender}><div className="roverGenderHeading"><UserRound size={17}/><div><b>{group.gender==="female"?"Feminino":"Masculino"}</b><small>{group.characters.length} elementos</small></div></div><div className="roverVariants">{group.characters.map(character=><button key={character.id} className={`element-${character.elementCode}`} onClick={()=>onOpen(character.id)}>{character.iconPath.startsWith("/cache/")&&<img src={character.iconPath} alt=""/>}<span><i className={`elementDot element-${character.elementCode}`}/><b>{character.element}</b><small>{character.owned?`Nv. ${character.level} · S${character.sequence}`:"Não registrado"}</small></span><ChevronRight size={16}/></button>)}</div></section>)}</div>
  </section></div>;
}

function EmptyState({ hasCatalog, syncing, onSync }: { hasCatalog: boolean; syncing: boolean; onSync: () => void }) {
  return <div className="emptyState">
    <div className="emptyGlyph" aria-hidden="true"><PackageOpen size={38} strokeWidth={1.5} /></div>
    <h2>{hasCatalog ? "Nenhum personagem encontrado" : "O arquivo está vazio"}</h2>
    <p>{hasCatalog ? "Ajuste ou limpe os filtros para voltar a explorar o catálogo." : "Faça a primeira sincronização para baixar o índice oficial e usá-lo offline."}</p>
    {!hasCatalog && <button className="primaryButton" onClick={onSync} disabled={syncing}>{syncing ? "Sincronizando…" : "Sincronizar personagens"}</button>}
  </div>;
}

function Loading() {
  return <div className="skeletonGrid" aria-label="Carregando personagens">{Array.from({ length: 8 }, (_, index) => <div className="skeleton" key={index} />)}</div>;
}

function initials(name: string) {
  return name.split(/\s+/).slice(0, 2).map((part) => part[0]).join("").toUpperCase();
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("pt-BR", { dateStyle: "short", timeStyle: "short" }).format(new Date(value));
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}

function isSyncProgress(value: unknown): value is SyncProgress {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Record<string, unknown>;
  return typeof candidate.stage === "string" && typeof candidate.progress === "number";
}

function stageLabel(stage?: string) {
  const labels: Record<string, string> = {
    detecting: "Detectando",
    downloading: "Baixando índice",
    normalizing: "Normalizando",
    details: "Personagens",
    weapons: "Armas",
    assets: "Imagens",
    saving: "Salvando",
    done: "Concluído",
    cancelled: "Cancelado",
    weapon_index: "Índice de armas",
    weapon_details: "Detalhes das armas",
    weapon_assets: "Imagens das armas",
    weapon_saving: "Salvando armas",
    weapon_done: "Armas concluídas",
    echo_index: "Índice de Echoes",
    echo_details: "Detalhes dos Echoes",
    echo_assets: "Imagens dos Echoes",
    echo_saving: "Salvando Echoes",
    echo_done: "Echoes concluídos"
  };
  return stage ? labels[stage] ?? stage : "Sincronizando";
}

function readStoredFilter(): CharacterFilter {
  try {
    const value = JSON.parse(localStorage.getItem("wavearchive:catalog-filter") ?? "");
    return { ...initialFilter, ...value };
  } catch {
    return initialFilter;
  }
}

function readStoredView(): "grid" | "table" {
  return localStorage.getItem("wavearchive:catalog-view") === "table" ? "table" : "grid";
}

function readStoredPage(): PageID {
  const stored = localStorage.getItem("wavearchive:page") as PageID | null;
  return navItems.some((item) => item.id === stored) ? stored! : "characters";
}

function pageLabel(page: PageID) {
  return navItems.find((item) => item.id === page)?.label ?? "WaveArchive";
}

function ComingSoon({ page, onNavigate }: { page: PageID; onNavigate: (page: PageID) => void }) {
  const priorities: Partial<Record<PageID, string>> = {
    dashboard: "A visão geral será conectada aos favoritos, builds e metas da conta.",
    builds: "Builds entram depois da estabilização do catálogo de armas.",
    teams: "Equipes de três personagens serão conectadas às builds e buffs.",
    settings: "Preferências de densidade, janela, cache e provedores serão centralizadas aqui."
  };
  return <div className="comingSoon"><span className="eyebrow">MÓDULO PLANEJADO</span><div className="emptyGlyph"><PackageOpen size={38} strokeWidth={1.5} /></div><h1>{pageLabel(page)}</h1><p>{priorities[page] ?? "Este módulo faz parte do roadmap e ainda não foi conectado à fundação atual."}</p><div><button className="primaryButton" onClick={() => onNavigate("characters")}>Abrir personagens</button><button onClick={() => onNavigate("weapons")}>Abrir armas</button></div></div>;
}

function RarityStars({ value }: { value: number }) {
  return <span className="rarityIcons" aria-label={`${value} estrelas`}>{Array.from({ length: value }, (_, index) => <Star key={index} size={9} fill="currentColor" strokeWidth={1.5} />)}</span>;
}
