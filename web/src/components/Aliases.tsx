import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

interface Alias {
  from: string;
  to: string;
}

function Aliases() {
  const { t } = useTranslation();
  const [aliases, setAliases] = useState<Alias[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [loading, setLoading] = useState(true);

  const [formData, setFormData] = useState({
    from: '',
    to: '',
  });

  useEffect(() => {
    fetchAliases();
  }, []);

  const fetchAliases = async () => {
    try {
      const res = await fetch('/api/aliases');
      const data = await res.json();
      setAliases(data || []);
    } catch (err) {
      console.error('Failed to fetch aliases:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await fetch('/api/aliases', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      setShowModal(false);
      resetForm();
      fetchAliases();
    } catch (err) {
      console.error('Failed to add alias:', err);
    }
  };

  const handleDelete = async (from: string) => {
    if (!confirm(t('aliases.confirmDelete'))) return;
    try {
      await fetch(`/api/aliases/${encodeURIComponent(from)}`, { method: 'DELETE' });
      fetchAliases();
    } catch (err) {
      console.error('Failed to delete alias:', err);
    }
  };

  const resetForm = () => {
    setFormData({
      from: '',
      to: '',
    });
  };

  if (loading) {
    return <div className="empty-state">{t('common.loading')}</div>;
  }

  return (
    <div className="section">
      <div className="section-header">
        <div>
          <h2 className="section-title">{t('aliases.title')}</h2>
          <p className="section-description">{t('aliases.description')}</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => {
            resetForm();
            setShowModal(true);
          }}
        >
          {t('aliases.add')}
        </button>
      </div>

      {aliases.length === 0 ? (
        <div className="empty-state">
          <p>{t('aliases.noAliases')}</p>
          <p><small>{t('aliases.example')}</small></p>
        </div>
      ) : (
        aliases.map((alias) => (
          <div key={alias.from} className="card">
            <div className="card-header">
              <div>
                <h3 className="card-title">
                  <code>{alias.from}</code>
                </h3>
                <p className="card-subtitle">
                  → <code>{alias.to}</code>
                </p>
              </div>
              <div className="actions">
                <button
                  className="btn btn-danger btn-sm"
                  onClick={() => handleDelete(alias.from)}
                >
                  {t('aliases.delete')}
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
              <h3 className="modal-title">{t('aliases.add')}</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>
                ×
              </button>
            </div>

            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label className="form-label">{t('aliases.from')}</label>
                <input
                  type="text"
                  className="form-input"
                  value={formData.from}
                  onChange={(e) => setFormData({ ...formData, from: e.target.value })}
                  placeholder="ora"
                  required
                />
              </div>

              <div className="form-group">
                <label className="form-label">{t('aliases.to')}</label>
                <input
                  type="text"
                  className="form-input"
                  value={formData.to}
                  onChange={(e) => setFormData({ ...formData, to: e.target.value })}
                  placeholder="openrouter/anthropic"
                  required
                />
              </div>

              <p className="form-hint">{t('aliases.example')}</p>

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

export default Aliases;
