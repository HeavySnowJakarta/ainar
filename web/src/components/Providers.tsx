import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

interface Provider {
  name: string;
  code: string;
  api_type: string;
  base_url: string;
  enabled: boolean;
  has_key: boolean;
}

interface KnownProvider {
  name: string;
  code: string;
  api_type: string;
  base_url: string;
}

function Providers() {
  const { t } = useTranslation();
  const [providers, setProviders] = useState<Provider[]>([]);
  const [knownProviders, setKnownProviders] = useState<KnownProvider[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    code: '',
    api_type: 'openai',
    base_url: '',
    api_key: '',
  });

  useEffect(() => {
    fetchProviders();
    fetchKnownProviders();
  }, []);

  const fetchProviders = async () => {
    try {
      const res = await fetch('/api/providers');
      const data = await res.json();
      setProviders(data);
    } catch (err) {
      console.error('Failed to fetch providers:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchKnownProviders = async () => {
    try {
      const res = await fetch('/api/known-providers');
      const data = await res.json();
      setKnownProviders(data);
    } catch (err) {
      console.error('Failed to fetch known providers:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingProvider) {
        await fetch(`/api/providers/${editingProvider.code}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData),
        });
      } else {
        await fetch('/api/providers', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData),
        });
      }
      setShowModal(false);
      resetForm();
      fetchProviders();
    } catch (err) {
      console.error('Failed to save provider:', err);
    }
  };

  const handleDelete = async (code: string) => {
    if (!confirm(t('providers.confirmDelete'))) return;
    try {
      await fetch(`/api/providers/${code}`, { method: 'DELETE' });
      fetchProviders();
    } catch (err) {
      console.error('Failed to delete provider:', err);
    }
  };

  const handleEdit = (provider: Provider) => {
    setEditingProvider(provider);
    setFormData({
      name: provider.name,
      code: provider.code,
      api_type: provider.api_type,
      base_url: provider.base_url,
      api_key: '',
    });
    setShowModal(true);
  };

  const resetForm = () => {
    setFormData({
      name: '',
      code: '',
      api_type: 'openai',
      base_url: '',
      api_key: '',
    });
    setEditingProvider(null);
    setSearchQuery('');
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    setSearchQuery(query);
    setShowSuggestions(query.length > 0);
  };

  const selectKnownProvider = (known: KnownProvider) => {
    setFormData({
      ...formData,
      name: known.name,
      code: known.code,
      api_type: known.api_type,
      base_url: known.base_url,
    });
    setSearchQuery('');
    setShowSuggestions(false);
  };

  const filteredKnown = knownProviders.filter(
    (p) =>
      p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      p.code.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return <div className="empty-state">{t('common.loading')}</div>;
  }

  return (
    <div className="section">
      <div className="section-header">
        <div>
          <h2 className="section-title">{t('providers.title')}</h2>
          <p className="section-description">{t('providers.description')}</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => {
            resetForm();
            setShowModal(true);
          }}
        >
          {t('providers.add')}
        </button>
      </div>

      {providers.length === 0 ? (
        <div className="empty-state">
          <p>{t('providers.noProviders')}</p>
        </div>
      ) : (
        providers.map((provider) => (
          <div key={provider.code} className="card">
            <div className="card-header">
              <div>
                <h3 className="card-title">{provider.name}</h3>
                <p className="card-subtitle">
                  <code>{provider.code}</code> · {provider.api_type} · {provider.base_url}
                </p>
              </div>
              <div className="actions">
                <span
                  className={`badge ${provider.enabled ? 'badge-success' : 'badge-warning'}`}
                >
                  {provider.enabled ? t('providers.enabled') : t('providers.disabled')}
                </span>
                <span className={`badge ${provider.has_key ? 'badge-success' : 'badge-warning'}`}>
                  {provider.has_key ? t('providers.hasKey') : t('providers.noKey')}
                </span>
                <button className="btn btn-secondary btn-sm" onClick={() => handleEdit(provider)}>
                  {t('providers.edit')}
                </button>
                <button
                  className="btn btn-danger btn-sm"
                  onClick={() => handleDelete(provider.code)}
                >
                  {t('providers.delete')}
                </button>
              </div>
            </div>
          </div>
        ))
      )}

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">
                {editingProvider ? t('providers.edit') : t('providers.add')}
              </h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>
                ×
              </button>
            </div>

            <form onSubmit={handleSubmit}>
              {!editingProvider && (
                <div className="form-group datalist-container">
                  <label className="form-label">{t('providers.searchKnown')}</label>
                  <input
                    type="text"
                    className="form-input"
                    value={searchQuery}
                    onChange={handleSearchChange}
                    placeholder={t('providers.searchKnown')}
                    onFocus={() => setShowSuggestions(searchQuery.length > 0)}
                    onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                  />
                  {showSuggestions && filteredKnown.length > 0 && (
                    <div className="datalist-dropdown">
                      {filteredKnown.slice(0, 10).map((known) => (
                        <div
                          key={known.code}
                          className="datalist-item"
                          onClick={() => selectKnownProvider(known)}
                        >
                          <strong>{known.name}</strong>
                          <br />
                          <small>{known.base_url}</small>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              <div className="form-group">
                <label className="form-label">{t('providers.name')}</label>
                <input
                  type="text"
                  className="form-input"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>

              <div className="form-group">
                <label className="form-label">{t('providers.code')}</label>
                <input
                  type="text"
                  className="form-input"
                  value={formData.code}
                  onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                  pattern="[a-z0-9._-]+"
                  required
                  disabled={!!editingProvider}
                />
                <p className="form-hint">{t('providers.codeHint')}</p>
              </div>

              <div className="form-group">
                <label className="form-label">{t('providers.apiType')}</label>
                <select
                  className="form-select"
                  value={formData.api_type}
                  onChange={(e) => setFormData({ ...formData, api_type: e.target.value })}
                >
                  <option value="openai">OpenAI</option>
                  <option value="anthropic">Anthropic</option>
                  <option value="google">Google</option>
                  <option value="ollama">Ollama</option>
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">{t('providers.baseUrl')}</label>
                <input
                  type="url"
                  className="form-input"
                  value={formData.base_url}
                  onChange={(e) => setFormData({ ...formData, base_url: e.target.value })}
                  required
                />
              </div>

              <div className="form-group">
                <label className="form-label">{t('providers.apiKey')}</label>
                <input
                  type="password"
                  className="form-input"
                  value={formData.api_key}
                  onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                  placeholder={editingProvider ? '(unchanged)' : ''}
                />
              </div>

              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  {t('common.cancel')}
                </button>
                <button type="submit" className="btn btn-primary">
                  {t('common.save')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default Providers;
