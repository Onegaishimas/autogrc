import { useState, useCallback } from 'react';

interface ControlsSearchProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function ControlsSearch({ value, onChange, placeholder = 'Search controls...' }: ControlsSearchProps) {
  const [localValue, setLocalValue] = useState(value);

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      onChange(localValue);
    },
    [localValue, onChange]
  );

  const handleClear = useCallback(() => {
    setLocalValue('');
    onChange('');
  }, [onChange]);

  return (
    <form className="controls-search" onSubmit={handleSubmit}>
      <div className="search-input-wrapper">
        <svg
          className="search-icon"
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <input
          type="text"
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          placeholder={placeholder}
          className="search-input"
        />
        {localValue && (
          <button type="button" onClick={handleClear} className="search-clear" aria-label="Clear search">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        )}
      </div>
      <button type="submit" className="btn btn-primary">
        Search
      </button>
    </form>
  );
}
