import { useEffect, useMemo, useState } from 'react';
import {
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Circle,
  CircleHelp,
  FileSearch,
  History,
  Import,
  Link2,
  LockKeyhole,
  RefreshCw,
  Search,
  ShieldCheck,
  Sparkles,
  Swords,
  Trash2,
  UserRound,
  X,
} from 'lucide-react';
import {
  deleteConveneHistory,
  getConveneOverview,
  importConveneFromGame,
  importConveneFromLogFile,
  importConveneURL,
} from './lib/backend';
import type {
  ConveneImportResult,
  ConveneOverview,
  ConvenePoolSummary,
  ConvenePull,
} from './types';

type Props = { onError(message: string): void };
type ImportMode = 'closed' | 'options' | 'url';

const pageSize = 40;

export function ConvenePage({ onError }: Props) {
  const [overview, setOverview] = useState<ConveneOverview>(emptyOverview);
  const [selectedPoolType, setSelectedPoolType] = useState(1);
  const [query, setQuery] = useState('');
  const [rarity, setRarity] = useState(0);
  const [resourceType, setResourceType] = useState('all');
  const [dateRange, setDateRange] = useState('all');
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [importing, setImporting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [importMode, setImportMode] = useState<ImportMode>('closed');
  const [rawURL, setRawURL] = useState('');
  const [notice, setNotice] = useState('');

  useEffect(() => {
    void getConveneOverview()
      .then((value) => {
        const normalized = normalizeOverview(value);
        setOverview(normalized);
        setSelectedPoolType(
          normalized.pools.find((pool) => pool.total > 0)?.poolType ??
            normalized.pools[0]?.poolType ??
            1
        );
      })
      .catch((cause) => onError(messageFrom(cause)))
      .finally(() => setLoading(false));
  }, [onError]);

  const selectedPool =
    overview.pools.find((pool) => pool.poolType === selectedPoolType) ?? overview.pools[0];

  const filteredPulls = useMemo(() => {
    const now = Date.now();
    return overview.pulls.filter((pull) => {
      if (selectedPoolType && pull.poolType !== selectedPoolType) return false;
      if (rarity && pull.rarity !== rarity) return false;
      if (resourceType !== 'all' && normalizeResourceType(pull.resourceType) !== resourceType)
        return false;
      if (
        query &&
        !`${pull.itemName} ${pull.poolName} ${pull.resourceType}`
          .toLowerCase()
          .includes(query.toLowerCase())
      )
        return false;
      if (dateRange !== 'all') {
        const days = Number(dateRange);
        const time = parseGameDate(pull.obtainedAt).getTime();
        if (Number.isFinite(time) && now - time > days * 86_400_000) return false;
      }
      return true;
    });
  }, [dateRange, overview.pulls, query, rarity, resourceType, selectedPoolType]);

  const totalPages = Math.max(1, Math.ceil(filteredPulls.length / pageSize));
  const visiblePulls = filteredPulls.slice((page - 1) * pageSize, page * pageSize);
  const timeline = useMemo(
    () => fiveStarTimeline(overview.pulls.filter((pull) => pull.poolType === selectedPoolType)),
    [overview.pulls, selectedPoolType]
  );
  const pityByPull = useMemo(
    () => buildPityMap(overview.pulls.filter((pull) => pull.poolType === selectedPoolType)),
    [overview.pulls, selectedPoolType]
  );

  useEffect(() => setPage(1), [dateRange, query, rarity, resourceType, selectedPoolType]);
  useEffect(() => {
    if (page > totalPages) setPage(totalPages);
  }, [page, totalPages]);

  async function runImport(task: () => Promise<ConveneImportResult>) {
    setImporting(true);
    setNotice('');
    onError('');
    try {
      const result = await task();
      const normalized = normalizeOverview(result.overview);
      setOverview(normalized);
      setSelectedPoolType(
        normalized.pools.find((pool) => pool.total > 0)?.poolType ??
          normalized.pools[0]?.poolType ??
          selectedPoolType
      );
      setImportMode('closed');
      setRawURL('');
      setNotice(
        result.imported > 0
          ? `${result.imported} novos giros importados · ${result.duplicates} já estavam no arquivo`
          : `Nenhum giro novo · ${result.duplicates} já estavam salvos`
      );
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setImporting(false);
    }
  }

  async function removeHistory() {
    if (
      !window.confirm(
        'Excluir todo o histórico de Convene importado? Esta ação remove os giros e banners salvos deste app.'
      )
    ) {
      return;
    }
    setDeleting(true);
    setNotice('');
    onError('');
    try {
      await deleteConveneHistory();
      setOverview({ ...emptyOverview, pools: [], pulls: [] });
      setSelectedPoolType(1);
      setPage(1);
      setQuery('');
      setRarity(0);
      setResourceType('all');
      setDateRange('all');
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setDeleting(false);
    }
  }

  return (
    <div className="convenePage">
      <header className="conveneHero">
        <div>
          <span className="eyebrow">ARQUIVO LOCAL DE CONVENE</span>
          <div className="conveneTitleLine">
            <h1>HISTÓRICO DE CONVENE</h1>
            <span className="convenePrivacyBadge">
              <LockKeyhole size={12} />
              LOCAL · PRIVADO
            </span>
          </div>
          <p>Importe seus giros, acompanhe o pity e preserve registros que saíram do jogo.</p>
        </div>
        <div className="conveneHeroActions">
          {overview.profile && (
            <button
              className="conveneDangerButton"
              disabled={importing || deleting}
              onClick={() => void removeHistory()}
            >
              <Trash2 size={15} />
              {deleting ? 'Excluindo…' : 'Excluir histórico'}
            </button>
          )}
          <button
            className="convenePrimaryButton"
            disabled={importing || deleting}
            onClick={() => setImportMode('options')}
          >
            <Import size={16} />
            Importar histórico
          </button>
        </div>
      </header>

      {notice && (
        <div className="conveneNotice" role="status">
          <ShieldCheck size={16} />
          {notice}
          <button aria-label="Fechar aviso" onClick={() => setNotice('')}>
            <X size={14} />
          </button>
        </div>
      )}

      <section className="conveneSummaryRibbon" aria-label="Resumo do histórico">
        <SummaryMetric
          label="Total de giros"
          value={formatNumber(overview.total)}
          detail="Todos os banners"
        />
        <SummaryMetric
          label="5 estrelas"
          value={formatNumber(overview.count5)}
          detail={rateLabel(overview.count5, overview.total)}
          tone="five"
        />
        <SummaryMetric
          label="4 estrelas"
          value={formatNumber(overview.count4)}
          detail={rateLabel(overview.count4, overview.total)}
          tone="four"
        />
        <SummaryMetric
          label="Última importação"
          value={overview.lastImportedAt ? formatDateTime(overview.lastImportedAt) : 'Nunca'}
          detail={
            overview.profile
              ? `UID ${maskPlayerID(overview.profile.playerId)}`
              : 'Nenhum perfil local'
          }
          tone="sync"
        />
      </section>

      {loading ? (
        <div className="conveneLoading">Carregando arquivo local…</div>
      ) : !overview.profile ? (
        <ConveneEmpty onImport={() => setImportMode('options')} />
      ) : (
        <>
          <section className="conveneTimelineSection">
            <div className="conveneSectionHeading">
              <div>
                <span className="eyebrow">LINHA DO TEMPO</span>
                <h2>Marcos de 5 estrelas</h2>
              </div>
              <span className="conveneTimelineBanner">
                <Sparkles size={13} />
                {selectedPool?.name ?? 'Banner selecionado'}
              </span>
            </div>
            <div className="conveneTimeline">
              {timeline.length > 0 ? (
                timeline.map((entry, index) => (
                  <article className="conveneMilestone" key={`${entry.pull.id}-${index}`}>
                    <div className="conveneMilestoneCard">
                      <div className="conveneMilestoneIdentity">
                        <ItemPortrait pull={entry.pull} />
                        <div>
                          <span>5★</span>
                          <strong>{entry.pull.itemName}</strong>
                        </div>
                      </div>
                      <div className="conveneMilestonePity">
                        <span>PITY</span>
                        <strong>{entry.pityExact ? entry.pity : `${entry.pity}+`}</strong>
                      </div>
                      <small>{entry.pityExact ? 'Registro exato' : 'Histórico parcial'}</small>
                    </div>
                    <div className="conveneMilestoneMarker" aria-hidden="true">
                      <Circle size={11} />
                    </div>
                    <div className="conveneMilestoneDate">
                      <strong>{formatDate(entry.pull.obtainedAt)}</strong>
                      <div>
                        Marco {index + 1} de {timeline.length}
                      </div>
                    </div>
                  </article>
                ))
              ) : (
                <div className="conveneTimelineEmpty">
                  <Sparkles size={18} />
                  Nenhum registro de 5 estrelas neste banner.
                </div>
              )}
            </div>
          </section>

          <div className="conveneWorkspace">
            <div className="conveneLedger">
              <ConveneFilters
                query={query}
                rarity={rarity}
                resourceType={resourceType}
                dateRange={dateRange}
                onQuery={setQuery}
                onRarity={setRarity}
                onResourceType={setResourceType}
                onDateRange={setDateRange}
                onClear={() => {
                  setQuery('');
                  setRarity(0);
                  setResourceType('all');
                  setDateRange('all');
                }}
              />
              <PullLedger pulls={visiblePulls} pityByPull={pityByPull} />
              <LedgerPagination
                page={page}
                totalPages={totalPages}
                total={filteredPulls.length}
                onPage={setPage}
              />
            </div>

            <aside className="conveneInspector">
              <PoolInspector pool={selectedPool} />
              <BannerSelector
                pools={overview.pools}
                selected={selectedPoolType}
                onSelect={setSelectedPoolType}
              />
            </aside>
          </div>
        </>
      )}

      {importMode !== 'closed' && (
        <ImportDialog
          mode={importMode}
          rawURL={rawURL}
          importing={importing}
          onURL={setRawURL}
          onMode={setImportMode}
          onClose={() => setImportMode('closed')}
          onAutomatic={() => void runImport(importConveneFromGame)}
          onLog={() => void runImport(importConveneFromLogFile)}
          onSubmitURL={() => void runImport(() => importConveneURL(rawURL))}
        />
      )}
    </div>
  );
}

function SummaryMetric({
  label,
  value,
  detail,
  tone = '',
}: {
  label: string;
  value: string;
  detail: string;
  tone?: string;
}) {
  return (
    <div className={`conveneMetric ${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{detail}</small>
    </div>
  );
}

function ConveneFilters({
  query,
  rarity,
  resourceType,
  dateRange,
  onQuery,
  onRarity,
  onResourceType,
  onDateRange,
  onClear,
}: {
  query: string;
  rarity: number;
  resourceType: string;
  dateRange: string;
  onQuery(value: string): void;
  onRarity(value: number): void;
  onResourceType(value: string): void;
  onDateRange(value: string): void;
  onClear(): void;
}) {
  const active = query || rarity || resourceType !== 'all' || dateRange !== 'all';
  return (
    <div className="conveneFilters">
      <label className="conveneSearch">
        <Search size={15} />
        <span className="srOnly">Buscar no histórico</span>
        <input
          value={query}
          onChange={(event) => onQuery(event.target.value)}
          placeholder="Buscar item, tipo ou banner…"
        />
      </label>
      <label>
        <span>Raridade</span>
        <select value={rarity} onChange={(event) => onRarity(Number(event.target.value))}>
          <option value={0}>Todas</option>
          <option value={5}>5 estrelas</option>
          <option value={4}>4 estrelas</option>
          <option value={3}>3 estrelas</option>
        </select>
      </label>
      <label>
        <span>Tipo</span>
        <select value={resourceType} onChange={(event) => onResourceType(event.target.value)}>
          <option value="all">Todos</option>
          <option value="character">Ressonador</option>
          <option value="weapon">Arma</option>
        </select>
      </label>
      <label>
        <span>Período</span>
        <select value={dateRange} onChange={(event) => onDateRange(event.target.value)}>
          <option value="all">Todo o arquivo</option>
          <option value="30">Últimos 30 dias</option>
          <option value="90">Últimos 90 dias</option>
          <option value="180">Últimos 180 dias</option>
        </select>
      </label>
      <button className="conveneClearFilters" disabled={!active} onClick={onClear}>
        Limpar
      </button>
    </div>
  );
}

function PullLedger({
  pulls,
  pityByPull,
}: {
  pulls: ConvenePull[];
  pityByPull: Map<number, number>;
}) {
  if (pulls.length === 0) {
    return (
      <div className="conveneLedgerEmpty">
        <Search size={20} />
        <strong>Nenhum giro corresponde aos filtros.</strong>
        <span>Altere os filtros ou selecione outro banner.</span>
      </div>
    );
  }
  let previousDate = '';
  return (
    <div className="conveneTableWrap">
      <table className="conveneTable">
        <thead>
          <tr>
            <th>Data</th>
            <th>Item</th>
            <th>Tipo</th>
            <th>Raridade</th>
            <th>Pity</th>
          </tr>
        </thead>
        <tbody>
          {pulls.map((pull) => {
            const date = formatDate(pull.obtainedAt);
            const showDate = date !== previousDate;
            previousDate = date;
            const rowPity = pityByPull.get(pull.id) ?? 0;
            return (
              <tr key={pull.id} className={`rarity-${pull.rarity}`}>
                <td>
                  {showDate && <strong>{date}</strong>}
                  <span>{formatTime(pull.obtainedAt)}</span>
                </td>
                <td>
                  <span className="conveneItemCell">
                    <ItemPortrait pull={pull} />
                    <span>
                      <strong>{pull.itemName}</strong>
                      <small>{pull.poolName}</small>
                    </span>
                  </span>
                </td>
                <td>
                  {normalizeResourceType(pull.resourceType) === 'character' ? (
                    <UserRound size={14} />
                  ) : (
                    <Swords size={14} />
                  )}
                  {resourceTypeLabel(pull.resourceType)}
                </td>
                <td>
                  <span className={`conveneStars stars-${pull.rarity}`}>
                    {'★'.repeat(pull.rarity)}
                  </span>
                </td>
                <td>{pull.rarity === 5 ? <strong>{rowPity || '—'}</strong> : rowPity || '—'}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function PoolInspector({ pool }: { pool?: ConvenePoolSummary }) {
  if (!pool) return null;
  const pityPercent = Math.min(100, (pool.currentPity / pool.hardPity) * 100);
  return (
    <section className="convenePoolInspector">
      <span className="eyebrow">BANNER SELECIONADO</span>
      <h2>{pool.name}</h2>
      <small>{pool.total} giros arquivados</small>
      <div className="convenePity">
        <div>
          <span>Pity atual</span>
          <strong>
            {pool.currentPity}
            <small>/{pool.hardPity}</small>
          </strong>
        </div>
        <div className="conveneGuarantee">
          <span>Garantia</span>
          <strong>{guaranteeLabel(pool.guaranteeState)}</strong>
        </div>
      </div>
      <div className="convenePityBar" aria-label={`Pity ${pool.currentPity} de ${pool.hardPity}`}>
        <i style={{ width: `${pityPercent}%` }} />
      </div>
      <div className="conveneDistribution">
        <Distribution label="5 estrelas" value={pool.count5} total={pool.total} rarity={5} />
        <Distribution label="4 estrelas" value={pool.count4} total={pool.total} rarity={4} />
        <Distribution label="3 estrelas" value={pool.count3} total={pool.total} rarity={3} />
      </div>
      <div className="conveneDisclosure">
        <CircleHelp size={14} />
        <span>
          Soft pity em 64 é uma <strong>estimativa da comunidade</strong>, não uma regra oficial
          publicada.
        </span>
      </div>
      {pool.historyPartial && (
        <div className="convenePartial">
          <History size={14} />
          <span>
            <strong>Histórico parcial.</strong> O pity exibido é o mínimo comprovado pelos dados
            disponíveis.
          </span>
        </div>
      )}
    </section>
  );
}

function Distribution({
  label,
  value,
  total,
  rarity,
}: {
  label: string;
  value: number;
  total: number;
  rarity: number;
}) {
  const percent = total ? (value / total) * 100 : 0;
  return (
    <div className={`conveneDistributionRow rarity-${rarity}`}>
      <span>{label}</span>
      <i>
        <b style={{ width: `${percent}%` }} />
      </i>
      <strong>{value}</strong>
      <small>{percent.toFixed(1)}%</small>
    </div>
  );
}

function BannerSelector({
  pools,
  selected,
  onSelect,
}: {
  pools: ConvenePoolSummary[];
  selected: number;
  onSelect(value: number): void;
}) {
  return (
    <section className="conveneBannerSelector">
      <div className="conveneBannerSelectorTitle">
        <span>BANNERS</span>
        <small>{pools.length} categorias</small>
      </div>
      <div className="conveneBannerList">
        {pools.map((pool) => (
          <button
            key={pool.poolType}
            className={selected === pool.poolType ? 'selected' : ''}
            onClick={() => onSelect(pool.poolType)}
          >
            {pool.kind.includes('weapon') ? <Swords size={15} /> : <UserRound size={15} />}
            <span>
              <strong>{pool.name}</strong>
              <small>
                {pool.total ? `${pool.total} giros · pity ${pool.currentPity}` : 'Sem registros'}
              </small>
            </span>
            <b>{String(pool.poolType).padStart(2, '0')}</b>
          </button>
        ))}
      </div>
    </section>
  );
}

function LedgerPagination({
  page,
  totalPages,
  total,
  onPage,
}: {
  page: number;
  totalPages: number;
  total: number;
  onPage(value: number): void;
}) {
  return (
    <div className="convenePagination">
      <span>
        {total ? `${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, total)}` : '0'} de{' '}
        {total}
      </span>
      <div>
        <button disabled={page === 1} onClick={() => onPage(page - 1)} aria-label="Página anterior">
          <ChevronLeft size={15} />
        </button>
        <strong>
          {page} / {totalPages}
        </strong>
        <button
          disabled={page === totalPages}
          onClick={() => onPage(page + 1)}
          aria-label="Próxima página"
        >
          <ChevronRight size={15} />
        </button>
      </div>
    </div>
  );
}

function ItemPortrait({ pull }: { pull: ConvenePull }) {
  return (
    <span className={`conveneItemPortrait rarity-${pull.rarity}`}>
      {pull.iconPath?.startsWith('/cache/') ? (
        <img src={pull.iconPath} alt="" />
      ) : normalizeResourceType(pull.resourceType) === 'character' ? (
        <UserRound size={17} />
      ) : (
        <Swords size={17} />
      )}
    </span>
  );
}

function ConveneEmpty({ onImport }: { onImport(): void }) {
  return (
    <section className="conveneEmpty">
      <div>
        <History size={30} />
      </div>
      <span className="eyebrow">ARQUIVO AINDA VAZIO</span>
      <h2>Preserve seu histórico de Convene</h2>
      <p>
        Abra o Histórico de Convocação no jogo. O WaveArchive encontra a URL no Client.log ou aceita
        uma URL colada manualmente.
      </p>
      <button className="convenePrimaryButton" onClick={onImport}>
        <Import size={16} />
        Importar primeiro histórico
      </button>
    </section>
  );
}

function ImportDialog({
  mode,
  rawURL,
  importing,
  onURL,
  onMode,
  onClose,
  onAutomatic,
  onLog,
  onSubmitURL,
}: {
  mode: ImportMode;
  rawURL: string;
  importing: boolean;
  onURL(value: string): void;
  onMode(value: ImportMode): void;
  onClose(): void;
  onAutomatic(): void;
  onLog(): void;
  onSubmitURL(): void;
}) {
  return (
    <div className="conveneModalBackdrop" role="presentation" onMouseDown={onClose}>
      <section
        className="conveneImportDialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="convene-import-title"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <header>
          <div>
            <span className="eyebrow">IMPORTAÇÃO LOCAL</span>
            <h2 id="convene-import-title">Adicionar histórico</h2>
          </div>
          <button onClick={onClose} aria-label="Fechar">
            <X size={17} />
          </button>
        </header>
        {mode === 'url' ? (
          <>
            <label className="conveneURLField">
              <span>URL do histórico oficial</span>
              <textarea
                autoFocus
                value={rawURL}
                onChange={(event) => onURL(event.target.value)}
                placeholder="https://aki-gm-resources-oversea.aki-game.net/aki/gacha/…"
                rows={5}
              />
            </label>
            <div className="conveneTokenNote">
              <LockKeyhole size={15} />A URL contém um token temporário de leitura. Ela não será
              salva no banco de dados.
            </div>
            <footer>
              <button className="conveneSecondaryButton" onClick={() => onMode('options')}>
                Voltar
              </button>
              <button
                className="convenePrimaryButton"
                disabled={importing || !rawURL.trim()}
                onClick={onSubmitURL}
              >
                {importing ? <RefreshCw className="spin" size={15} /> : <Import size={15} />}
                Importar
              </button>
            </footer>
          </>
        ) : (
          <>
            <p>
              Primeiro abra Convene → Histórico no jogo. Escolha a forma de leitura mais
              conveniente.
            </p>
            <div className="conveneImportOptions">
              <button disabled={importing} onClick={onAutomatic}>
                <FileSearch size={22} />
                <span>
                  <strong>Encontrar automaticamente</strong>
                  <small>Procura o Client.log nas instalações conhecidas sem modificar nada.</small>
                </span>
              </button>
              <button disabled={importing} onClick={onLog}>
                <CalendarDays size={22} />
                <span>
                  <strong>Selecionar Client.log</strong>
                  <small>Escolha manualmente o arquivo de log da instalação do jogo.</small>
                </span>
              </button>
              <button disabled={importing} onClick={() => onMode('url')}>
                <Link2 size={22} />
                <span>
                  <strong>Colar URL</strong>
                  <small>Use a URL copiada da página de histórico oficial.</small>
                </span>
              </button>
            </div>
            {importing && (
              <div className="conveneImporting" role="status">
                <RefreshCw className="spin" size={16} />
                Consultando as categorias de banner…
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
}

function fiveStarTimeline(pulls: ConvenePull[]) {
  const chronological = [...pulls].sort((a, b) => {
    if (a.obtainedAt === b.obtainedAt) return b.sourceIndex - a.sourceIndex;
    return a.obtainedAt.localeCompare(b.obtainedAt);
  });
  let pity = 0;
  let seenFive = false;
  const entries: { pull: ConvenePull; pity: number; pityExact: boolean }[] = [];
  for (const pull of chronological) {
    pity += 1;
    if (pull.rarity === 5) {
      entries.push({ pull, pity, pityExact: seenFive });
      pity = 0;
      seenFive = true;
    }
  }
  return entries.slice(-6);
}

function buildPityMap(pulls: ConvenePull[]) {
  const chronological = [...pulls].sort((a, b) => {
    if (a.obtainedAt === b.obtainedAt) return b.sourceIndex - a.sourceIndex;
    return a.obtainedAt.localeCompare(b.obtainedAt);
  });
  const result = new Map<number, number>();
  let pity = 0;
  for (const pull of chronological) {
    pity += 1;
    result.set(pull.id, pity);
    if (pull.rarity === 5) pity = 0;
  }
  return result;
}

function normalizeResourceType(value: string) {
  return /resonator|resonante|ressonante|ressonador|character|personagem/i.test(value)
    ? 'character'
    : 'weapon';
}

function resourceTypeLabel(value: string) {
  return normalizeResourceType(value) === 'character' ? 'Ressonador' : 'Arma';
}

function guaranteeLabel(value: ConvenePoolSummary['guaranteeState']) {
  if (value === 'guaranteed') return 'Próximo 5★ garantido';
  if (value === 'not_guaranteed') return '50/50';
  if (value === 'unknown') return 'Não comprovada';
  return 'Não se aplica';
}

function parseGameDate(value: string) {
  return new Date(value.replace(' ', 'T') + (value.includes('T') ? '' : ''));
}

function formatDate(value: string) {
  const date = parseGameDate(value);
  return Number.isNaN(date.getTime())
    ? value.slice(0, 10)
    : date.toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', year: 'numeric' });
}

function formatTime(value: string) {
  const date = parseGameDate(value);
  return Number.isNaN(date.getTime())
    ? value.slice(11, 16)
    : date.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
}

function formatDateTime(value: string) {
  const date = parseGameDate(value);
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleString('pt-BR', {
        day: '2-digit',
        month: 'short',
        hour: '2-digit',
        minute: '2-digit',
      });
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('pt-BR').format(value);
}

function rateLabel(value: number, total: number) {
  return total ? `${((value / total) * 100).toFixed(2)}% dos giros` : 'Sem dados';
}

function maskPlayerID(value: string) {
  if (value.length <= 4) return value;
  return `${value.slice(0, 3)}${'•'.repeat(Math.min(5, value.length - 4))}${value.slice(-1)}`;
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}

function normalizeOverview(value?: Partial<ConveneOverview> | null): ConveneOverview {
  return {
    profile: value?.profile,
    pools: Array.isArray(value?.pools) ? value.pools : [],
    pulls: Array.isArray(value?.pulls) ? value.pulls : [],
    total: Number(value?.total) || 0,
    count5: Number(value?.count5) || 0,
    count4: Number(value?.count4) || 0,
    count3: Number(value?.count3) || 0,
    lastImportedAt: value?.lastImportedAt ?? '',
  };
}

const emptyOverview: ConveneOverview = {
  pools: [],
  pulls: [],
  total: 0,
  count5: 0,
  count4: 0,
  count3: 0,
  lastImportedAt: '',
};
