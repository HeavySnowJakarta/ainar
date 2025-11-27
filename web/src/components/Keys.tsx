import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

interface TokenLimit {
  type: string;
  max_tokens: number;
}

interface TokenUsage {
  tokens_used: number;
  last_reset: string;
}

interface DownstreamKey {
  name: string;
  created_at: string;
  enabled: boolean;
  limit?: TokenLimit;
  usage?: TokenUsage;
}

function Keys() {
  const { t } = useTranslation();
  const [keys, setKeys] = useState<DownstreamKey[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [showKeyModal, setShowKeyModal] = useState(false);
  const [generatedKey, setGeneratedKey] = useState('');
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    limitType: '',
    limitTokens: 0,
  });

  useEffect(() => {
    fetchKeys();
  }, []);

  const fetchKeys = async () => {
    try {
      const res = await fetch('/api/keys');
      const data = await res.json();
      setKeys(data);
    } catch (err) {
      console.error('Failed to fetch keys:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const body: { name: string; limit?: TokenLimit } = { name: formData.name };
      if (formData.limitType && formData.limitTokens > 0) {
        body.limit = {
          type: formData.limitType,
          max_tokens: formData.limitTokens,
        };
      }

      const res = await fetch('/api/keys', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      const data = await res.json();
      setGeneratedKey(data.key);
      setShowModal(false);
      setShowKeyModal(true);
      fetchKeys();
      resetForm();
    } catch (err) {
      console.error('Failed to generate key:', err);
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(t('keys.confirmDelete'))) return;
    try {
      await fetch(`/api/keys/${encodeURIComponent(name)}`, { method: 'DELETE' });
      fetchKeys();
    } catch (err) {
      console.error('Failed to delete key:', err);
    }
  };

  const toggleEnabled = async (key: DownstreamKey) => {
    try {
      await fetch(`/api/keys/${encodeURIComponent(key.name)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled: !key.enabled }),
      });
      fetchKeys();
    } catch (err) {
      console.error('Failed to toggle key:', err);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      limitType: '',
      limitTokens: 0,
    });
  };

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(generatedKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const getUsagePercentage = (key: DownstreamKey) => {
    if (!key.limit || !key.usage) return 0;
    return Math.min(100, (key.usage.tokens_used / key.limit.max_tokens) * 100);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString();
  };

  if (loading) {
    return <div className="empty-state">{t('common.loading')}</div>;
  }

  return (
    <div className="section">
      <div className="section-header">
        <div>
          <h2 className="section-title">{t('keys.title')}</h2>
          <p className="section-description">{t('keys.description')}</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => {
            resetForm();
            setShowModal(true);
          }}
        >
          {t('keys.generate')}
        </button>
      </div>

      {keys.length === 0 ? (
        <div className="empty-state">
          <p>{t('keys.noKeys')}</p>
        </div>
      ) : (
        keys.map((key) => (
          <div key={key.name} className="card">
            <div className="card-header">
              <div style={{ flex: 1 }}>
                <h3 className="card-title">{key.name}</h3>
                <p className="card-subtitle">
                  {t('keys.created')}: {formatDate(key.created_at)}
                  {key.limit && ` · ${t(`keys.limitTypes.${key.limit.type}`)}: ${key.limit.max_tokens.toLocaleString()} tokens`}
                </p>
                {key.limit && key.usage && (
                  <div style={{ marginTop: '0.5rem' }}>
                    <small>
                      {t('keys.usage')}: {key.usage.tokens_used.toLocaleString()} / {key.limit.max_tokens.toLocaleString()} ({getUsagePercentage(key).toFixed(1)}%)
                    </small>
                    <div className="progress-bar">
                      <div
                        className="progress-fill"
                        style={{ width: `${getUsagePercentage(key)}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
              <div className="actions">
                <span className={`badge ${key.enabled ? 'badge-success' : 'badge-warning'}`}>
                  {key.enabled ? t('providers.enabled') : t('providers.disabled')}
                </span>
                <button className="btn btn-secondary btn-sm" onClick={() => toggleEnabled(key)}>
                  {key.enabled ? t('providers.disabled') : t('providers.enabled')}
                </button>
                <button
                  className="btn btn-danger btn-sm"
                  onClick={() => handleDelete(key.name)}
                >
                  {t('keys.delete')}
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
              <h3 className="modal-title">{t('keys.generate')}</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>
                ×
              </button>
            </div>

            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label className="form-label">{t('keys.name')}</label>
                <input
                  type="text"
                  className="form-input"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>

              <div className="form-group">
                <label className="form-label">{t('keys.limitType')}</label>
                <select
                  className="form-select"
                  value={formData.limitType}
                  onChange={(e) => setFormData({ ...formData, limitType: e.target.value })}
                >
                  <option value="">{t('keys.limitTypes.none')}</option>
                  <option value="daily">{t('keys.limitTypes.daily')}</option>
                  <option value="weekly">{t('keys.limitTypes.weekly')}</option>
                  <option value="monthly">{t('keys.limitTypes.monthly')}</option>
                  <option value="total">{t('keys.limitTypes.total')}</option>
                </select>
              </div>

              {formData.limitType && (
                <div className="form-group">
                  <label className="form-label">{t('keys.limitTokens')}</label>
                  <input
                    type="number"
                    className="form-input"
                    value={formData.limitTokens}
                    onChange={(e) => setFormData({ ...formData, limitTokens: parseInt(e.target.value) || 0 })}
                    min="0"
                    required
                  />
                </div>
              )}

              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  {t('common.cancel')}
                </button>
                <button type="submit" className="btn btn-primary">
                  {t('keys.generate')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showKeyModal && (
        <div className="modal-overlay" onClick={() => setShowKeyModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">{t('keys.keyGenerated')}</h3>
              <button className="modal-close" onClick={() => setShowKeyModal(false)}>
                ×
              </button>
            </div>

            <p style={{ color: 'var(--warning)' }}>⚠️ {t('keys.keyWarning')}</p>

            <div className="key-display">{generatedKey}</div>

            <div className="modal-footer">
              <button className="btn btn-primary" onClick={copyToClipboard}>
                {copied ? t('keys.copied') : t('keys.copy')}
              </button>
              <button className="btn btn-secondary" onClick={() => setShowKeyModal(false)}>
                {t('common.close')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Keys;
