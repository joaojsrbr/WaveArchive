import type { ReactNode } from 'react';
import { RotateCcw, SlidersHorizontal, X } from 'lucide-react';

export function AdvancedFilters({
  open,
  activeCount,
  onToggle,
  onReset,
  children,
}: {
  open: boolean;
  activeCount: number;
  onToggle: () => void;
  onReset: () => void;
  children: ReactNode;
}) {
  return (
    <section className={open ? 'advancedFilters open' : 'advancedFilters'}>
      <button
        type="button"
        className={activeCount ? 'advancedFiltersTrigger active' : 'advancedFiltersTrigger'}
        aria-expanded={open}
        onClick={onToggle}
      >
        <SlidersHorizontal size={15} />
        Filtros avançados
        {activeCount > 0 && <span>{activeCount}</span>}
      </button>
      {open && (
        <div className="advancedFiltersPanel">
          <header>
            <div>
              <span className="sectionLabel">REFINAR RESULTADOS</span>
              <strong>Filtros avançados</strong>
            </div>
            <div>
              {activeCount > 0 && (
                <button type="button" onClick={onReset}>
                  <RotateCcw size={14} />
                  Limpar
                </button>
              )}
              <button type="button" aria-label="Fechar filtros" onClick={onToggle}>
                <X size={15} />
              </button>
            </div>
          </header>
          <div className="advancedFiltersGrid">{children}</div>
        </div>
      )}
    </section>
  );
}

export function FilterField({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: ReactNode;
}) {
  return (
    <label className="advancedFilterField">
      <span>{label}</span>
      {children}
      {hint && <small>{hint}</small>}
    </label>
  );
}

export function FilterRange({
  label,
  min,
  max,
  minValue,
  maxValue,
  onMinChange,
  onMaxChange,
  prefix,
}: {
  label: string;
  min: number;
  max: number;
  minValue?: number;
  maxValue?: number;
  onMinChange: (value: number) => void;
  onMaxChange: (value: number) => void;
  prefix?: string;
}) {
  return (
    <fieldset className="advancedFilterRange">
      <legend>{label}</legend>
      <label>
        <span>De</span>
        <div>
          {prefix && <i>{prefix}</i>}
          <input
            type="number"
            min={min}
            max={max}
            value={minValue || ''}
            placeholder={String(min)}
            onChange={(event) => onMinChange(Number(event.target.value))}
          />
        </div>
      </label>
      <label>
        <span>Até</span>
        <div>
          {prefix && <i>{prefix}</i>}
          <input
            type="number"
            min={min}
            max={max}
            value={maxValue || ''}
            placeholder={String(max)}
            onChange={(event) => onMaxChange(Number(event.target.value))}
          />
        </div>
      </label>
    </fieldset>
  );
}
