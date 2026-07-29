import { useEffect, useMemo, useState } from 'react';
import {
  Bot,
  Database,
  Search,
  ShieldCheck,
  Sparkles,
  Star,
  Swords,
  UserRound,
  UsersRound,
  Waves,
  X,
} from 'lucide-react';
import {
  listAIConversations,
  listBuilds,
  listCharacters,
  listEchoes,
  listSonatas,
  listTeams,
  listWeapons,
} from './lib/backend';

type SearchKind =
  'all' | 'character' | 'weapon' | 'echo' | 'sonata' | 'build' | 'team' | 'conversation';
type SearchScope = 'all' | 'official' | 'personal';
type SearchResult = {
  key: string;
  kind: Exclude<SearchKind, 'all'>;
  id: number;
  title: string;
  subtitle: string;
  meta: string;
  iconPath?: string;
  favorite?: boolean;
  owned?: boolean;
};

const categories: Array<[SearchKind, string]> = [
  ['all', 'Tudo'],
  ['character', 'Personagens'],
  ['weapon', 'Armas'],
  ['echo', 'Echoes'],
  ['sonata', 'Sonatas'],
  ['build', 'Builds'],
  ['team', 'Equipes'],
  ['conversation', 'IA'],
];

export function GlobalSearch({
  open,
  onClose,
  onNavigate,
}: {
  open: boolean;
  onClose: () => void;
  onNavigate: (kind: SearchResult['kind'], id: number, title: string) => void;
}) {
  const [query, setQuery] = useState('');
  const [kind, setKind] = useState<SearchKind>('all');
  const [scope, setScope] = useState<SearchScope>('all');
  const [rarity, setRarity] = useState(0);
  const [element, setElement] = useState(0);
  const [weaponType, setWeaponType] = useState(0);
  const [cost, setCost] = useState(0);
  const [ownedOnly, setOwnedOnly] = useState(false);
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    if (!open) return;
    const timer = window.setTimeout(async () => {
      if (query.trim().length < 2) {
        setResults([]);
        return;
      }
      setLoading(true);
      try {
        const official = scope !== 'personal';
        const personal = scope !== 'official';
        const [characters, weapons, echoes, sonatas, builds, teams, conversations] =
          await Promise.all([
            official && (kind === 'all' || kind === 'character')
              ? listCharacters({
                  query,
                  element,
                  rarity,
                  weaponType,
                  ownedOnly,
                  favorites: favoritesOnly,
                  sort: 'api',
                })
              : [],
            official && (kind === 'all' || kind === 'weapon')
              ? listWeapons({
                  query,
                  type: weaponType,
                  rarity,
                  ownedOnly,
                  favorites: favoritesOnly,
                  sort: 'rarity',
                })
              : [],
            official && (kind === 'all' || kind === 'echo')
              ? listEchoes({
                  query,
                  cost,
                  sonataId: 0,
                  rarity,
                  ownedOnly,
                  favorites: favoritesOnly,
                  sort: 'name',
                })
              : [],
            official && (kind === 'all' || kind === 'sonata') ? listSonatas() : [],
            personal && (kind === 'all' || kind === 'build') ? listBuilds() : [],
            personal && (kind === 'all' || kind === 'team') ? listTeams() : [],
            personal && (kind === 'all' || kind === 'conversation') ? listAIConversations() : [],
          ]);
        const normalized = query.trim().toLocaleLowerCase('pt-BR');
        const next: SearchResult[] = [
          ...characters.slice(0, 16).map((item) => ({
            key: `character-${item.id}`,
            kind: 'character' as const,
            id: item.id,
            title: item.name,
            subtitle: `${item.element} · ${item.weaponType}`,
            meta: `${item.rarity} estrelas${item.owned ? ` · Nv. ${item.level}` : ''}`,
            iconPath: item.iconPath,
            favorite: item.favorite,
            owned: item.owned,
          })),
          ...weapons.slice(0, 16).map((item) => ({
            key: `weapon-${item.id}`,
            kind: 'weapon' as const,
            id: item.id,
            title: item.name,
            subtitle: `${item.type} · ${item.subStat || 'Sem atributo'}`,
            meta: `${item.rarity} estrelas · ATK ${item.baseAtk}`,
            iconPath: item.iconPath,
            favorite: item.favorite,
            owned: item.owned,
          })),
          ...echoes.slice(0, 16).map((item) => ({
            key: `echo-${item.id}`,
            kind: 'echo' as const,
            id: item.id,
            title: item.name,
            subtitle: `${item.class || item.type} · Custo ${item.cost}`,
            meta: item.ownedCount ? `${item.ownedCount} no inventário` : 'Não registrado',
            iconPath: item.iconPath,
            favorite: item.favorite,
            owned: item.ownedCount > 0,
          })),
          ...sonatas
            .filter((item) =>
              `${item.name} ${item.twoPiece} ${item.fivePiece}`
                .toLocaleLowerCase('pt-BR')
                .includes(normalized)
            )
            .slice(0, 12)
            .map((item) => ({
              key: `sonata-${item.id}`,
              kind: 'sonata' as const,
              id: item.id,
              title: item.name,
              subtitle: 'Sonata Effect',
              meta: item.gameVersion,
              iconPath: item.iconPath,
            })),
          ...builds
            .filter(
              (item) =>
                `${item.name} ${item.characterName} ${item.weaponName}`
                  .toLocaleLowerCase('pt-BR')
                  .includes(normalized) &&
                (!favoritesOnly || item.favorite)
            )
            .slice(0, 12)
            .map((item) => ({
              key: `build-${item.id}`,
              kind: 'build' as const,
              id: item.id,
              title: item.name,
              subtitle: `${item.characterName} · ${item.weaponName || 'Sem arma'}`,
              meta: `${item.echoes.length} Echoes`,
              iconPath: item.characterIcon,
              favorite: item.favorite,
              owned: true,
            })),
          ...teams
            .filter(
              (item) =>
                `${item.name} ${item.members.map((member) => member.characterName).join(' ')}`
                  .toLocaleLowerCase('pt-BR')
                  .includes(normalized) &&
                (!favoritesOnly || item.favorite)
            )
            .slice(0, 12)
            .map((item) => ({
              key: `team-${item.id}`,
              kind: 'team' as const,
              id: item.id,
              title: item.name,
              subtitle: item.members
                .map((member) => member.characterName)
                .filter(Boolean)
                .join(' · '),
              meta: `${item.members.filter((member) => member.characterId).length}/3 personagens`,
              favorite: item.favorite,
              owned: true,
            })),
          ...conversations
            .filter((item) =>
              `${item.title} ${item.messages.map((message) => message.content).join(' ')}`
                .toLocaleLowerCase('pt-BR')
                .includes(normalized)
            )
            .slice(0, 8)
            .map((item) => ({
              key: `conversation-${item.id}`,
              kind: 'conversation' as const,
              id: item.id,
              title: item.title,
              subtitle: `${item.provider} · ${item.model}`,
              meta: 'Conversa da IA',
            })),
        ];
        setResults(next.slice(0, 60));
        setActiveIndex(0);
      } finally {
        setLoading(false);
      }
    }, 160);
    return () => window.clearTimeout(timer);
  }, [cost, element, favoritesOnly, kind, open, ownedOnly, query, rarity, scope, weaponType]);

  const groupedCount = useMemo(
    () =>
      new Map(
        categories.map(([value]) => [
          value,
          value === 'all' ? results.length : results.filter((item) => item.kind === value).length,
        ])
      ),
    [results]
  );

  if (!open) return null;
  const choose = (result: SearchResult) => {
    onNavigate(result.kind, result.id, result.title);
    onClose();
  };

  return (
    <div className="globalSearchBackdrop" role="presentation" onMouseDown={onClose}>
      <section
        className="globalSearchDialog"
        role="dialog"
        aria-modal="true"
        aria-label="Pesquisa global"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <header className="globalSearchHeader">
          <Search size={20} />
          <input
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Pesquisar em todo o WaveArchive..."
            onKeyDown={(event) => {
              if (event.key === 'Escape') onClose();
              if (event.key === 'ArrowDown') {
                event.preventDefault();
                setActiveIndex((index) => Math.min(index + 1, results.length - 1));
              }
              if (event.key === 'ArrowUp') {
                event.preventDefault();
                setActiveIndex((index) => Math.max(index - 1, 0));
              }
              if (event.key === 'Enter' && results[activeIndex]) choose(results[activeIndex]);
            }}
          />
          <kbd>ESC</kbd>
          <button aria-label="Fechar pesquisa" onClick={onClose}>
            <X size={17} />
          </button>
        </header>
        <div className="globalSearchControls">
          <div className="globalSearchCategories">
            {categories.map(([value, label]) => (
              <button
                className={kind === value ? 'active' : ''}
                onClick={() => setKind(value)}
                key={value}
              >
                {label}
                <span>{groupedCount.get(value) || 0}</span>
              </button>
            ))}
          </div>
          <div className="globalSearchFilters">
            <label>
              Escopo
              <select
                value={scope}
                onChange={(event) => setScope(event.target.value as SearchScope)}
              >
                <option value="all">Tudo</option>
                <option value="official">Dados oficiais</option>
                <option value="personal">Dados do usuário</option>
              </select>
            </label>
            <label>
              Elemento
              <select value={element} onChange={(event) => setElement(Number(event.target.value))}>
                <option value={0}>Todos</option>
                <option value={1}>Glacio</option>
                <option value={2}>Fusion</option>
                <option value={3}>Electro</option>
                <option value={4}>Aero</option>
                <option value={5}>Spectro</option>
                <option value={6}>Havoc</option>
              </select>
            </label>
            <label>
              Raridade
              <select value={rarity} onChange={(event) => setRarity(Number(event.target.value))}>
                <option value={0}>Todas</option>
                <option value={5}>5 estrelas</option>
                <option value={4}>4 estrelas</option>
                <option value={3}>3 estrelas</option>
              </select>
            </label>
            <label>
              Arma
              <select
                value={weaponType}
                onChange={(event) => setWeaponType(Number(event.target.value))}
              >
                <option value={0}>Todas</option>
                <option value={1}>Broadblade</option>
                <option value={2}>Sword</option>
                <option value={3}>Pistols</option>
                <option value={4}>Gauntlets</option>
                <option value={5}>Rectifier</option>
              </select>
            </label>
            <label>
              Custo
              <select value={cost} onChange={(event) => setCost(Number(event.target.value))}>
                <option value={0}>Todos</option>
                <option value={1}>1</option>
                <option value={3}>3</option>
                <option value={4}>4</option>
              </select>
            </label>
            <button
              className={ownedOnly ? 'active' : ''}
              onClick={() => setOwnedOnly((value) => !value)}
            >
              <Database size={13} />
              Possuídos
            </button>
            <button
              className={favoritesOnly ? 'active' : ''}
              onClick={() => setFavoritesOnly((value) => !value)}
            >
              <Star size={13} fill={favoritesOnly ? 'currentColor' : 'none'} />
              Favoritos
            </button>
          </div>
        </div>
        <div className="globalSearchResults">
          {query.trim().length < 2 ? (
            <div className="globalSearchEmpty">
              <Search size={28} />
              <strong>Pesquise todos os dados do app</strong>
              <p>Digite pelo menos dois caracteres. Use as setas e Enter para navegar.</p>
            </div>
          ) : loading ? (
            <div className="globalSearchEmpty">
              <span className="spin">
                <Sparkles size={24} />
              </span>
              <strong>Consultando o índice local...</strong>
            </div>
          ) : results.length === 0 ? (
            <div className="globalSearchEmpty">
              <X size={28} />
              <strong>Nenhum resultado</strong>
              <p>Remova alguns filtros ou tente outro termo.</p>
            </div>
          ) : (
            results.map((result, index) => (
              <button
                className={
                  index === activeIndex ? 'globalSearchResult active' : 'globalSearchResult'
                }
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => choose(result)}
                key={result.key}
              >
                <span className="globalSearchResultIcon">
                  {result.iconPath?.startsWith('/cache/') ? (
                    <img src={result.iconPath} alt="" />
                  ) : (
                    <ResultIcon kind={result.kind} />
                  )}
                </span>
                <span>
                  <strong>{result.title}</strong>
                  <small>{result.subtitle}</small>
                </span>
                <span className="globalSearchMeta">
                  {result.favorite && <Star size={12} fill="currentColor" />}
                  {result.owned && <i>NA CONTA</i>}
                  {result.meta}
                </span>
              </button>
            ))
          )}
        </div>
        <footer>
          <span>
            <kbd>↑</kbd>
            <kbd>↓</kbd> navegar
          </span>
          <span>
            <kbd>Enter</kbd> abrir
          </span>
          <span>Índice local · sem internet</span>
        </footer>
      </section>
    </div>
  );
}

function ResultIcon({ kind }: { kind: SearchResult['kind'] }) {
  const Icon = {
    character: UserRound,
    weapon: Swords,
    echo: Waves,
    sonata: Sparkles,
    build: ShieldCheck,
    team: UsersRound,
    conversation: Bot,
  }[kind];
  return <Icon size={19} />;
}
