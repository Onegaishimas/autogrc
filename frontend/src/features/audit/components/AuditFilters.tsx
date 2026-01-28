import type { AuditFilters as AuditFiltersType, EventType, EntityType } from '../types';

const EVENT_TYPES: { value: EventType; label: string }[] = [
  { value: 'pull', label: 'Pull' },
  { value: 'push', label: 'Push' },
  { value: 'edit', label: 'Edit' },
  { value: 'conflict_detected', label: 'Conflict Detected' },
  { value: 'conflict_resolved', label: 'Conflict Resolved' },
  { value: 'connection_test', label: 'Connection Test' },
  { value: 'connection_config', label: 'Connection Config' },
  { value: 'system_import', label: 'System Import' },
  { value: 'system_delete', label: 'System Delete' },
];

const ENTITY_TYPES: { value: EntityType; label: string }[] = [
  { value: 'system', label: 'System' },
  { value: 'control', label: 'Control' },
  { value: 'statement', label: 'Statement' },
  { value: 'connection', label: 'Connection' },
];

interface AuditFiltersProps {
  filters: AuditFiltersType;
  onChange: (filters: AuditFiltersType) => void;
}

export function AuditFilters({ filters, onChange }: AuditFiltersProps) {
  const handleEventTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value as EventType | '';
    onChange({
      ...filters,
      event_types: value ? [value] : undefined,
      page: 1,
    });
  };

  const handleEntityTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value as EntityType | '';
    onChange({
      ...filters,
      entity_types: value ? [value] : undefined,
      page: 1,
    });
  };

  const handleStatusChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    onChange({
      ...filters,
      status: value || undefined,
      page: 1,
    });
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({
      ...filters,
      search: e.target.value || undefined,
      page: 1,
    });
  };

  const handleStartDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    onChange({
      ...filters,
      start_date: value ? new Date(value).toISOString() : undefined,
      page: 1,
    });
  };

  const handleEndDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    onChange({
      ...filters,
      end_date: value ? new Date(value).toISOString() : undefined,
      page: 1,
    });
  };

  const handleClearFilters = () => {
    onChange({ page: 1, page_size: filters.page_size || 50 });
  };

  const hasActiveFilters =
    filters.event_types?.length ||
    filters.entity_types?.length ||
    filters.status ||
    filters.search ||
    filters.start_date ||
    filters.end_date;

  return (
    <div className="audit-filters">
      <div className="audit-filters-row">
        <div className="audit-filter-group">
          <label htmlFor="event-type">Event Type</label>
          <select
            id="event-type"
            value={filters.event_types?.[0] || ''}
            onChange={handleEventTypeChange}
          >
            <option value="">All Events</option>
            {EVENT_TYPES.map((t) => (
              <option key={t.value} value={t.value}>
                {t.label}
              </option>
            ))}
          </select>
        </div>

        <div className="audit-filter-group">
          <label htmlFor="entity-type">Entity Type</label>
          <select
            id="entity-type"
            value={filters.entity_types?.[0] || ''}
            onChange={handleEntityTypeChange}
          >
            <option value="">All Entities</option>
            {ENTITY_TYPES.map((t) => (
              <option key={t.value} value={t.value}>
                {t.label}
              </option>
            ))}
          </select>
        </div>

        <div className="audit-filter-group">
          <label htmlFor="status">Status</label>
          <select
            id="status"
            value={filters.status || ''}
            onChange={handleStatusChange}
          >
            <option value="">All Statuses</option>
            <option value="success">Success</option>
            <option value="failure">Failure</option>
          </select>
        </div>

        <div className="audit-filter-group">
          <label htmlFor="search">Search</label>
          <input
            id="search"
            type="text"
            placeholder="Search events..."
            value={filters.search || ''}
            onChange={handleSearchChange}
          />
        </div>
      </div>

      <div className="audit-filters-row">
        <div className="audit-filter-group">
          <label htmlFor="start-date">Start Date</label>
          <input
            id="start-date"
            type="date"
            value={filters.start_date?.split('T')[0] || ''}
            onChange={handleStartDateChange}
          />
        </div>

        <div className="audit-filter-group">
          <label htmlFor="end-date">End Date</label>
          <input
            id="end-date"
            type="date"
            value={filters.end_date?.split('T')[0] || ''}
            onChange={handleEndDateChange}
          />
        </div>

        {hasActiveFilters && (
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={handleClearFilters}
          >
            Clear Filters
          </button>
        )}
      </div>
    </div>
  );
}

export default AuditFilters;
