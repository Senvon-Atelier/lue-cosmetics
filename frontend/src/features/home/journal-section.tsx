import { Icon } from '../shared/ui/icons';

// Mock blog post data for now
const blogPosts = [
  {
    id: '1',
    tag: 'Skincare',
    readMin: 5,
    date: 'Mar 15',
    title: 'The Art of Double Cleansing: Why Your Skin Needs It',
    excerpt: 'Discover the Korean skincare secret that\'s transforming routines worldwide.',
    tone: 'lavender' as const,
  },
  {
    id: '2',
    tag: 'Ingredients',
    readMin: 7,
    date: 'Mar 12',
    title: 'Niacinamide: The Multi-Tasking Hero Your Routine Needs',
    excerpt: 'From pore minimization to barrier repair, this ingredient does it all.',
    tone: 'cream' as const,
  },
  {
    id: '3',
    tag: 'Routine',
    readMin: 4,
    date: 'Mar 10',
    title: 'Building Your First Skincare Routine: A Beginner\'s Guide',
    excerpt: 'Start with the essentials and build from there. Here\'s everything you need to know.',
    tone: 'rose' as const,
  },
];

export function JournalSection() {
  return (
    <section id="journal" className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">The Journal</div>
            <h2
              className="font-display"
              style={{ fontSize: 'clamp(32px, 4vw, 56px)', fontWeight: 400, lineHeight: 1.2 }}
            >
              Read the <em>ritual.</em>
            </h2>
          </div>
          <a className="section-link">
            All stories <Icon name="arrow" size={14} />
          </a>
        </div>
        <div className="grid-3 journal-grid">
          {blogPosts.map((post) => (
            <a key={post.id} className="journal-card">
              <div className="ph ph--lavender" style={{ aspectRatio: '4/3' }}>
                <span className="ph-label">{post.tag} · editorial</span>
              </div>
              <div className="journal-body">
                <div className="journal-meta">
                  <span className="chip">{post.tag}</span>
                  <span>
                    {post.readMin} min · {post.date}
                  </span>
                </div>
                <h3 className="journal-title">{post.title}</h3>
                <p className="journal-excerpt">{post.excerpt}</p>
                <div className="journal-more">
                  Read story <Icon name="arrow" size={12} />
                </div>
              </div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}
