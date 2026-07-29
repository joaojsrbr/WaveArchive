import type { ReactNode } from 'react';
import { ListFilter, RotateCcw, Search, X } from 'lucide-react';

export type LibrarySortOption = {
  value: string;
  label: string;
};

export function LibraryFilterBar({
  title,
  resultLabel,
  query,
  placeholder,
  sortValue,
  sortLabel,
  sortOptions,
  active,
  onQueryChange,
  onSortChange,
  onReset,
  className,
  contentClassName,
  children,
}: {
  title: string;
  resultLabel: string;
  query: string;
  placeholder: string;
  sortValue: string;
  sortLabel: string;
  sortOptions: LibrarySortOption[];
  active: boolean;
  onQueryChange: (value: string) => void;
  onSortChange: (value: string) => void;
  onReset: () => void;
  className?: string;
  contentClassName?: string;
  children: ReactNode;
}) {
  return (
    <div className={className ? `catalogFilterBar ${className}` : 'catalogFilterBar'}>
      <header className="catalogFilterHeader">
        <div className="catalogFilterTitle">
          <ListFilter size={16} />
          <div>
            <span className="sectionLabel">BIBLIOTECA</span>
            <strong>{title}</strong>
          </div>
          <small>{resultLabel}</small>
        </div>
        <label className="catalogFilterSearch">
          <Search size={17} />
          <span className="srOnly">{placeholder}</span>
          <input
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder={placeholder}
          />
          {query && (
            <button type="button" onClick={() => onQueryChange('')} aria-label="Limpar busca">
              <X size={13} />
            </button>
          )}
        </label>
        <select
          className="catalogSort"
          aria-label={sortLabel}
          value={sortValue}
          onChange={(event) => onSortChange(event.target.value)}
        >
          {sortOptions.map((option) => (
            <option value={option.value} key={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        {active && (
          <button type="button" className="catalogFilterReset" onClick={onReset}>
            <RotateCcw size={14} />
            Limpar filtros
          </button>
        )}
      </header>
      <div
        className={contentClassName ? `catalogFilterGrid ${contentClassName}` : 'catalogFilterGrid'}
      >
        {children}
      </div>
    </div>
  );
}
