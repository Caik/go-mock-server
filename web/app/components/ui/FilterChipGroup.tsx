// FilterChipGroup - Reusable multi-select chip filter component
import React from 'react';

interface FilterChipGroupProps<T extends string> {
  /** Label displayed before the chips */
  label: string;
  /** Array of all possible options */
  options: readonly T[];
  /** Currently selected options (empty array = none selected) */
  selected: T[];
  /** Callback when selection changes */
  onChange: (selected: T[]) => void;
  /** Optional function to get display label for an option */
  getLabel?: (option: T) => string;
}

export function FilterChipGroup<T extends string>({
  label,
  options,
  selected,
  onChange,
  getLabel = (option) => option,
}: FilterChipGroupProps<T>) {
  const handleToggle = (option: T) => {
    if (selected.includes(option)) {
      onChange(selected.filter((s) => s !== option));
    } else {
      onChange([...selected, option]);
    }
  };

  return (
    <div className="filter-chip-group">
      <span className="filter-label">{label}:</span>
      {options.map((option) => (
        <button
          key={option}
          type="button"
          className={`filter-chip ${selected.includes(option) ? 'active' : ''}`}
          onClick={() => handleToggle(option)}
        >
          {getLabel(option)}
        </button>
      ))}
    </div>
  );
}

