import { useState, useCallback } from 'react';
import { usePolicyStatements } from '../hooks/useControls';
import { ControlCard } from './ControlCard';
import { ControlsSearch } from './ControlsSearch';
import { ControlsPagination } from './ControlsPagination';
import type { PolicyStatementsParams } from '../types';

const DEFAULT_PAGE_SIZE = 10;

export function ControlsList() {
  const [params, setParams] = useState<PolicyStatementsParams>({
    page: 1,
    page_size: DEFAULT_PAGE_SIZE,
  });

  const { data, isLoading, isError, error, isFetching } = usePolicyStatements(params);

  const handleSearch = useCallback((search: string) => {
    setParams((prev) => ({
      ...prev,
      search: search || undefined,
      page: 1, // Reset to first page on search
    }));
  }, []);

  const handlePageChange = useCallback((page: number) => {
    setParams((prev) => ({ ...prev, page }));
  }, []);

  return (
    <div className="controls-list">
      <header className="controls-header">
        <div className="header-content">
          <h1>Policy Statements</h1>
          {isFetching && !isLoading && <span className="sync-indicator">Syncing...</span>}
        </div>
        <p className="controls-description">
          View and manage compliance policy statements from ServiceNow GRC.
        </p>
      </header>

      <div className="controls-toolbar">
        <ControlsSearch value={params.search || ''} onChange={handleSearch} />
      </div>

      <main className="controls-content">
        {isLoading ? (
          <div className="controls-loading">
            <div className="loading-spinner" />
            <p>Loading policy statements...</p>
          </div>
        ) : isError ? (
          <div className="controls-error">
            <h3>Failed to load policy statements</h3>
            <p>{error instanceof Error ? error.message : 'An unexpected error occurred'}</p>
            <button
              className="btn btn-primary"
              onClick={() => setParams({ ...params })}
            >
              Retry
            </button>
          </div>
        ) : !data?.items.length ? (
          <div className="controls-empty">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="48"
              height="48"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="9" x2="15" y1="15" y2="15" />
            </svg>
            <h3>No policy statements found</h3>
            <p>
              {params.search
                ? 'Try adjusting your search terms'
                : 'Configure your ServiceNow connection to fetch policy statements'}
            </p>
          </div>
        ) : (
          <>
            <div className="controls-grid">
              {data.items.map((control) => (
                <ControlCard key={control.id} control={control} />
              ))}
            </div>
            <ControlsPagination
              pagination={data.pagination}
              onPageChange={handlePageChange}
            />
          </>
        )}
      </main>
    </div>
  );
}
