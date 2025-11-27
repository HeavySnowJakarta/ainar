import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import './i18n';
import './App.css';

import Providers from './components/Providers';
import Keys from './components/Keys';
import Aliases from './components/Aliases';
import LanguageSelector from './components/LanguageSelector';

type Tab = 'providers' | 'keys' | 'aliases';

function App() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<Tab>('providers');

  return (
    <div className="app">
      <header className="header">
        <div className="header-content">
          <h1>{t('app.title')}</h1>
          <p className="subtitle">{t('app.subtitle')}</p>
        </div>
        <LanguageSelector />
      </header>

      <nav className="nav">
        <button
          className={`nav-btn ${activeTab === 'providers' ? 'active' : ''}`}
          onClick={() => setActiveTab('providers')}
        >
          {t('nav.providers')}
        </button>
        <button
          className={`nav-btn ${activeTab === 'keys' ? 'active' : ''}`}
          onClick={() => setActiveTab('keys')}
        >
          {t('nav.keys')}
        </button>
        <button
          className={`nav-btn ${activeTab === 'aliases' ? 'active' : ''}`}
          onClick={() => setActiveTab('aliases')}
        >
          {t('nav.aliases')}
        </button>
      </nav>

      <main className="main">
        {activeTab === 'providers' && <Providers />}
        {activeTab === 'keys' && <Keys />}
        {activeTab === 'aliases' && <Aliases />}
      </main>

      <footer className="footer">
        <p>{t('app.description')}</p>
      </footer>
    </div>
  );
}

export default App;
