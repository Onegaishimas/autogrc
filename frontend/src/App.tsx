import { useState, useCallback } from 'react';
import { ConnectionSettings } from './features/connection';
import { ControlsList } from './features/controls';
import { SystemsView } from './features/systems';
import { PullWizard } from './features/pull';
import { ModifiedStatementsList, SystemStatementsList } from './features/statements';
import { PushWorkflow } from './features/push';
import { AuditLog } from './features/audit';
import './features/connection/connection.css';
import './features/controls/controls.css';
import './features/systems/systems.css';
import './features/pull/pull.css';
import './features/statements/statements.css';
import './features/push/push.css';
import './features/audit/audit.css';

type View = 'connection' | 'controls' | 'systems' | 'system-detail' | 'pull' | 'pending' | 'push' | 'audit';

interface SelectedSystem {
  id: string;
  name: string;
}

function App() {
  const [currentView, setCurrentView] = useState<View>('systems');
  const [selectedSystem, setSelectedSystem] = useState<SelectedSystem | null>(null);

  const handlePullComplete = useCallback(() => {
    setCurrentView('systems');
  }, []);

  const handleSelectSystem = useCallback((system: SelectedSystem) => {
    setSelectedSystem(system);
    setCurrentView('system-detail');
  }, []);

  const handleBackToSystems = useCallback(() => {
    setSelectedSystem(null);
    setCurrentView('systems');
  }, []);

  return (
    <div className="app">
      <nav className="app-nav">
        <div className="nav-brand">ControlCRUD</div>
        <div className="nav-links">
          <button
            type="button"
            className={`nav-link ${currentView === 'systems' || currentView === 'system-detail' ? 'active' : ''}`}
            onClick={() => {
              setSelectedSystem(null);
              setCurrentView('systems');
            }}
          >
            Systems
          </button>
          <button
            type="button"
            className={`nav-link ${currentView === 'pull' ? 'active' : ''}`}
            onClick={() => setCurrentView('pull')}
          >
            Import
          </button>
          <button
            type="button"
            className={`nav-link ${currentView === 'controls' ? 'active' : ''}`}
            onClick={() => setCurrentView('controls')}
          >
            Controls
          </button>
          <button
            type="button"
            className={`nav-link ${currentView === 'pending' ? 'active' : ''}`}
            onClick={() => setCurrentView('pending')}
          >
            Pending Changes
          </button>
          <button
            type="button"
            className={`nav-link ${currentView === 'push' ? 'active' : ''}`}
            onClick={() => setCurrentView('push')}
          >
            Push
          </button>
          <button
            type="button"
            className={`nav-link ${currentView === 'audit' ? 'active' : ''}`}
            onClick={() => setCurrentView('audit')}
          >
            Audit Trail
          </button>
          <button
            type="button"
            className={`nav-link ${currentView === 'connection' ? 'active' : ''}`}
            onClick={() => setCurrentView('connection')}
          >
            Settings
          </button>
        </div>
      </nav>
      <main className="app-content">
        {currentView === 'systems' && <SystemsView onSelectSystem={handleSelectSystem} />}
        {currentView === 'system-detail' && selectedSystem && (
          <SystemStatementsList
            systemId={selectedSystem.id}
            systemName={selectedSystem.name}
            onBack={handleBackToSystems}
          />
        )}
        {currentView === 'pull' && <PullWizard onComplete={handlePullComplete} />}
        {currentView === 'controls' && <ControlsList />}
        {currentView === 'pending' && (
          <div className="pending-view">
            <div className="pending-header">
              <h1>Pending Changes</h1>
              <p>Review and manage your local modifications before pushing to ServiceNow.</p>
            </div>
            <ModifiedStatementsList />
          </div>
        )}
        {currentView === 'push' && <PushWorkflow onComplete={() => setCurrentView('pending')} />}
        {currentView === 'audit' && <AuditLog />}
        {currentView === 'connection' && <ConnectionSettings />}
      </main>
    </div>
  );
}

export default App;
