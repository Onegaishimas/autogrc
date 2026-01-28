import { useState } from 'react';
import { ConnectionSettings } from './features/connection';
import { ControlsList } from './features/controls';
import './features/connection/connection.css';
import './features/controls/controls.css';

type View = 'connection' | 'controls';

function App() {
  const [currentView, setCurrentView] = useState<View>('controls');

  return (
    <div className="app">
      <nav className="app-nav">
        <div className="nav-brand">ControlCRUD</div>
        <div className="nav-links">
          <button
            type="button"
            className={`nav-link ${currentView === 'controls' ? 'active' : ''}`}
            onClick={() => setCurrentView('controls')}
          >
            Controls
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
        {currentView === 'controls' && <ControlsList />}
        {currentView === 'connection' && <ConnectionSettings />}
      </main>
    </div>
  );
}

export default App;
