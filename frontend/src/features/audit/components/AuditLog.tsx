import { useState, useCallback } from 'react';
import { useAuditEvents, useAuditStats, useExportAudit } from '../hooks/useAudit';
import { AuditFilters } from './AuditFilters';
import { AuditTable } from './AuditTable';
import { AuditDetail } from './AuditDetail';
import type { AuditFilters as AuditFiltersType } from '../types';

export function AuditLog() {
  const [filters, setFilters] = useState<AuditFiltersType>({
    page: 1,
    page_size: 50,
  });
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);

  const { data: eventsData, isLoading: eventsLoading } = useAuditEvents(filters);
  const { data: statsData } = useAuditStats();
  const exportMutation = useExportAudit();

  const handleFilterChange = useCallback((newFilters: AuditFiltersType) => {
    setFilters({ ...newFilters, page: 1 });
  }, []);

  const handlePageChange = useCallback((page: number) => {
    setFilters((prev) => ({ ...prev, page }));
  }, []);

  const handleSelectEvent = useCallback((id: string) => {
    setSelectedEventId(id);
  }, []);

  const handleCloseDetail = useCallback(() => {
    setSelectedEventId(null);
  }, []);

  const handleExport = useCallback(() => {
    exportMutation.mutate(filters);
  }, [filters, exportMutation]);

  const pageSize = filters.page_size || 50;
  const totalPages = eventsData?.total_pages || 1;
  const currentPage = eventsData?.page || 1;

  return (
    <div className="audit-log">
      <div className="audit-log-header">
        <div className="audit-log-title">
          <h1>Audit Trail</h1>
          <p>Track and review all synchronization activities and changes.</p>
        </div>
        <button
          type="button"
          className="btn btn-secondary"
          onClick={handleExport}
          disabled={exportMutation.isPending}
        >
          {exportMutation.isPending ? 'Exporting...' : 'Export CSV'}
        </button>
      </div>

      {statsData && (
        <div className="audit-stats">
          <div className="audit-stat-card">
            <span className="audit-stat-value">{statsData.total_events}</span>
            <span className="audit-stat-label">Total Events</span>
          </div>
          <div className="audit-stat-card">
            <span className="audit-stat-value">{statsData.events_today}</span>
            <span className="audit-stat-label">Today</span>
          </div>
          <div className="audit-stat-card">
            <span className="audit-stat-value">{statsData.events_this_week}</span>
            <span className="audit-stat-label">This Week</span>
          </div>
          <div className="audit-stat-card">
            <span className="audit-stat-value">{statsData.events_this_month}</span>
            <span className="audit-stat-label">This Month</span>
          </div>
        </div>
      )}

      <AuditFilters filters={filters} onChange={handleFilterChange} />

      <AuditTable
        events={eventsData?.events || []}
        isLoading={eventsLoading}
        onSelectEvent={handleSelectEvent}
      />

      {eventsData && eventsData.total_count > 0 && (
        <div className="audit-pagination">
          <div className="audit-pagination-info">
            Showing {(currentPage - 1) * pageSize + 1} -{' '}
            {Math.min(currentPage * pageSize, eventsData.total_count)} of{' '}
            {eventsData.total_count} events
          </div>
          <div className="audit-pagination-controls">
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage <= 1}
            >
              Previous
            </button>
            <span className="audit-pagination-page">
              Page {currentPage} of {totalPages}
            </span>
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
            >
              Next
            </button>
          </div>
        </div>
      )}

      {selectedEventId && (
        <AuditDetail eventId={selectedEventId} onClose={handleCloseDetail} />
      )}
    </div>
  );
}

export default AuditLog;
