import { SystemCard } from './SystemCard';
import type { DiscoveredSystem, LocalSystem } from '../types';

interface SystemListProps {
  systems: (DiscoveredSystem | LocalSystem)[];
  selectedIds: Set<string>;
  onSelect: (sysId: string) => void;
  onDeselect: (sysId: string) => void;
  disabledIds?: Set<string>;
  showStats?: boolean;
  isLoading?: boolean;
  emptyMessage?: string;
}

function getSystemId(system: DiscoveredSystem | LocalSystem): string {
  return 'id' in system ? system.id : system.sn_sys_id;
}

export function SystemList({
  systems,
  selectedIds,
  onSelect,
  onDeselect,
  disabledIds = new Set(),
  showStats = false,
  isLoading = false,
  emptyMessage = 'No systems found',
}: SystemListProps) {
  if (isLoading) {
    return (
      <div className="system-list-loading">
        <div className="loading-skeleton">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    );
  }

  if (systems.length === 0) {
    return (
      <div className="system-list-empty">
        <p>{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="system-list">
      {systems.map((system) => {
        const sysId = getSystemId(system);
        return (
          <SystemCard
            key={sysId}
            system={system}
            isSelected={selectedIds.has(sysId)}
            onSelect={onSelect}
            onDeselect={onDeselect}
            disabled={disabledIds.has(sysId)}
            showStats={showStats}
          />
        );
      })}
    </div>
  );
}

export default SystemList;
