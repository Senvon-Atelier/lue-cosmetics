import { Link } from '@tanstack/react-router';
import { LEGAL_PAGES, getLegalPage } from '../../content/legal';

function Sidebar({ activeSlug }: { activeSlug: string }) {
  return (
    <nav className="legal-side" aria-label="Legal pages">
      <h4>Legal</h4>
      {LEGAL_PAGES.map((p) => {
        const isActive = p.slug === activeSlug;
        return (
          <Link
            key={p.slug}
            to="/legal/$slug"
            params={{ slug: p.slug }}
            className={isActive ? 'active' : undefined}
            aria-current={isActive ? 'page' : undefined}
          >
            {p.navLabel}
          </Link>
        );
      })}
    </nav>
  );
}

export function LegalPageView({ slug }: { slug: string }) {
  const page = getLegalPage(slug);

  if (!page) {
    return (
      <>
        <div className="legal-hero">
          <div className="legal-hero-inner">
            <h1>Not found</h1>
          </div>
        </div>
        <div className="legal-wrap">
          <Sidebar activeSlug="" />
          <div className="legal-body">
            <p>This policy could not be found.</p>
            <p>
              <Link to="/legal/$slug" params={{ slug: 'privacy' }}>
                Back to Privacy policy
              </Link>
            </p>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <div className="legal-hero">
        <div className="legal-hero-inner">
          <h1>{page.title}</h1>
          <div className="meta">Last updated {page.lastUpdated}</div>
        </div>
      </div>
      <div className="legal-wrap">
        <Sidebar activeSlug={page.slug} />
        <div className="legal-body">
          <p className="lead">{page.lead}</p>
          {page.body}
        </div>
      </div>
    </>
  );
}
