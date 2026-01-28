import type { DiscoveredSystem, LocalSystem } from '../types';

interface SystemCardProps {
  system: DiscoveredSystem | LocalSystem;
  isSelected?: boolean;
  onSelect?: (sysId: string) => void;
  onDeselect?: (sysId: string) => void;
  disabled?: boolean;
  showStats?: boolean;
}

function isLocalSystem(system: DiscoveredSystem | LocalSystem): system is LocalSystem {
  return 'id' in system && 'control_count' in system;
}

export function SystemCard({
  system,
  isSelected = false,
  onSelect,
  onDeselect,
  disabled = false,
  showStats = false,
}: SystemCardProps) {
  const sysId = isLocalSystem(system) ? system.id : system.sn_sys_id;
  const isImported = isLocalSystem(system) || system.is_imported;

  const handleClick = () => {
    if (disabled) return;

    if (isSelected && onDeselect) {
      onDeselect(sysId);
    } else if (!isSelected && onSelect) {
      onSelect(sysId);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  };

  return (
    <div
      className={`system-card ${isSelected ? 'selected' : ''} ${disabled ? 'disabled' : ''} ${isImported ? 'imported' : ''}`}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      tabIndex={disabled ? -1 : 0}
      role="button"
      aria-pressed={isSelected}
      aria-disabled={disabled}
    >
      <div className="system-card-header">
        <div className="system-card-checkbox">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={handleClick}
            disabled={disabled}
            tabIndex={-1}
            aria-hidden="true"
          />
        </div>
        <h3 className="system-card-name">{system.name}</h3>
        {isImported && (
          <span className="system-badge imported">Imported</span>
        )}
      </div>

      {system.description && (
        <p className="system-card-description">{system.description}</p>
      )}

      {system.owner && (
        <div className="system-card-owner">
          <span className="label">Owner:</span> {system.owner}
        </div>
      )}

      {showStats && isLocalSystem(system) && (
        <div className="system-card-stats">
          <div className="stat">
            <span className="stat-value">{system.control_count}</span>
            <span className="stat-label">Controls</span>
          </div>
          <div className="stat">
            <span className="stat-value">{system.statement_count}</span>
            <span className="stat-label">Statements</span>
          </div>
          {system.modified_count > 0 && (
            <div className="stat modified">
              <span className="stat-value">{system.modified_count}</span>
              <span className="stat-label">Modified</span>
            </div>
          )}
        </div>
      )}

      {showStats && isLocalSystem(system) && system.last_pull_at && (
        <div className="system-card-sync">
          Last synced: {new Date(system.last_pull_at).toLocaleDateString()}
        </div>
      )}
    </div>
  );
}

export default SystemCard;
