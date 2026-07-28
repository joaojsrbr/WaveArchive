import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { getWeapon, listWeapons, updateWeaponAccount } from './lib/backend';
import type { Weapon, WeaponAccountUpdate, WeaponFilter } from './types';

const weaponTypes = [
  [0, 'Todos'],
  [1, 'Broadblade'],
  [2, 'Sword'],
  [3, 'Pistols'],
  [4, 'Gauntlets'],
  [5, 'Rectifier'],
] as const;

const initialFilter: WeaponFilter = {
  query: '',
  type: 0,
  rarity: 0,
  ownedOnly: false,
  favorites: false,
  sort: 'name',
};

export function WeaponsPage({
  version,
  onError,
}: {
  version: string;
  onError: (message: string) => void;
}) {
  const [filter, setFilter] = useState<WeaponFilter>(readFilter);
  const [weapons, setWeapons] = useState<Weapon[]>([]);
  const [selected, setSelected] = useState<Weapon>();
  const [loading, setLoading] = useState(true);
  const [draft, setDraft] = useState<WeaponAccountUpdate>();
  const [saving, setSaving] = useState(false);

  async function load(next = filter) {
    try {
      setWeapons(await listWeapons(next));
      onError('');
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    const timer = window.setTimeout(() => void load(filter), 180);
    localStorage.setItem('wavearchive:weapon-filter', JSON.stringify(filter));
    return () => window.clearTimeout(timer);
  }, [filter]);

  async function open(id: number) {
    setLoading(true);
    try {
      setSelected(await getWeapon(id));
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setLoading(false);
    }
  }

  async function toggleFavorite(weapon: Weapon) {
    const update = {
      weaponId: weapon.id,
      owned: weapon.owned,
      level: weapon.level || 1,
      rank: weapon.rank || 1,
      favorite: !weapon.favorite,
    };
    try {
      await updateWeaponAccount(update);
      await load();
      if (selected?.id === weapon.id) setSelected(await getWeapon(weapon.id));
    } catch (cause) {
      onError(messageFrom(cause));
    }
  }

  async function saveAccount(event: FormEvent) {
    event.preventDefault();
    if (!draft) return;
    setSaving(true);
    try {
      await updateWeaponAccount(draft);
      setSelected(await getWeapon(draft.weaponId));
      await load();
      setDraft(undefined);
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setSaving(false);
    }
  }

  const resultLabel = useMemo(
    () => `${weapons.length} ${weapons.length === 1 ? 'arma' : 'armas'}`,
    [weapons.length]
  );

  if (selected)
    return (
      <div className="weaponDetailPage">
        <button className="backButton" onClick={() => setSelected(undefined)}>
          ← Voltar às armas
        </button>
        <section className="weaponHero">
          <div className="weaponShowcase">
            {selected.iconPath.startsWith('/cache/') ? (
              <img src={selected.iconPath} alt="" />
            ) : (
              <span>◇</span>
            )}
          </div>
          <div className="weaponHeroContent">
            <span className="eyebrow">
              {selected.type} · ARMA {selected.rarity} ESTRELAS
            </span>
            <h1>{selected.name}</h1>
            <div className="heroTags">
              <span>{'◆'.repeat(selected.rarity)}</span>
              <span>ATK {selected.baseAtk}</span>
              {selected.subStat && <span>{selected.subStat}</span>}
            </div>
            <p className="heroDescription">{selected.description}</p>
            <div className="heroActions">
              <button onClick={() => setDraft(accountDraft(selected))}>
                {selected.owned ? 'Editar minha arma' : 'Registrar na conta'}
              </button>
              <button className="primaryButton">Comparar arma</button>
            </div>
          </div>
          <aside className="heroMeta">
            <span>STATUS</span>
            <strong>
              {selected.owned ? `Nv. ${selected.level} · R${selected.rank}` : 'Não registrada'}
            </strong>
            <span>VERSÃO</span>
            <strong>{selected.gameVersion}</strong>
          </aside>
        </section>
        <section className="weaponEffectPanel">
          <span className="sectionLabel">PASSIVA · REFINAMENTO 1</span>
          <h2>{selected.effectName || 'Sem passiva'}</h2>
          <p>{selected.effect || 'Esta arma não possui uma passiva registrada.'}</p>
          <div className="weaponStats">
            <div>
              <span>ATK NV. 90</span>
              <strong>{selected.baseAtk}</strong>
            </div>
            <div>
              <span>ATRIBUTO</span>
              <strong>{selected.subStat || '—'}</strong>
            </div>
            <div>
              <span>TIPO</span>
              <strong>{selected.type}</strong>
            </div>
          </div>
        </section>
        {draft && (
          <WeaponAccountModal
            weapon={selected}
            draft={draft}
            setDraft={setDraft}
            saving={saving}
            onSubmit={saveAccount}
          />
        )}
      </div>
    );

  return (
    <div className="weaponsCatalog">
      <div className="pageIntro">
        <div>
          <span className="eyebrow">ARSENAL SINCRONIZADO</span>
          <h1>ARMAS</h1>
          <p>Compare atributos, passivas e registre o equipamento disponível na sua conta.</p>
        </div>
        <div className="versionCard">
          <span>VERSÃO DOS DADOS</span>
          <strong>{version || '—'}</strong>
          <small>{resultLabel}</small>
        </div>
      </div>
      <div className="toolbar">
        <label className="search">
          <span aria-hidden="true">⌕</span>
          <span className="srOnly">Pesquisar armas</span>
          <input
            value={filter.query}
            onChange={(event) => setFilter({ ...filter, query: event.target.value })}
            placeholder="Buscar arma ou passiva…"
          />
        </label>
        <select
          aria-label="Filtrar raridade"
          value={filter.rarity}
          onChange={(event) => setFilter({ ...filter, rarity: Number(event.target.value) })}
        >
          <option value={0}>Todas as raridades</option>
          {[5, 4, 3, 2, 1].map((rarity) => (
            <option value={rarity} key={rarity}>
              {rarity} estrelas
            </option>
          ))}
        </select>
        <select
          aria-label="Ordenar armas"
          value={filter.sort}
          onChange={(event) =>
            setFilter({ ...filter, sort: event.target.value as WeaponFilter['sort'] })
          }
        >
          <option value="name">Nome A–Z</option>
          <option value="rarity">Maior raridade</option>
          <option value="atk">Maior ATK</option>
          <option value="type">Tipo</option>
        </select>
      </div>
      <div className="filterRow">
        <div className="chips">
          {weaponTypes.map(([type, label]) => (
            <button
              key={type}
              className={filter.type === type ? 'chip active' : 'chip'}
              onClick={() => setFilter({ ...filter, type })}
            >
              {type > 0 && <i className="navDiamond" />}
              {label}
            </button>
          ))}
        </div>
        <label className="check">
          <input
            type="checkbox"
            checked={filter.ownedOnly}
            onChange={(event) => setFilter({ ...filter, ownedOnly: event.target.checked })}
          />
          Somente possuídas
        </label>
        <label className="check">
          <input
            type="checkbox"
            checked={filter.favorites}
            onChange={(event) => setFilter({ ...filter, favorites: event.target.checked })}
          />
          Favoritas
        </label>
      </div>
      <div className="resultMeta">
        <strong>{resultLabel}</strong>
        {(filter.query || filter.type || filter.rarity || filter.ownedOnly || filter.favorites) && (
          <button onClick={() => setFilter(initialFilter)}>Limpar filtros</button>
        )}
      </div>
      {loading ? (
        <div className="skeletonGrid">
          {Array.from({ length: 10 }, (_, index) => (
            <div className="skeleton weaponSkeleton" key={index} />
          ))}
        </div>
      ) : weapons.length === 0 ? (
        <div className="emptyState">
          <div className="emptyGlyph">◇</div>
          <h2>Nenhuma arma encontrada</h2>
          <p>Sincronize os dados ou ajuste os filtros.</p>
        </div>
      ) : (
        <div className="weaponGrid">
          {weapons.map((weapon) => (
            <article
              className="weaponCard"
              key={weapon.id}
              tabIndex={0}
              role="button"
              onClick={() => void open(weapon.id)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') void open(weapon.id);
              }}
            >
              <div className="weaponImage">
                {weapon.iconPath.startsWith('/cache/') ? (
                  <img src={weapon.iconPath} alt="" loading="lazy" />
                ) : (
                  <span>◇</span>
                )}
                <div className="rarity">{'◆'.repeat(weapon.rarity)}</div>
              </div>
              <div className="weaponCardBody">
                <div>
                  <span>{weapon.type}</span>
                  <button
                    onClick={(event) => {
                      event.stopPropagation();
                      void toggleFavorite(weapon);
                    }}
                    aria-label={`Favoritar ${weapon.name}`}
                  >
                    {weapon.favorite ? '★' : '☆'}
                  </button>
                </div>
                <h2>{weapon.name}</h2>
                <p>
                  ATK <strong>{weapon.baseAtk}</strong>
                  <i />
                  {weapon.subStat || 'Sem atributo'}
                </p>
                <footer>
                  {weapon.owned ? `NV. ${weapon.level} · R${weapon.rank}` : 'NÃO REGISTRADA'}
                  <b>›</b>
                </footer>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

function WeaponAccountModal({
  weapon,
  draft,
  setDraft,
  saving,
  onSubmit,
}: {
  weapon: Weapon;
  draft: WeaponAccountUpdate;
  setDraft: (value?: WeaponAccountUpdate) => void;
  saving: boolean;
  onSubmit: (event: FormEvent) => void;
}) {
  return (
    <div className="modalBackdrop" onMouseDown={() => setDraft(undefined)}>
      <form
        className="accountModal"
        role="dialog"
        aria-modal="true"
        onSubmit={onSubmit}
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="modalHeader">
          <div>
            <span className="sectionLabel">INVENTÁRIO DE ARMAS</span>
            <h2>{weapon.name}</h2>
          </div>
          <button type="button" onClick={() => setDraft(undefined)}>
            ×
          </button>
        </div>
        <label className="ownershipToggle">
          <input
            type="checkbox"
            checked={draft.owned}
            onChange={(event) => setDraft({ ...draft, owned: event.target.checked })}
          />
          <span>
            <strong>Arma possuída</strong>
            <small>Disponibiliza a arma para builds e comparações da conta.</small>
          </span>
        </label>
        <div className="accountFields">
          <label>
            Nível
            <input
              type="number"
              min={1}
              max={90}
              disabled={!draft.owned}
              value={draft.level}
              onChange={(event) => setDraft({ ...draft, level: Number(event.target.value) })}
            />
          </label>
          <label>
            Refinamento
            <select
              disabled={!draft.owned}
              value={draft.rank}
              onChange={(event) => setDraft({ ...draft, rank: Number(event.target.value) })}
            >
              {[1, 2, 3, 4, 5].map((rank) => (
                <option value={rank} key={rank}>
                  R{rank}
                </option>
              ))}
            </select>
          </label>
        </div>
        <label className="ownershipToggle compact">
          <input
            type="checkbox"
            checked={draft.favorite}
            onChange={(event) => setDraft({ ...draft, favorite: event.target.checked })}
          />
          <span>
            <strong>Favorita</strong>
            <small>Mantém a arma nos seus atalhos.</small>
          </span>
        </label>
        <div className="modalActions">
          <button type="button" onClick={() => setDraft(undefined)}>
            Cancelar
          </button>
          <button className="primaryButton" disabled={saving}>
            {saving ? 'Salvando…' : 'Salvar arma'}
          </button>
        </div>
      </form>
    </div>
  );
}

function accountDraft(weapon: Weapon): WeaponAccountUpdate {
  return {
    weaponId: weapon.id,
    owned: weapon.owned,
    level: weapon.level || 1,
    rank: weapon.rank || 1,
    favorite: weapon.favorite,
  };
}

function readFilter(): WeaponFilter {
  try {
    return {
      ...initialFilter,
      ...JSON.parse(localStorage.getItem('wavearchive:weapon-filter') ?? ''),
    };
  } catch {
    return initialFilter;
  }
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}
