import { useState, useEffect } from 'react';
import { Icon } from '../ui/icons';

const STORAGE_KEY = 'rue_case_study_dismissed';

export function CaseStudyBanner() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const dismissed = localStorage.getItem(STORAGE_KEY);
    if (!dismissed) setVisible(true);
  }, []);

  const dismiss = () => {
    localStorage.setItem(STORAGE_KEY, 'true');
    setVisible(false);
  };

  if (!visible) return null;

  return (
    <div className="cs-overlay" onClick={dismiss}>
      <div className="cs-modal" onClick={(e) => e.stopPropagation()}>
        <button className="cs-close" onClick={dismiss} aria-label="Close">
          <Icon name="close" size={18} />
        </button>

        <div className="cs-badge">Demo Case Study</div>

        <h2 className="cs-title">
          This is a <em>working prototype.</em>
        </h2>

        <p className="cs-body">
          Lue Cosmetics mimics a live e-commerce experience — browse products,
          add to cart, and go through checkout. It's built with a real backend,
          database, and payment integration (test mode).
        </p>

        <div className="cs-credentials">
          <div className="cs-creds-label">Quick login</div>
          <div className="cs-creds-row">
            <span className="cs-creds-key">Email</span>
            <span className="cs-creds-value">customer@luecosmetics.com</span>
          </div>
          <div className="cs-creds-row">
            <span className="cs-creds-key">Password</span>
            <span className="cs-creds-value">hunter22</span>
          </div>
        </div>

        <button className="btn btn-primary cs-cta" onClick={dismiss}>
          Continue to site <Icon name="arrow" size={14} />
        </button>
      </div>
    </div>
  );
}
